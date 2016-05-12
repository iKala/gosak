package mocks

import "straas.io/sauron"
import "github.com/stretchr/testify/mock"

// Engine is an autogenerated mock type for the Engine type
type Engine struct {
	mock.Mock
}

// AddPlugin provides a mock function with given fields: p
func (_m *Engine) AddPlugin(p sauron.Plugin) error {
	ret := _m.Called(p)

	var r0 error
	if rf, ok := ret.Get(0).(func(sauron.Plugin) error); ok {
		r0 = rf(p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Run provides a mock function with given fields:
func (_m *Engine) Run() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetJobMeta provides a mock function with given fields: meta
func (_m *Engine) SetJobMeta(meta sauron.JobMeta) error {
	ret := _m.Called(meta)

	var r0 error
	if rf, ok := ret.Get(0).(func(sauron.JobMeta) error); ok {
		r0 = rf(meta)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
