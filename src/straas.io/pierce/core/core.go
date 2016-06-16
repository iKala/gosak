package core

import (
	"fmt"
	"time"

	"straas.io/base/etcd"
	"straas.io/base/logger"
	"straas.io/pierce"
)

const (
	maintainInterval = 30 * time.Second
)

var (
	log = logger.Get()
)

// NewCore creates an instance of core manager
func NewCore(etcdAPI etcd.Etcd) pierce.Core {
	rFactory := func(roomId, etcdKey string) Room {
		return newRoom(roomId, etcdKey, etcdAPI)
	}
	return &coreImpl{
		rooms:    map[string]Room{},
		etcdAPI:  etcdAPI,
		rFactory: rFactory,
		chJoin:   make(chan pierce.SocketConnection, 1000),
		chLeave:  make(chan pierce.SocketConnection, 1000),
		chDone:   make(chan bool),
	}
}

// toEtcdKey converts room + key to etcd key
// TODO: need to design convention
func toEtcdKey(room, key string) string {
	if key == "" {
		return fmt.Sprintf("/pierce/%s", room)
	}
	return fmt.Sprintf("/pierce/%s/%s", room, key)
}

type roomFactory func(string, string) Room

type coreImpl struct {
	rooms    map[string]Room
	etcdAPI  etcd.Etcd
	rFactory roomFactory

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

func (r *coreImpl) Get(roomId, key string) (interface{}, error) {
	etcdKey := toEtcdKey(roomId, key)
	resp, err := r.etcdAPI.Get(etcdKey, true)
	if err != nil {
		return nil, err
	}
	v, _, err := toValue(resp.Node, unmarshaller)
	return v, err
}

func (r *coreImpl) GetAll(roomId string) (interface{}, error) {
	return r.Get(roomId, "")
}

func (r *coreImpl) Set(roomId, key string, v interface{}) error {
	value, err := marshaller(v)
	if err != nil {
		return err
	}
	etcdKey := toEtcdKey(roomId, key)
	_, err = r.etcdAPI.Set(etcdKey, value)
	return err
}

func (r *coreImpl) Join(conn pierce.SocketConnection) {
	// push to event loop
	r.chJoin <- conn
}

func (r *coreImpl) Leave(conn pierce.SocketConnection) {
	// push to event loop
	r.chLeave <- conn
}

func (r *coreImpl) mainLoop() {
	// leverage event loop to avoid any racing conditions
	// how to make sure alreay watching ?!
	maintain := time.NewTicker(maintainInterval).C
	for {
		select {
		case <-r.chDone:
			// leave main loop
			log.Info("core leave main loop")
			for _, room := range r.rooms {
				room.Stop()
			}
			return

		case conn := <-r.chJoin:
			for _, roomId := range conn.RoomIds() {
				r.ensureRoom(roomId).Join(conn)
			}

		case conn := <-r.chLeave:
			for _, roomId := range conn.RoomIds() {
				r.ensureRoom(roomId).Leave(conn)
			}

		case <-maintain:
			// clean unnecessary room
			// resend for fail ?!
			r.maintain()
		}
	}
}

// maintain cleans up emtpy room
func (r *coreImpl) maintain() {
	// cleanup empty room
	for roomId, room := range r.rooms {
		if room.Empty() {
			log.Infof("remove empty room %s", roomId)
			room.Stop()
			delete(r.rooms, roomId)
		}
	}
}

// ensureRoom returns the room of the give roomId, and creates
// the room if necessary
func (r *coreImpl) ensureRoom(roomId string) Room {
	// how to implement ?!
	room, ok := r.rooms[roomId]
	if ok {
		return room
	}

	log.Infof("create room %s", roomId)
	room = r.rFactory(roomId, toEtcdKey(roomId, ""))
	r.rooms[roomId] = room
	room.Start()

	return room
}
