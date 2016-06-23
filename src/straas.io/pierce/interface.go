package pierce

import (
	"time"
)

// Core defines an interface Pierce core operations
type Core interface {
	// Start starts pierce core
	Start()
	// Stop stops pierce core
	Stop()
	// Get gets data of the give room and key
	Get(roomID, key string) (interface{}, error)
	// Get gets all data of the give room
	GetAll(roomID string) (interface{}, error)
	// Set sets data of the given room and key
	Set(roomID, key string, value interface{}, ttl time.Duration) error
	// Join adds socket connections
	Join(SocketConnection)
	// Leave removes socket connection
	Leave(SocketConnection)
}

// SocketConnection defines a wrapper interface to abstract web socket
type SocketConnection interface {
	// RoomIds returns belonged roooms of the connection
	RoomIds() []string
	// Emit sends data to connection
	Emit(roomID, data string, version uint64)
	// ID returns connection id
	ID() string
}
