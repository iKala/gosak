package sync

import (
	"time"

	// "github.com/coreos/etcd/client"
	"github.com/cenk/backoff"

	"straas.io/pierce"
)

const (
	minRetryInterval = 10 * time.Millisecond
	maxRetryInterval = 3 * time.Second
	watchBuffer      = 10000
)

type Record struct {
	// ID is auto increment primary key
	ID uint64 `gorm:"primary_key,AUTO_INCREMENT"`
	// Namespace is room namespace
	Namespace string `gorm:"size:60" sql:"not null"`
	// Room is room id
	Room string `gorm:"size:100" sql:"not null"`
	// Value is json value of the room
	Value string `gorm:"size:0"`
	// Cluster is etcd cluster number, we might have multiple clusters
	// or cluster rebuild, add uniqu constrain (Cluster, Version) to
	// void etcd verion number confliction
	Cluster uint32 `sql:"not null"`
	// Version is etcd version
	Version uint64 `sql:"not null"`
	// CreateAt for gorm to insert create time
	CreatedAt time.Time
}

func New(coreMgr pierce.Core, namespace string) pierce.Syncer {
	bf := backoff.NewExponentialBackOff()
	bf.MaxInterval = maxRetryInterval
	bf.InitialInterval = minRetryInterval
	// always retry
	bf.MaxElapsedTime = 0

	return &syncer{
		coreMgr:   coreMgr,
		queue:     make(chan pierce.RoomMeta),
		backoff:   bf,
		namespace: namespace,
		// sinker: sinker,
	}
}

type sinker interface {
	Sink(roomMeta pierce.RoomMeta, data interface{}, version uint64) error
	Diff(index uint64, size int) ([]Record, error)
	LatestVersion() (uint64, error)
}

type syncer struct {
	// pierce core manager
	namespace string
	coreMgr   pierce.Core
	queue     chan pierce.RoomMeta
	done      chan bool
	sinker    sinker
	backoff   backoff.BackOff
}

func (s *syncer) Start() {
	go s.syncLoop()
	go s.watch()
}

func (s *syncer) Stop() {
	close(s.done)
}

func (s *syncer) Add(roomMeta pierce.RoomMeta) {
	// add room for sync
	s.queue <- roomMeta
}

func (s *syncer) watch() {
	ch := make(chan pierce.RoomMeta, watchBuffer)
	// always retry
	backoff.Retry(func() error {
		idx, err := s.sinker.LatestVersion()
		if err != nil {
			return err
		}
		// TODO: IMPORTANT handle critical error here
		return s.coreMgr.Watch(s.namespace, idx, ch)
	}, s.backoff)
}

func (s *syncer) syncLoop() {
	for s.syncOnce() {
	}
}

func (s *syncer) syncOnce() bool {
	select {
	case <-s.done:
		return false

	case resp := <-s.queue:
		// TODO: comparison mechanism
		// TODO: send metric if fail too many times
		backoff.Retry(func() error {
			// TODO: add metrics
			return s.doSync(resp, false)
		}, s.backoff)
	}
	return true
}

func (s *syncer) doSync(roomMeta pierce.RoomMeta, compare bool) error {
	data, version, err := s.coreMgr.GetAll(roomMeta)
	if err != nil {
		// backoff and retry
		return err
	}
	// how to know it's newer or older changes ?!
	if err := s.sinker.Sink(roomMeta, data, version); err != nil {
		return err
	}
	return nil
}
