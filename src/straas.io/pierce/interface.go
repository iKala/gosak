package pierce

import (
	"fmt"
	"time"
)

// Core defines an interface Pierce core operations
type Core interface {
	// Start starts pierce core
	Start()
	// Stop stops pierce core
	Stop()
	// Get gets data of the give room and key
	Get(room RoomMeta, key string) (interface{}, error)
	// Get gets all data of the give room
	GetAll(room RoomMeta) (interface{}, error)
	// Set sets data of the given room and key
	Set(room RoomMeta, key string, value interface{}, ttl time.Duration) error
	// Join adds socket connections
	Join(SocketConnection)
	// Leave removes socket connection
	Leave(SocketConnection)
}

// RoomMeta includes namespace and id of a room
type RoomMeta struct {
	// Namespace of the room
	Namespace string
	// RoomId of the room
	ID string
}

func (r *RoomMeta) String() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.ID)
}

// SocketConnection defines a wrapper interface to abstract web socket
type SocketConnection interface {
	// Rooms returns belonged roooms of the connection
	Rooms() []RoomMeta
	// Emit sends data to connection
	Emit(room RoomMeta, data interface{}, version uint64)
	// ID returns connection id
	ID() string
}
