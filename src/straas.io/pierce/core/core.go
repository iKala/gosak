package core

import (
	"fmt"
	"time"

	"straas.io/base/logger"
	"straas.io/external"
	"straas.io/pierce"
)

const (
	maintainInterval = 30 * time.Second
	// room TTL is 3 days
	roomTTL = 3 * 24 * time.Hour
)

var (
	log = logger.Get()
)

// NewCore creates an instance of core manager
func NewCore(etcdAPI external.Etcd, keyPrefix string) pierce.Core {
	rFactory := func(roomID, etcdKey string) Room {
		return newRoom(roomID, etcdKey, etcdAPI)
	}
	return &coreImpl{
		rooms:     map[string]Room{},
		etcdAPI:   etcdAPI,
		rFactory:  rFactory,
		keyPrefix: keyPrefix,
		chJoin:    make(chan pierce.SocketConnection, 1000),
		chLeave:   make(chan pierce.SocketConnection, 1000),
		chDone:    make(chan bool),
	}
}

type roomFactory func(string, string) Room

type coreImpl struct {
	rooms     map[string]Room
	etcdAPI   external.Etcd
	rFactory  roomFactory
	keyPrefix string

	// channels
	chJoin  chan pierce.SocketConnection
	chLeave chan pierce.SocketConnection
	chDone  chan bool
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

func (r *coreImpl) Get(roomID, key string) (interface{}, error) {
	etcdKey := r.toEtcdKey(roomID, key)
	resp, err := r.etcdAPI.Get(etcdKey, true)
	if err != nil {
		return nil, err
	}
	v, _, err := toValue(resp.Node, unmarshaller)
	return v, err
}

func (r *coreImpl) GetAll(roomID string) (interface{}, error) {
	return r.Get(roomID, "")
}

func (r *coreImpl) Set(roomID, key string, v interface{}, ttl time.Duration) error {
	value, err := marshaller(v)
	if err != nil {
		return err
	}

	roomKey := r.toEtcdKey(roomID, "")
	etcdKey := r.toEtcdKey(roomID, key)

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
func (r *coreImpl) toEtcdKey(room, key string) string {
	if key == "" {
		return fmt.Sprintf("%s/%s", r.keyPrefix, room)
	}
	return fmt.Sprintf("%s/%s/%s", r.keyPrefix, room, key)
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
		log.Info("core leave main loop")
		for _, room := range r.rooms {
			room.Stop()
		}
		return false

	case conn := <-r.chJoin:
		for _, roomID := range conn.RoomIds() {
			r.ensureRoom(roomID).Join(conn)
		}

	case conn := <-r.chLeave:
		for _, roomID := range conn.RoomIds() {
			r.ensureRoom(roomID).Leave(conn)
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
	for roomID, room := range r.rooms {
		if room.Empty() {
			log.Infof("remove empty room %s", roomID)
			room.Stop()
			delete(r.rooms, roomID)
		}
	}
}

// ensureRoom returns the room of the give roomID, and creates
// the room if necessary
func (r *coreImpl) ensureRoom(roomID string) Room {
	// how to implement ?!
	room, ok := r.rooms[roomID]
	if ok {
		return room
	}

	log.Infof("create room %s", roomID)
	room = r.rFactory(roomID, r.toEtcdKey(roomID, ""))
	r.rooms[roomID] = room
	room.Start()

	return room
}
