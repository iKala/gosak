package socket

import (
	socketio "github.com/googollee/go-socket.io"

	"straas.io/pierce"
)

// NewConn creates an instance to wrapper socket.io connect
func NewConn(socket socketio.Socket, roomIds []string) pierce.SocketConnection {
	return &connImpl{
		socket:  socket,
		roomIds: roomIds,
	}
}

type connImpl struct {
	socket  socketio.Socket
	version uint64 // current seen version
	roomIds []string
}

func (c *connImpl) Id() string {
	return c.socket.Id()
}

func (c *connImpl) RoomIds() []string {
	// copy ?!
	return c.roomIds
}

func (c *connImpl) Emit(data string, version uint64) {
	// skip old data
	if version <= c.version {
		return
	}
	c.version = version
	if err := c.socket.Emit("data", data); err != nil {
		// TODO: log & metric
		log.Errorf("emit data fail, err:%v", err)
		c.close()
	}
}

func (c *connImpl) close() {
	c.socket.Emit("disconnect")
}
