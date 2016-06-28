package core

import (
	"fmt"

	"github.com/coreos/etcd/client"

	"straas.io/base/logmetric"
	"straas.io/external"
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

func newRoom(roomMeta pierce.RoomMeta, etcdKey string, etcdAPI external.Etcd,
	logm logmetric.LogMetric) Room {

	logm.Debugf("create room with key %s", etcdKey)
	return &roomImpl{
		roomMeta:   roomMeta,
		conns:      map[pierce.SocketConnection]bool{},
		connJoined: map[pierce.SocketConnection]bool{},

		etcdAPI: etcdAPI,
		etcdKey: etcdKey,
		chJoin:  make(chan pierce.SocketConnection, 10),
		chLeave: make(chan pierce.SocketConnection, 10),
		chDone:  make(chan bool),

		logm: logm,
	}
}

type roomImpl struct {
	// room meta
	roomMeta pierce.RoomMeta
	// all pierce.SocketConnections in this room
	connJoined map[pierce.SocketConnection]bool
	// keep track real pierce.SocketConnection count (some might still in the channel)
	conns   map[pierce.SocketConnection]bool
	etcdAPI external.Etcd
	// channels
	chJoin  chan pierce.SocketConnection
	chLeave chan pierce.SocketConnection
	chDone  chan bool
	//data
	etcdKey string
	data    interface{}
	version uint64
	// other
	logm logmetric.LogMetric
}

func (r *roomImpl) Start() {
	go r.mainLoop()
}

func (r *roomImpl) Stop() {
	close(r.chDone)
	close(r.chJoin)
	close(r.chLeave)
}

func (r *roomImpl) Empty() bool {
	return len(r.conns) == 0
}

func (r *roomImpl) Join(conn pierce.SocketConnection) {
	r.logm.Infof("connection %s join %v", conn.ID(), r.roomMeta)
	r.logm.BumpSum("core.room.join", 1)
	r.conns[conn] = true
	r.chJoin <- conn
}

func (r *roomImpl) Leave(conn pierce.SocketConnection) {
	r.logm.Infof("connection %s leave %v", conn.ID(), r.roomMeta)
	r.logm.BumpSum("core.room.leave", 1)
	delete(r.conns, conn)
	r.chLeave <- conn
}

func (r *roomImpl) join(conn pierce.SocketConnection) {
	r.connJoined[conn] = true
	r.logm.Infof("there %d conns in room %v", len(r.connJoined), r.roomMeta)

	// send if has data
	if r.version > 0 {
		conn.Emit(r.roomMeta, r.data, r.version)
	}
}

func (r *roomImpl) leave(conn pierce.SocketConnection) {
	// TODO: furthor notification ?!
	delete(r.connJoined, conn)
}

func (r *roomImpl) mainLoop() {
	wch := make(chan *client.Response, 10)
	go func() {
		r.etcdAPI.GetAndWatch(r.etcdKey, wch, r.chDone)
		close(wch)
	}()
	for r.loopOnce(wch) {
	}
}

func (r *roomImpl) loopOnce(wch <-chan *client.Response) bool {
	select {
	case <-r.chDone:
		return false

	case conn := <-r.chJoin:
		r.join(conn)

	case conn := <-r.chLeave:
		r.leave(conn)

	case resp := <-wch:
		// apply change does not involve any IO operation
		// it should be no error
		if err := r.applyChange(resp); err != nil {
			r.logm.Errorf("fail to apply resp %+v, err:%v", resp, err)
			return true
		}
		r.broadcast()
	}
	return true
}

func (r *roomImpl) applyChange(resp *client.Response) error {
	r.logm.Debugf("%+v\n", resp)
	cur := resp.Node

	// get room and key from etcd key
	key, err := subkey(r.etcdKey, cur.Key)
	if err != nil {
		r.logm.BumpSum("core.room.illegal_key.err", 1)
		return err
	}

	data, version, err := toValue(cur, unmarshaller)
	if err != nil {
		r.logm.BumpSum("core.room.to_value.err", 1)
		return err
	}
	// older changes, just ignore it
	if version <= r.version {
		r.logm.BumpSum("core.room.old_version", 1)
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
		// should not reach here
		r.logm.BumpSum("core.room.unknown_action.err", 1)
		return fmt.Errorf("unknown action %s", resp.Action)
	}
	return nil
}

func (r *roomImpl) broadcast() {
	r.logm.BumpSum("core.room.broadcast", 1)
	for conn := range r.connJoined {
		conn.Emit(r.roomMeta, r.data, r.version)
	}
}
