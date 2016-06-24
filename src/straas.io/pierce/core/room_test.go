package core

import (
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/suite"

	etcdMocks "straas.io/external/mocks"
	"straas.io/pierce"
	"straas.io/pierce/mocks"
)

const (
	testNamespace = "test-ns"
	testRoomID    = "test-room-id"
	testEtcdKey   = "/pierce/test-etcd-key"
)

var (
	testRoomMeta = pierce.RoomMeta{
		Namespace: testNamespace,
		ID:        testRoomID,
	}
)

func TestRoom(t *testing.T) {
	suite.Run(t, new(roomTestSuite))
}

type roomTestSuite struct {
	suite.Suite
	impl     *roomImpl
	etcdMock *etcdMocks.Etcd
}

func (s *roomTestSuite) SetupTest() {
	s.etcdMock = &etcdMocks.Etcd{}
	s.impl = newRoom(testRoomMeta, testEtcdKey, s.etcdMock).(*roomImpl)

}

func (s *roomTestSuite) TestJoin() {
	wch := make(chan *client.Response)

	c1 := &mocks.SocketConnection{}
	c2 := &mocks.SocketConnection{}
	c3 := &mocks.SocketConnection{}
	// for logs
	c1.On("ID").Return("conn1")
	c2.On("ID").Return("conn2")
	c3.On("ID").Return("conn3")

	// only c3 got emit
	c3.On("Emit", testRoomMeta, "xxx", uint64(10)).Return().Once()

	s.impl.Join(c1)
	s.impl.Join(c2)
	s.impl.loopOnce(wch)
	s.impl.loopOnce(wch)
	s.Equal(len(s.impl.conns), 2)

	s.impl.data = "xxx"
	s.impl.version = 10

	s.impl.Join(c3)
	s.impl.loopOnce(wch)
	s.Equal(len(s.impl.conns), 3)

	s.Equal(s.impl.connJoined, map[pierce.SocketConnection]bool{
		c1: true,
		c2: true,
		c3: true,
	})
	c1.AssertExpectations(s.T())
	c2.AssertExpectations(s.T())
	c3.AssertExpectations(s.T())
}

func (s *roomTestSuite) TestLeave() {
	wch := make(chan *client.Response)

	c1 := &mocks.SocketConnection{}
	c2 := &mocks.SocketConnection{}
	c3 := &mocks.SocketConnection{}
	// for logs
	c1.On("ID").Return("conn1")
	c2.On("ID").Return("conn2")

	s.impl.conns = map[pierce.SocketConnection]bool{
		c1: true,
		c2: true,
		c3: true,
	}
	s.impl.connJoined = map[pierce.SocketConnection]bool{
		c1: true,
		c2: true,
		c3: true,
	}

	s.impl.Leave(c1)
	s.impl.Leave(c2)
	s.impl.loopOnce(wch)
	s.impl.loopOnce(wch)

	s.Equal(len(s.impl.conns), 1)
	s.Equal(s.impl.connJoined, map[pierce.SocketConnection]bool{
		c3: true,
	})
	c1.AssertExpectations(s.T())
	c2.AssertExpectations(s.T())
	c3.AssertExpectations(s.T())
}

func (s *roomTestSuite) TestEmpty() {
	s.True(s.impl.Empty())

	c1 := &mocks.SocketConnection{}
	c2 := &mocks.SocketConnection{}
	c3 := &mocks.SocketConnection{}
	s.impl.conns = map[pierce.SocketConnection]bool{
		c1: true,
		c2: true,
		c3: true,
	}
	s.False(s.impl.Empty())
}

func (s *roomTestSuite) TestApplyChange() {
	wch := make(chan *client.Response, 10)

	resp1 := &client.Response{
		Action: "get",
		Node: &client.Node{
			Key:           testEtcdKey,
			Dir:           true,
			ModifiedIndex: 101,
			Nodes: []*client.Node{
				&client.Node{
					Key:           testEtcdKey + "/aaa",
					Dir:           false,
					ModifiedIndex: 102,
					Value:         "1234",
				},
			},
		},
	}
	resp2 := &client.Response{
		Action: "set",
		Node: &client.Node{
			Key:           testEtcdKey + "/bbb",
			Dir:           false,
			Value:         "4567",
			ModifiedIndex: 103,
		},
	}
	resp3 := &client.Response{
		Action: "update",
		Node: &client.Node{
			Key:           testEtcdKey + "/aaa",
			Dir:           false,
			Value:         "1357",
			ModifiedIndex: 104,
		},
	}
	resp4 := &client.Response{
		Action: "delete",
		Node: &client.Node{
			Key:           testEtcdKey + "/aaa",
			Dir:           false,
			ModifiedIndex: 105,
		},
	}
	resp5 := &client.Response{
		Action: "expire",
		Node: &client.Node{
			Key:           testEtcdKey + "/aaa",
			Dir:           false,
			ModifiedIndex: 106,
		},
	}
	resp6 := &client.Response{
		Action: "set",
		Node: &client.Node{
			Key:           testEtcdKey + "/aaa",
			Dir:           false,
			ModifiedIndex: 100,
		},
	}

	wch <- resp1
	wch <- resp2
	wch <- resp3
	wch <- resp4
	wch <- resp5
	wch <- resp6

	// test get
	s.impl.loopOnce(wch)
	s.Equal(s.impl.data, map[string]interface{}{
		"aaa": float64(1234),
	})
	s.Equal(s.impl.version, uint64(102))

	// test set
	s.impl.loopOnce(wch)
	s.Equal(s.impl.data, map[string]interface{}{
		"aaa": float64(1234),
		"bbb": float64(4567),
	})
	s.Equal(s.impl.version, uint64(103))

	// test update
	s.impl.loopOnce(wch)
	s.Equal(s.impl.data, map[string]interface{}{
		"aaa": float64(1357),
		"bbb": float64(4567),
	})
	s.Equal(s.impl.version, uint64(104))

	// test del
	s.impl.loopOnce(wch)
	s.Equal(s.impl.data, map[string]interface{}{
		"bbb": float64(4567),
	})
	s.Equal(s.impl.version, uint64(105))

	// test expire
	s.impl.loopOnce(wch)
	s.Equal(s.impl.data, map[string]interface{}{
		"bbb": float64(4567),
	})
	s.Equal(s.impl.version, uint64(106))

	// ignore old changes
	s.impl.loopOnce(wch)
	s.Equal(s.impl.data, map[string]interface{}{
		"bbb": float64(4567),
	})
	s.Equal(s.impl.version, uint64(106))
}

func (s *roomTestSuite) TestBroadcast() {
	c1 := &mocks.SocketConnection{}
	c2 := &mocks.SocketConnection{}

	// for logs
	c1.On("Emit", testRoomMeta, "1234", uint64(101)).Return().Once()
	c2.On("Emit", testRoomMeta, "1234", uint64(101)).Return().Once()

	s.impl.connJoined = map[pierce.SocketConnection]bool{
		c1: true,
		c2: true,
	}
	s.impl.data = "1234"
	s.impl.version = 101
	s.impl.broadcast()

	c1.AssertExpectations(s.T())
	c2.AssertExpectations(s.T())
}
