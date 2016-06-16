package core

import (
	"fmt"

	"github.com/coreos/etcd/client"

	"straas.io/base/etcd"
	"straas.io/pierce"
)

// Room defines an interface for room operation
// Note that Room operation is not thread safe
type Room interface {
	// Start starts the room
	Start()
	// Stop stops the room
	Stop()
	// Join adds conn to the room
	Join(pierce.SocketConnection)
	// Leave removes conn to the room
	Leave(pierce.SocketConnection)
	// Empty checks if room is empty
	Empty() bool
}

func newRoom(id, etcdKey string, etcdAPI etcd.Etcd) Room {
	log.Debugf("create room with key %s", etcdKey)

	return &roomImpl{
		id:        id,
		conns:     map[pierce.SocketConnection]bool{},
		connCount: 0,
		etcdAPI:   etcdAPI,
		etcdKey:   etcdKey,
		chJoin:    make(chan pierce.SocketConnection, 10),
		chLeave:   make(chan pierce.SocketConnection, 10),
		chDone:    make(chan bool),
	}
}

type roomImpl struct {
	// room id
	id string
	// all pierce.SocketConnections in this room
	conns map[pierce.SocketConnection]bool
	// keep track real pierce.SocketConnection count (some might still in the channel)
	connCount int
	etcdAPI   etcd.Etcd
	// channels
	chJoin  chan pierce.SocketConnection
	chLeave chan pierce.SocketConnection
	chDone  chan bool

	etcdKey string
	data    interface{}
	dataStr string // cache data to avoid redundant marshalling
	version uint64
}

func (r *roomImpl) Start() {
	go r.mainLoop()
}

func (r *roomImpl) Stop() {
	close(r.chDone)
}

func (r *roomImpl) Empty() bool {
	return r.connCount == 0
}

func (r *roomImpl) Join(conn pierce.SocketConnection) {
	log.Infof("connection %s join %s", conn.Id(), r.id)
	r.connCount++
	r.chJoin <- conn
}

func (r *roomImpl) Leave(conn pierce.SocketConnection) {
	log.Infof("connection %s leave %s", conn.Id(), r.id)
	r.connCount--
	r.chLeave <- conn
}

func (r *roomImpl) join(conn pierce.SocketConnection) {
	r.conns[conn] = true
	log.Infof("there %d conns in room %s", len(r.conns), r.id)

	// send if has data
	if r.version > 0 {
		conn.Emit(r.dataStr, r.version)
	}
}

func (r *roomImpl) leave(conn pierce.SocketConnection) {
	// TODO: furthor notification ?!
	delete(r.conns, conn)
}

func (r *roomImpl) mainLoop() {
	// how to make sure alreay watching ?!
	wch := r.etcdAPI.GetAndWatch(r.etcdKey, r.chDone)
	for {
		select {
		case conn := <-r.chJoin:
			r.join(conn)

		case conn := <-r.chLeave:
			r.leave(conn)

		case resp := <-wch:
			// apply change does not involve any IO operation
			// it should be no error
			if err := r.applyChange(resp); err != nil {
				log.Errorf("fail to apply resp %+v, err:%v", resp, err)
				// WTD
				continue
			}
			r.broadcast()
		}
	}
}

func (r *roomImpl) applyChange(resp *client.Response) error {
	log.Debugf("%+v\n", resp)
	cur := resp.Node

	// get room and key from etcd key
	key, err := subkey(r.etcdKey, cur.Key)
	if err != nil {
		// illegal key
		return err
	}

	data, version, err := toValue(cur, unmarshaller)
	if err != nil {
		// WTF
		return err
	}
	// older changes, just ignore it
	if version <= r.version {
		// TODO: log
		return nil
	}
	r.version = version

	switch resp.Action {
	case "get":
		// only get all
		r.data = data

	case "create", "set", "update":
		r.data, err = setByPath(r.data, key, data)
		if err != nil {
			return err
		}

	case "delete", "expire":
		r.data, err = delByPath(r.data, key)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown action %s", resp.Action)
		// should not reach here
		// TODO: keep log and metrics
	}
	return nil
}

func (r *roomImpl) broadcast() {
	r.dataStr, _ = marshaller(r.data)

	// TODO: check previous value
	for conn := range r.conns {
		conn.Emit(r.dataStr, r.version)
	}
}
