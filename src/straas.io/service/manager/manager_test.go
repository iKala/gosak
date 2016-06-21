package manager

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"straas.io/service/common"
)

func TestManager(t *testing.T) {
	suite.Run(t, new(ManangerTestSuite))
}

type ManangerTestSuite struct {
	suite.Suite
	services map[common.ServiceType]common.Service
}

func (s *ManangerTestSuite) SetupTest() {
	// clean up services
	s.services = map[common.ServiceType]common.Service{}
}

func (s *ManangerTestSuite) TestEmptyManager() {
	s.Assert().Panics(func() {
		newMgr(s.services)
	})
}

func (s *ManangerTestSuite) TestInitError() {
	s1 := &common.ServiceMock{}
	s1.On("Type").Return(common.ServiceType("s1"))
	s1.On("Dependencies").Return([]common.ServiceType{})
	s1.On("New", mock.Anything).Return(nil, fmt.Errorf("some err")).Once()
	s1.On("AddFlags").Return().Once()

	s.services["s1"] = s1

	m := newMgr(s.services, "s1")
	s.Error(m.Init())
	// second init
	s.Assert().Panics(func() {
		m.Init()
	})
}

func (s *ManangerTestSuite) TestManager() {
	var m *managerImpl

	v1, v2, v3 := interface{}(1), interface{}(2), interface{}(3)

	s1 := &common.ServiceMock{}
	s2 := &common.ServiceMock{}
	s3 := &common.ServiceMock{}
	s1.On("Dependencies").Return([]common.ServiceType{"s2", "s3"})
	s1.On("New", mock.Anything).Return(func(common.ServiceGetter) interface{} {
		// dependencies must already inited
		s.Equal(v3, m.MustGet("s3"))
		s.Equal(v2, m.MustGet("s2"))
		return v1
	}, nil).Once()
	s1.On("AddFlags").Return().Once()

	s2.On("Dependencies").Return([]common.ServiceType{"s3"})
	s2.On("New", mock.Anything).Return(func(common.ServiceGetter) interface{} {
		// dependencies must already inited
		s.Equal(v3, m.MustGet("s3"))
		return v2
	}, nil).Once()
	s2.On("AddFlags").Return().Once()

	s3.On("Dependencies").Return([]common.ServiceType{})
	s3.On("New", mock.Anything).Return(v3, nil).Once()
	s3.On("AddFlags").Return().Once()

	s.services["s1"] = s1
	s.services["s2"] = s2
	s.services["s3"] = s3

	m = newMgr(s.services, "s1").(*managerImpl)
	s.Assert().Panics(func() {
		// no-existed service
		newMgr(s.services, "s4")
	})

	_, err := m.Get("s1")
	s.Error(err)

	s.NoError(m.Init())

	s.Assert().Panics(func() {
		m.Init()
	})

	v, err := m.Get("s1")
	s.NoError(err)
	s.Equal(v, v1)

	v, err = m.Get("s2")
	s.NoError(err)
	s.Equal(v, v2)

	v, err = m.Get("s3")
	s.NoError(err)
	s.Equal(v, v3)

	_, err = m.Get("s4")
	s.Error(err)

	s1.AssertExpectations(s.T())
	s2.AssertExpectations(s.T())
	s3.AssertExpectations(s.T())
}
