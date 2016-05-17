package mocks

import "straas.io/sauron"
import "github.com/stretchr/testify/mock"

// This is an autogenerated mock type for the Sinker type
type Sinker struct {
	mock.Mock
}

// ConfigFactory provides a mock function with given fields:
func (_m *Sinker) ConfigFactory() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// Sink provides a mock function with given fields: config, severity, recovery, desc
func (_m *Sinker) Sink(config interface{}, severity sauron.Severity, recovery bool, desc string) error {
	ret := _m.Called(config, severity, recovery, desc)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, sauron.Severity, bool, string) error); ok {
		r0 = rf(config, severity, recovery, desc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
