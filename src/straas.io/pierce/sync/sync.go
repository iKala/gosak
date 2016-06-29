package sync

import (
	"encoding/json"
	"time"

	"github.com/cenk/backoff"
	"github.com/jinzhu/gorm"

	"straas.io/base/logmetric"
	"straas.io/pierce"
)

const (
	minRetryInterval = 10 * time.Millisecond
	maxRetryInterval = 3 * time.Second
	watchBuffer      = 10000
)

// Record defines how to keep changes in sinkers
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
	// or cluster rebuild, add unique constrain (Cluster, Version) to
	// void etcd verion number confliction
	Cluster uint32 `sql:"not null"`
	// Version is etcd version
	Version uint64 `sql:"not null"`
	// CreateAt for gorm to insert create time
	CreatedAt time.Time
}

// New creates an instance of Syncer
func New(coreMgr pierce.Core, db *gorm.DB,
	logm logmetric.LogMetric) (pierce.Syncer, error) {

	sinker, err := newSinker(db)
	if err != nil {
		return nil, err
	}

	return &syncerImpl{
		coreMgr: coreMgr,
		queue:   make(chan pierce.RoomMeta, watchBuffer),
		chDone:  make(chan bool),
		sinker:  sinker,
		logm:    logm,
	}, nil
}

// sinker is an interface for sink operations
type sinker interface {
	// Sink writes diffs
	Sink(roomMeta pierce.RoomMeta, data interface{}, version uint64) error
	// Diff gets diffs
	Diff(namespace string, index uint64, size int) ([]Record, error)
	// LastVersion get the latest verion of diff
	LatestVersion() (uint64, error)
}

type syncerImpl struct {
	// pierce core manager
	coreMgr pierce.Core
	queue   chan pierce.RoomMeta
	chDone  chan bool
	sinker  sinker
	backoff backoff.BackOff
	logm    logmetric.LogMetric
}

func (s *syncerImpl) Start() {
	go s.syncLoop()
	go s.watch()
}

func (s *syncerImpl) Stop() {
	close(s.chDone)
}

func (s *syncerImpl) Add(roomMeta pierce.RoomMeta) {
	s.queue <- roomMeta
}

func (s *syncerImpl) Diff(namespace string, afterIdx uint64, size int) ([]pierce.Record, error) {
	recs, err := s.sinker.Diff(namespace, afterIdx, size)
	if err != nil {
		return nil, err
	}
	result := make([]pierce.Record, 0, len(recs))
	for _, rec := range recs {
		var v interface{}
		if err := json.Unmarshal([]byte(rec.Value), &v); err != nil {
			return nil, err
		}
		result = append(result, pierce.Record{
			Index: rec.ID,
			Room: pierce.RoomMeta{
				Namespace: rec.Namespace,
				ID:        rec.Room,
			},
			Value: v,
		})
	}
	return result, nil
}

func (s *syncerImpl) watch() {
	for {
		// check done
		select {
		case <-s.chDone:
			return
		default:
		}

		var idx uint64
		backoff.Retry(func() error {
			var err error
			idx, err = s.sinker.LatestVersion()
			if err != nil {
				s.logm.BumpSum("syncer.lastversion.err", 1)
				s.logm.Errorf("fail to get latest version, err:%v", err)
				return err
			}
			return nil

		}, genBackoff())

		backoff.Retry(func() error {
			// TODO: IMPORTANT handle critical error here
			// TODO: watch must suggest next idx
			if err := s.coreMgr.Watch(idx, s.queue); err != nil {
				s.logm.BumpSum("syncer.corewatch.err", 1)
				s.logm.Errorf("fail to watch core, err:%v", err)
				return err
			}
			return nil

		}, genBackoff())
	}
}

func (s *syncerImpl) syncLoop() {
	for s.syncOnce() {
	}
}

func (s *syncerImpl) syncOnce() bool {
	select {
	case <-s.chDone:
		return false

	case resp := <-s.queue:
		backoff.Retry(func() error {
			if err := s.doSync(resp, false); err != nil {
				s.logm.BumpSum("syncer.dosync.err", 1)
				s.logm.Errorf("syncer do sync fail, %v", err)
				return err
			}
			return nil
		}, genBackoff())
	}
	return true
}

func (s *syncerImpl) doSync(roomMeta pierce.RoomMeta, compare bool) error {
	data, version, err := s.coreMgr.GetAll(roomMeta)
	if err != nil {
		return err
	}
	if err := s.sinker.Sink(roomMeta, data, version); err != nil {
		return err
	}
	return nil
}

func genBackoff() backoff.BackOff {
	bf := backoff.NewExponentialBackOff()
	bf.MaxInterval = maxRetryInterval
	bf.InitialInterval = minRetryInterval
	// always retry
	bf.MaxElapsedTime = 0
	return bf
}
