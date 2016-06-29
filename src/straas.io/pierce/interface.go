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
	Get(room RoomMeta, key string) (interface{}, uint64, error)
	// Get gets all data of the give room
	GetAll(room RoomMeta) (interface{}, uint64, error)
	// Set sets data of the given room and key
	Set(room RoomMeta, key string, value interface{}, ttl time.Duration) error
	// Join adds socket connections
	Join(SocketConnection)
	// Leave removes socket connection
	Leave(SocketConnection)
	// Watch watches changes
	Watch(afterVersion uint64, resp chan<- RoomMeta) error
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

// Record stores changed data of syncer
type Record struct {
	// Index of the Record
	Index uint64
	// Room meta of the Record
	Room RoomMeta
	// Value is value of the room
	Value interface{}
}

// Syncer defines an interface for sync
type Syncer interface {
	// Start starts syncer
	Start()
	// Stop stops syncer
	Stop()
	// Add adds room meta for sync
	Add(roomMeta RoomMeta)

	Diff(namespace string, afterIdx uint64, size int) ([]Record, error)
}
