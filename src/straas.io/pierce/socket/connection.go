package socket

import (
	socketio "github.com/googollee/go-socket.io"

	"straas.io/pierce"
)

// NewConn creates an instance to wrapper socket.io connect
func NewConn(socket socketio.Socket, roomMetas []pierce.RoomMeta) pierce.SocketConnection {
	return &connImpl{
		socket:    socket,
		versions:  map[pierce.RoomMeta]uint64{},
		roomMetas: roomMetas,
	}
}

type connImpl struct {
	socket    socketio.Socket
	versions  map[pierce.RoomMeta]uint64 // current seen version
	roomMetas []pierce.RoomMeta
}

func (c *connImpl) ID() string {
	return c.socket.Id()
}

func (c *connImpl) Rooms() []pierce.RoomMeta {
	return c.roomMetas
}

func (c *connImpl) Emit(roomMeta pierce.RoomMeta, data interface{}, version uint64) {
	// skip old data
	if version <= c.versions[roomMeta] {
		return
	}
	c.versions[roomMeta] = version
	if err := c.socket.Emit("data", roomMeta.Namespace, roomMeta.ID, data); err != nil {
		// TODO: log & metric
		log.Errorf("emit data fail, err:%v", err)
		c.close()
	}
}

func (c *connImpl) close() {
	c.socket.Emit("disconnect")
}
