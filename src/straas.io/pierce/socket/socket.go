package socket

import (
	"fmt"
	"net/http"
	"sync"

	socketio "github.com/googollee/go-socket.io"

	"straas.io/base/logger"
	"straas.io/pierce"
)

var (
	log = logger.Get()
)

// NewSocketServer creates an instance of socket server
func NewSockerServer(coreMgr pierce.Core) *SocketServer {
	return &SocketServer{
		coreMgr:  coreMgr,
		connLock: sync.Mutex{},
		conns:    map[string]int{},
	}
}

// SocketServer handles socket.io events
type SocketServer struct {
	coreMgr  pierce.Core
	connLock sync.Mutex
	conns    map[string]int
}

// Create creates a http handler for socket server
func (s *SocketServer) Create() (http.Handler, error) {
	server, err := socketio.NewServer(nil)

	if err != nil {
		return nil, fmt.Errorf("fail to create socketio server, err:%v", err)
	}

	// short path
	conns := s.conns
	connLock := s.connLock
	coreMgr := s.coreMgr

	// IMPORTANT:
	// there is a problem, golang socket.io will trigger
	// connection event twice, and we should be the second
	// socket
	server.On("connection", func(so socketio.Socket) {
		log.Info("url", so.Request().URL)

		// already in
		var conn pierce.SocketConnection
		var connId = so.Id()

		connLock.Lock()
		switch v := conns[connId]; v {
		// first connect event
		case 0:
			conns[connId]++
			connLock.Unlock()

		// second connect event
		case 1:
			conns[connId]++
			connLock.Unlock()
			// TODO: extra room from token
			conn = NewConn(so, []string{"aaa", "bbb"})
			coreMgr.Join(conn)

		default:
			connLock.Unlock()
			return
		}
		so.On("disconnection", func() {
			log.Infof("disconnect %s", so.Id())

			connLock.Lock()
			delete(conns, connId)
			connLock.Unlock()

			if conn != nil {
				coreMgr.Leave(conn)
			}
		})
	})

	server.On("error", func(so socketio.Socket, err error) {
		// TODO: record more
		log.Errorf("socket fail error:", err)
	})

	return server, nil
}
