package mocks

import "straas.io/sauron"
import "github.com/stretchr/testify/mock"

// Config is an autogenerated mock type for the Config type
type Config struct {
	mock.Mock
}

// AddChangeListener provides a mock function with given fields: _a0
func (_m *Config) AddChangeListener(_a0 func()) {
	_m.Called(_a0)
}

// LoadConfig provides a mock function with given fields: path, v
func (_m *Config) LoadConfig(path string, v interface{}) error {
	ret := _m.Called(path, v)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}) error); ok {
		r0 = rf(path, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LoadJobs provides a mock function with given fields: env
func (_m *Config) LoadJobs(env string) ([]sauron.JobMeta, error) {
	ret := _m.Called(env)

	var r0 []sauron.JobMeta
	if rf, ok := ret.Get(0).(func(string) []sauron.JobMeta); ok {
		r0 = rf(env)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sauron.JobMeta)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(env)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
