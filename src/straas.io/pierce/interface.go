package pierce

// Core defines an interface Pierce core operations
type Core interface {
	// Start starts pierce core
	Start()
	// Stop stops pierce core
	Stop()
	// Get gets data of the give room and key
	Get(roomId, key string) (interface{}, error)
	// Get gets all data of the give room
	GetAll(roomId string) (map[string]interface{}, error)
	// Set sets data of the given room and key
	Set(roomId, key string, value interface{}) error
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
	Emit(data string, version uint64)
}
