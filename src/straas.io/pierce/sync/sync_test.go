package sync

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	// register sqlite driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"straas.io/base/logmetric"
	"straas.io/pierce"
	"straas.io/pierce/mocks"
)

const (
	testNamespace  = "xxx"
	testNamespace2 = "yyy"
)

func TestSyncer(t *testing.T) {
	suite.Run(t, new(syncerTestSuite))
}

type syncerTestSuite struct {
	suite.Suite
	db         *gorm.DB
	impl       *syncerImpl
	coreMock   *mocks.Core
	sinkerMock *sinkerMock
}

func (s *syncerTestSuite) SetupTest() {
	var err error
	// create mock object of core manager
	s.coreMock = &mocks.Core{}
	s.sinkerMock = &sinkerMock{}

	// create in memory sqlite for test
	s.db, err = gorm.Open("sqlite3", "file::memory:")
	s.NoError(err)

	syncer, err := New(s.coreMock, s.db, logmetric.NewDummy())
	s.impl = syncer.(*syncerImpl)
	s.impl.sinker = s.sinkerMock

	s.NoError(err)
}

func (s *syncerTestSuite) TestAdd() {
	meta := pierce.RoomMeta{
		Namespace: testNamespace,
		ID:        "aaa",
	}

	s.impl.Add(meta)
	s.Equal(len(s.impl.queue), 1)

	// test with fail once
	s.coreMock.On("GetAll", meta).Return(nil, uint64(0), fmt.Errorf("some error")).Once()
	s.coreMock.On("GetAll", meta).Return("1234", uint64(33), nil).Twice()
	s.sinkerMock.On("Sink", meta, "1234", uint64(33)).Return(fmt.Errorf("some error")).Once()
	s.sinkerMock.On("Sink", meta, "1234", uint64(33)).Return(nil).Once()

	s.impl.syncOnce()
	s.coreMock.AssertExpectations(s.T())
	s.sinkerMock.AssertExpectations(s.T())
}

func (s *syncerTestSuite) TestDiff() {
	s.sinkerMock.On("Diff", "xxx", uint64(9), 100).Return([]Record{
		Record{
			ID:        10,
			Namespace: testNamespace,
			Room:      "aaa",
			Value:     "1234",
			Version:   555,
		},
		Record{
			ID:        12,
			Namespace: testNamespace,
			Room:      "ccc",
			Value:     "1235",
			Version:   666,
		},
	}, nil).Once()
	result, err := s.impl.Diff("xxx", uint64(9), 100)
	s.NoError(err)
	s.Equal(result, []pierce.Record{
		pierce.Record{
			Index: 10,
			Room: pierce.RoomMeta{
				Namespace: testNamespace,
				ID:        "aaa",
			},
			Value: float64(1234),
		},
		pierce.Record{
			Index: 12,
			Room: pierce.RoomMeta{
				Namespace: testNamespace,
				ID:        "ccc",
			},
			Value: float64(1235),
		},
	})
	s.sinkerMock.AssertExpectations(s.T())
}

func (s *syncerTestSuite) TestWatch() {
	idx := uint64(9)
	meta := pierce.RoomMeta{
		Namespace: testNamespace,
		ID:        "aaa",
	}
	// done := make(chan bool)

	s.sinkerMock.On("LatestVersion").Return(uint64(0), fmt.Errorf("some error")).Once()
	s.sinkerMock.On("LatestVersion").Return(idx, nil).Once()

	s.coreMock.On("Watch", idx, mock.Anything).Return(fmt.Errorf("some error")).Once()
	s.coreMock.On("Watch", idx, mock.Anything).Return(func(afterIdx uint64, ch chan<- pierce.RoomMeta) error {
		ch <- meta
		ch <- meta
		close(s.impl.chDone)
		return nil

	}).Once()

	s.impl.watch()
	s.Equal(len(s.impl.queue), 2)

	s.coreMock.AssertExpectations(s.T())
	s.sinkerMock.AssertExpectations(s.T())
}
