package socket

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"straas.io/external/mocks"
	"straas.io/pierce"
)

var (
	troom1 = pierce.RoomMeta{
		Namespace: "xxx",
		ID:        "aaa",
	}
	troom2 = pierce.RoomMeta{
		Namespace: "xxx",
		ID:        "bbb",
	}
	troom3 = pierce.RoomMeta{
		Namespace: "xxx",
		ID:        "ccc",
	}
)

func TestConnectionSuite(t *testing.T) {
	suite.Run(t, new(ConnectionTestSuite))
}

type ConnectionTestSuite struct {
	suite.Suite
}

func (s *ConnectionTestSuite) TestEmit() {

	sk := &mocks.SocketIO{}
	conn := NewConn(sk, []pierce.RoomMeta{troom1, troom2, troom3})

	sk.Mock.On("Emit", "data", []interface{}{"xxx", "aaa", "v1"}).Return(nil).Once()
	conn.Emit(troom1, "v1", 10)
	sk.AssertExpectations(s.T())

	// old data
	conn.Emit(troom1, "v1", 9)
	sk.AssertExpectations(s.T())

	// another room
	sk.Mock.On("Emit", "data", []interface{}{"xxx", "bbb", "v2"}).Return(nil).Once()
	conn.Emit(troom2, "v2", 8)
	sk.AssertExpectations(s.T())

	// fail
	sk.Mock.On("Emit", "data", []interface{}{"xxx", "ccc", "v3"}).Return(fmt.Errorf("some err")).Once()
	sk.Mock.On("Emit", "disconnect", []interface{}(nil)).Return(nil).Once()
	conn.Emit(troom3, "v3", 11)
	sk.AssertExpectations(s.T())

}
