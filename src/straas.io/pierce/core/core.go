package core

import (
	"crypto/md5"
	"fmt"
	"time"

	"straas.io/base/logmetric"
	"straas.io/external"
	"straas.io/pierce"
)

const (
	maintainInterval = 30 * time.Second
	// room TTL is 3 days
	roomTTL = 3 * 24 * time.Hour
)

// NewCore creates an instance of core manager
func NewCore(etcdAPI external.Etcd, keyPrefix string, logm logmetric.LogMetric) pierce.Core {
	rFactory := func(roomMeta pierce.RoomMeta, etcdKey string) Room {
		return newRoom(roomMeta, etcdKey, etcdAPI, logm)
	}
	return &coreImpl{
		rooms:     map[pierce.RoomMeta]Room{},
		etcdAPI:   etcdAPI,
		rFactory:  rFactory,
		keyPrefix: keyPrefix,
		chJoin:    make(chan pierce.SocketConnection, 1000),
		chLeave:   make(chan pierce.SocketConnection, 1000),
		chDone:    make(chan bool),
		logm:      logm,
	}
}

type roomFactory func(roomMeta pierce.RoomMeta, etcdKey string) Room

type coreImpl struct {
	rooms     map[pierce.RoomMeta]Room
	etcdAPI   external.Etcd
	rFactory  roomFactory
	keyPrefix string

	// channels
	chJoin  chan pierce.SocketConnection
	chLeave chan pierce.SocketConnection
	chDone  chan bool
	logm    logmetric.LogMetric
}

func (r *coreImpl) Start() {
	// TODO: check status ?!
	// start main loop
	go r.mainLoop()
}

func (r *coreImpl) Stop() {
	// TODO: check status ?!
	close(r.chDone)
}

func (r *coreImpl) Get(roomMeta pierce.RoomMeta, key string) (interface{}, uint64, error) {
	etcdKey := r.toEtcdKey(roomMeta, key)
	resp, err := r.etcdAPI.Get(etcdKey, true)
	if err != nil {
		return nil, 0, err
	}
	return toValue(resp.Node, unmarshaller)
}

func (r *coreImpl) GetAll(roomMeta pierce.RoomMeta) (interface{}, uint64, error) {
	return r.Get(roomMeta, "")
}

func (r *coreImpl) Set(roomMeta pierce.RoomMeta, key string, v interface{}, ttl time.Duration) error {
	value, err := marshaller(v)
	if err != nil {
		return err
	}

	roomKey := r.toEtcdKey(roomMeta, "")
	etcdKey := r.toEtcdKey(roomMeta, key)

	// fresh room dir ttl first in case server crash after key updated immediately
	_, err = r.etcdAPI.RefreshTTL(roomKey, roomTTL)
	first := r.etcdAPI.IsNotFound(err)
	if err != nil && !first {
		return err
	}
	// update value with TTL
	if _, err = r.etcdAPI.SetWithTTL(etcdKey, value, ttl); err != nil {
		return err
	}
	// set room key TTL for newly created dir
	if first {
		_, err = r.etcdAPI.RefreshTTL(roomKey, roomTTL)
		return err
	}
	return nil
}

func (r *coreImpl) Join(conn pierce.SocketConnection) {
	// push to event loop
	r.chJoin <- conn
}

func (r *coreImpl) Leave(conn pierce.SocketConnection) {
	// push to event loop
	r.chLeave <- conn
}

// toEtcdKey converts room + key to etcd key
func (r *coreImpl) toEtcdKey(roomMeta pierce.RoomMeta, key string) string {
	roomHash := fmt.Sprintf("%x", md5.Sum([]byte(roomMeta.ID)))
	roomKey := fmt.Sprintf("%s/%s/%s/%s/%s", r.keyPrefix, roomMeta.Namespace,
		roomHash[0:2], roomHash[2:4], roomMeta.ID)
	if key == "" {
		return roomKey
	}
	return fmt.Sprintf("%s/%s", roomKey, key)
}

func (r *coreImpl) mainLoop() {
	// leverage event loop to avoid any racing conditions
	// how to make sure alreay watching ?!
	maintain := time.NewTicker(maintainInterval).C
	for r.loopOnce(maintain) {
	}
}

func (r *coreImpl) loopOnce(maintain <-chan time.Time) bool {
	select {
	case <-r.chDone:
		// leave main loop
		r.logm.Info("core leave main loop")
		for _, room := range r.rooms {
			room.Stop()
		}
		return false

	case conn := <-r.chJoin:
		for _, room := range conn.Rooms() {
			r.ensureRoom(room).Join(conn)
		}

	case conn := <-r.chLeave:
		for _, room := range conn.Rooms() {
			r.ensureRoom(room).Leave(conn)
		}

	case <-maintain:
		// clean unnecessary room
		// resend for fail ?!
		r.maintain()
	}
	return true
}

// maintain cleans up emtpy room
func (r *coreImpl) maintain() {
	// cleanup empty room
	for roomMeta, room := range r.rooms {
		if room.Empty() {
			r.logm.Infof("remove empty room %v", roomMeta)
			room.Stop()
			delete(r.rooms, roomMeta)
		}
	}
}

// ensureRoom returns the room of the give roomID, and creates
// the room if necessary
func (r *coreImpl) ensureRoom(roomMeta pierce.RoomMeta) Room {
	// how to implement ?!
	room, ok := r.rooms[roomMeta]
	if ok {
		return room
	}

	r.logm.Infof("create room %v", roomMeta)
	room = r.rFactory(roomMeta, r.toEtcdKey(roomMeta, ""))
	r.rooms[roomMeta] = room
	room.Start()

	return room
}
