package socket

import (
	"fmt"
	"net/http"

	socketio "github.com/googollee/go-socket.io"

	"straas.io/base/logmetric"
	"straas.io/pierce"
)

// NewServer creates an instance of socket server
func NewServer(coreMgr pierce.Core, logm logmetric.LogMetric) *Server {
	return &Server{
		coreMgr: coreMgr,
		logm:    logm,
	}
}

// Server handles socket.io events
type Server struct {
	coreMgr pierce.Core
	logm    logmetric.LogMetric
}

// Create creates a http handler for socket server
func (s *Server) Create() (http.Handler, error) {
	server, err := socketio.NewServer(nil)

	if err != nil {
		return nil, fmt.Errorf("fail to create socketio server, err:%v", err)
	}

	// short path
	logm := s.logm
	coreMgr := s.coreMgr

	// IMPORTANT:
	// there is a problem, golang socket.io will trigger
	// connection event twice, and we should be the second
	// socket, now socket askes client side to send join event for simplify the problem
	server.On("connection", func(so socketio.Socket) {
		logm.Info("url", so.Request().URL)
		logm.BumpSum("socket.conn", 1)

		var conn pierce.SocketConnection
		var err error

		err = so.On("join", func(msg string) {
			conn = NewConn(so, []pierce.RoomMeta{
				pierce.RoomMeta{
					Namespace: "xxx",
					ID:        "aaa",
				},
				pierce.RoomMeta{
					Namespace: "xxx",
					ID:        "bbb",
				},
			}, logm)
			coreMgr.Join(conn)
		})
		if err != nil {
			// only get error when caller mis-uses the api
			logm.Fatalf("fail to listen join event, err:%v", err)
		}

		err = so.On("disconnection", func() {
			logm.Infof("disconnect %s", so.Id())
			logm.BumpSum("socket.disconn", 1)
			if conn != nil {
				coreMgr.Leave(conn)
			}
		})
		if err != nil {
			// only get error when caller mis-uses the api
			logm.Fatalf("fail to listen disconnect event, err:%v", err)
		}
	})

	server.On("error", func(so socketio.Socket, err error) {
		logm.BumpSum("socket.err", 1)
		logm.Errorf("socket fail, err: %v", err)
	})

	return server, nil
}
