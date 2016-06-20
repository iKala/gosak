package socket

import (
	"fmt"
	"net/http"

	socketio "github.com/googollee/go-socket.io"

	"straas.io/base/logger"
	"straas.io/pierce"
)

var (
	log = logger.Get()
)

// NewServer creates an instance of socket server
func NewServer(coreMgr pierce.Core) *Server {
	return &Server{
		coreMgr: coreMgr,
	}
}

// Server handles socket.io events
type Server struct {
	coreMgr pierce.Core
}

// Create creates a http handler for socket server
func (s *Server) Create() (http.Handler, error) {
	server, err := socketio.NewServer(nil)

	if err != nil {
		return nil, fmt.Errorf("fail to create socketio server, err:%v", err)
	}

	// short path
	coreMgr := s.coreMgr

	// IMPORTANT:
	// there is a problem, golang socket.io will trigger
	// connection event twice, and we should be the second
	// socket, now socket askes client side to send join event for simplify the problem
	server.On("connection", func(so socketio.Socket) {
		log.Info("url", so.Request().URL)

		var conn pierce.SocketConnection
		so.On("join", func(msg string) {
			conn = NewConn(so, []string{"aaa", "bbb"})
			coreMgr.Join(conn)
		})

		so.On("disconnection", func() {
			log.Infof("disconnect %s", so.Id())
			if conn != nil {
				coreMgr.Leave(conn)
			}
		})
	})

	server.On("error", func(so socketio.Socket, err error) {
		// TODO: record more
		log.Errorf("socket fail, err: %v", err)
	})

	return server, nil
}
