package etcd

import (
	"fmt"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"

	"straas.io/base/logger"
	"straas.io/base/metric"
)

const (
	testKey   = "/aaa"
	testValue = "1234"
)

func TestEtcd(t *testing.T) {
	suite.Run(t, new(etcdTestSuite))
}

type etcdTestSuite struct {
	suite.Suite
	impl   *etcdImpl
	keyAPI *keysAPIMock
}

func (s *etcdTestSuite) SetupTest() {
	s.keyAPI = &keysAPIMock{}
	s.impl = &etcdImpl{
		log:  logger.Get(),
		stat: metric.New("test"),
		api:  s.keyAPI,
	}
}

func (s *etcdTestSuite) TestEmptyDirResponse() {
	resp := emptyDirResponse("/aaa", 333)
	s.Equal(resp, &client.Response{
		Action: "get",
		Index:  333,
		Node: &client.Node{
			Dir: true,
			Key: "/aaa",
		},
	})
}

func (s *etcdTestSuite) TestGet() {
	opt := &client.GetOptions{
		Recursive: true,
	}
	testResp := &client.Response{
		Action: "get",
		Node: &client.Node{
			Dir:   false,
			Key:   testKey,
			Value: testValue,
		},
	}

	s.keyAPI.On("Get", mock.Anything, testKey, opt).Return(testResp, nil).Once()
	resp, err := s.impl.Get(testKey, true)
	s.NoError(err)
	s.Equal(resp, testResp)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestGetError() {
	opt := &client.GetOptions{
		Recursive: true,
	}
	s.keyAPI.On("Get", mock.Anything, testKey, opt).Return(nil, fmt.Errorf("some")).Once()
	_, err := s.impl.Get(testKey, true)
	s.Error(err)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestGetNotFound() {
	opt := &client.GetOptions{
		Recursive: true,
	}
	testResp := &client.Response{
		Action: "get",
		Index:  333,
		Node: &client.Node{
			Dir: true,
			Key: testKey,
		},
	}
	testErr := client.Error{
		Code:  client.ErrorCodeKeyNotFound,
		Index: 333,
	}

	s.keyAPI.On("Get", mock.Anything, testKey, opt).Return(nil, testErr).Once()
	resp, err := s.impl.Get(testKey, true)
	s.NoError(err)
	s.Equal(resp, testResp)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestSet() {
	opt := &client.SetOptions{}
	testResp := &client.Response{
		Action: "set",
		Node: &client.Node{
			Dir:   false,
			Key:   testKey,
			Value: testValue,
		},
	}

	s.keyAPI.On("Set", mock.Anything, testKey, testValue, opt).Return(testResp, nil).Once()
	resp, err := s.impl.Set(testKey, testValue)
	s.NoError(err)
	s.Equal(resp, testResp)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestSetError() {
	opt := &client.SetOptions{}
	s.keyAPI.On("Set", mock.Anything, testKey, testValue, opt).Return(nil, fmt.Errorf("some")).Once()
	_, err := s.impl.Set(testKey, testValue)
	s.Error(err)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestWatchAndGet() {
	done := make(chan bool)

	testIdx := uint64(333)
	getOpt := &client.GetOptions{
		Recursive: true,
	}
	testResp := &client.Response{
		Action: "get",
		Index:  testIdx,
		Node: &client.Node{
			Dir: true,
			Key: testKey,
		},
	}
	watchOpt := &client.WatcherOptions{
		Recursive:  true,
		AfterIndex: testIdx,
	}

	change1 := &client.Response{
		Action: "set",
		Index:  testIdx + 1,
	}
	change2 := &client.Response{
		Action: "set",
		Index:  testIdx + 2,
	}

	changes := make(chan *client.Response, 10)
	respRtnFn := func(context.Context) *client.Response {
		return <-changes
	}
	mwatcher := &watcherMock{}
	mwatcher.On("Next", mock.Anything).Return(respRtnFn, nil)

	s.keyAPI.On("Get", mock.Anything, testKey, getOpt).Return(testResp, nil).Once()
	s.keyAPI.On("Watcher", testKey, watchOpt).Return(mwatcher).Once()

	wch := s.impl.GetAndWatch(testKey, done)

	changes <- change1
	changes <- change2

	s.Equal(<-wch, testResp)
	s.Equal(<-wch, change1)
	s.Equal(<-wch, change2)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestGetAndWatchOutdate() {
	done := make(chan bool)

	testIdx := uint64(333)
	getOpt := &client.GetOptions{
		Recursive: true,
	}
	testResp := &client.Response{
		Action: "get",
		Index:  testIdx,
		Node: &client.Node{
			Dir: true,
			Key: testKey,
		},
	}
	watchOpt := &client.WatcherOptions{
		Recursive:  true,
		AfterIndex: testIdx,
	}

	err1 := &client.Error{
		Code: client.ErrorCodeEventIndexCleared,
	}
	change2 := &client.Response{
		Action: "set",
		Index:  testIdx + 2,
	}

	changes := make(chan *client.Response, 10)
	respRtnFn := func(context.Context) *client.Response {
		return <-changes
	}
	errors := make(chan error, 10)
	errRtnFn := func(context.Context) error {
		return <-errors
	}

	mwatcher := &watcherMock{}
	mwatcher.On("Next", mock.Anything).Return(respRtnFn, errRtnFn)

	s.keyAPI.On("Get", mock.Anything, testKey, getOpt).Return(testResp, nil).Twice()
	s.keyAPI.On("Watcher", testKey, watchOpt).Return(mwatcher).Twice()

	wch := s.impl.GetAndWatch(testKey, done)

	changes <- nil
	errors <- err1
	changes <- change2
	errors <- nil

	s.Equal(<-wch, testResp)
	s.Equal(<-wch, testResp)
	s.Equal(<-wch, change2)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestGetAndWatchError() {
	done := make(chan bool)

	testIdx := uint64(333)
	getOpt := &client.GetOptions{
		Recursive: true,
	}
	testResp := &client.Response{
		Action: "get",
		Index:  testIdx,
		Node: &client.Node{
			Dir: true,
			Key: testKey,
		},
	}
	watchOpt := &client.WatcherOptions{
		Recursive:  true,
		AfterIndex: testIdx,
	}

	err1 := fmt.Errorf("some error")
	change2 := &client.Response{
		Action: "set",
		Index:  testIdx + 2,
	}

	changes := make(chan *client.Response, 10)
	respRtnFn := func(context.Context) *client.Response {
		return <-changes
	}
	errors := make(chan error, 10)
	errRtnFn := func(context.Context) error {
		return <-errors
	}

	mwatcher := &watcherMock{}
	mwatcher.On("Next", mock.Anything).Return(respRtnFn, errRtnFn)

	s.keyAPI.On("Get", mock.Anything, testKey, getOpt).Return(testResp, nil).Once()
	s.keyAPI.On("Watcher", testKey, watchOpt).Return(mwatcher).Once()

	wch := s.impl.GetAndWatch(testKey, done)

	changes <- nil
	errors <- err1
	changes <- change2
	errors <- nil

	s.Equal(<-wch, testResp)
	s.Equal(<-wch, change2)
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestDone() {
	done := make(chan bool)
	fin := make(chan bool)

	close(done)

	wch := s.impl.getAndWatch(testKey, done, fin)

	// wait getAndWatch done
	<-fin
	select {
	case <-wch:
		assert.Fail(s.T(), "should be no resp")
	default:
	}
	s.keyAPI.AssertExpectations(s.T())
}

func (s *etcdTestSuite) TestDoneWhenNext() {
	done := make(chan bool)
	fin := make(chan bool)

	testIdx := uint64(333)
	getOpt := &client.GetOptions{
		Recursive: true,
	}
	testResp := &client.Response{
		Action: "get",
		Index:  testIdx,
		Node: &client.Node{
			Dir: true,
			Key: testKey,
		},
	}
	watchOpt := &client.WatcherOptions{
		Recursive:  true,
		AfterIndex: testIdx,
	}

	changes := make(chan *client.Response, 10)
	respRtnFn := func(context.Context) *client.Response {
		close(done)
		return <-changes
	}

	mwatcher := &watcherMock{}
	mwatcher.On("Next", mock.Anything).Return(respRtnFn, nil).Once()
	s.keyAPI.On("Get", mock.Anything, testKey, getOpt).Return(testResp, nil).Once()
	s.keyAPI.On("Watcher", testKey, watchOpt).Return(mwatcher).Once()

	wch := s.impl.getAndWatch(testKey, done, fin)
	s.Equal(<-wch, testResp)

	// wait getAndWatch done
	<-fin
	select {
	case <-wch:
		assert.Fail(s.T(), "should be no resp")
	default:
	}
	s.keyAPI.AssertExpectations(s.T())
}
