package common

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

type ServiceTestSuite struct {
	suite.Suite
}

func (s *ServiceTestSuite) SetupTest() {
	// clean up services
	services = map[ServiceType]Service{}
}

func (s *ServiceTestSuite) TestDependOn() {
	s1 := &ServiceMock{}
	s2 := &ServiceMock{}
	s3 := &ServiceMock{}
	s4 := &ServiceMock{}
	s5 := &ServiceMock{}
	s6 := &ServiceMock{}

	s1.On("Dependencies").Return([]ServiceType{"s2", "s3"})
	s2.On("Dependencies").Return([]ServiceType{"s3", "s7"})
	s3.On("Dependencies").Return([]ServiceType{})
	s4.On("Dependencies").Return([]ServiceType{"s1"})
	s5.On("Dependencies").Return([]ServiceType{"s1", "s9"})
	s6.On("Dependencies").Return([]ServiceType{"s2"})

	services = map[ServiceType]Service{
		"s1": s1,
		"s2": s2,
		"s3": s3,
		"s4": s4,
		"s5": s5,
		"s6": s6,
	}
	s.True(dependOn("s1", "s2"))
	s.False(dependOn("s2", "s1"))
	s.True(dependOn("s6", "s3"))
	s.False(dependOn("s3", "s6"))
	s.False(dependOn("s4", "s5"))
	s.False(dependOn("s5", "s4"))
	s.False(dependOn("s7", "s9"))
}

func (s *ServiceTestSuite) TestRegister() {
	s1 := &ServiceMock{}
	s2 := &ServiceMock{}
	s1.On("Type").Return(ServiceType("s1"))
	s1.On("Dependencies").Return([]ServiceType{})

	s2.On("Type").Return(ServiceType("s2"))
	s2.On("Dependencies").Return([]ServiceType{})

	Register(s1)
	Register(s2)
	s.Equal(services, map[ServiceType]Service{
		ServiceType("s1"): s1,
		ServiceType("s2"): s2,
	})
	// register duplicated service
	s.Assert().Panics(func() {
		Register(s1)
	})
}

func (s *ServiceTestSuite) TestRegisterCyclic() {
	s1 := &ServiceMock{}
	s2 := &ServiceMock{}
	s3 := &ServiceMock{}
	s4 := &ServiceMock{}
	s5 := &ServiceMock{}
	s1.On("Type").Return(ServiceType("s1"))
	s1.On("Dependencies").Return([]ServiceType{"s2", "s3"})

	s2.On("Type").Return(ServiceType("s2"))
	s2.On("Dependencies").Return([]ServiceType{"s3"})

	s3.On("Type").Return(ServiceType("s3"))
	s3.On("Dependencies").Return([]ServiceType{"s1"})

	s4.On("Type").Return(ServiceType("s4"))
	s4.On("Dependencies").Return([]ServiceType{"s2"})

	s5.On("Type").Return(ServiceType("s5"))
	s5.On("Dependencies").Return([]ServiceType{"s5"})

	Register(s1)
	Register(s2)
	// register duplicated service
	s.Assert().Panics(func() {
		Register(s3)
	})
	Register(s4)
	// depend on itself
	s.Assert().Panics(func() {
		Register(s5)
	})
}
