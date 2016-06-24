package sync

import (
	"time"

	// "github.com/coreos/etcd/client"

	"straas.io/external"
	"straas.io/pierce"
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

type sinker func(roomMeta pierce.RoomMeta, data interface{}, version uint64) error

type syncQueue chan string

type syncer struct {
	// etcd api
	api external.Etcd
	// pierce core manager
	coreMgr pierce.Core
	// named queues for better performance
	queues map[string]syncQueue
	done   chan bool
	sinker sinker
}

func (s *syncer) Start() {
	for _, q := range s.queues {
		go s.syncLoop(q)
	}
	go s.watcher()
}

func (s *syncer) Stop() {
	close(s.done)
}

func (s *syncer) Add(roomMeta pierce.RoomMeta) {
	// conver to roomid

}

func (s *syncer) watch() {
	// index out of sync

}

func (s *syncer) syncLoop(queue syncQueue) {
	for s.syncOnce(queue) {
	}
}

func (s *syncer) syncOnce(queue syncQueue) bool {
	select {
	case <-s.done:
		return false

	case resp := <-queue:
		// TODO: compare
		s.doSync(resp, false)
	}
	return true
}

// leverge namedQueue for better performance
// TODO: retry sync process or backoff unit success
func (s *syncer) doSync(etcdKey string, compare bool) error {
	// toKey ?!
	var roomMeta pierce.RoomMeta

	data, err := s.coreMgr.GetAll(roomMeta)
	if err != nil {
		// WTD ?!
		return err
	}

	// how to know it's newer or older changes ?!
	// sinker
	if err := s.sinker(roomMeta, data, 0); err != nil {
		return err
	}
	return nil
}
