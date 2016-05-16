package mocks

import "straas.io/sauron"
import "github.com/stretchr/testify/mock"

// PluginContext is an autogenerated mock type for the PluginContext type
type PluginContext struct {
	mock.Mock
}

// ArgBoolean provides a mock function with given fields: i
func (_m *PluginContext) ArgBoolean(i int) (bool, error) {
	ret := _m.Called(i)

	var r0 bool
	if rf, ok := ret.Get(0).(func(int) bool); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(i)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ArgFloat provides a mock function with given fields: i
func (_m *PluginContext) ArgFloat(i int) (float64, error) {
	ret := _m.Called(i)

	var r0 float64
	if rf, ok := ret.Get(0).(func(int) float64); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Get(0).(float64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(i)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ArgInt provides a mock function with given fields: i
func (_m *PluginContext) ArgInt(i int) (int64, error) {
	ret := _m.Called(i)

	var r0 int64
	if rf, ok := ret.Get(0).(func(int) int64); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(i)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ArgLen provides a mock function with given fields:
func (_m *PluginContext) ArgLen() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// ArgObject provides a mock function with given fields: i, v
func (_m *PluginContext) ArgObject(i int, v interface{}) error {
	ret := _m.Called(i, v)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, interface{}) error); ok {
		r0 = rf(i, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ArgString provides a mock function with given fields: i
func (_m *PluginContext) ArgString(i int) (string, error) {
	ret := _m.Called(i)

	var r0 string
	if rf, ok := ret.Get(0).(func(int) string); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(i)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CallFunction provides a mock function with given fields: i, args
func (_m *PluginContext) CallFunction(i int, args ...interface{}) (interface{}, error) {
	ret := _m.Called(i, args)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(int, ...interface{}) interface{}); ok {
		r0 = rf(i, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, ...interface{}) error); ok {
		r1 = rf(i, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsCallable provides a mock function with given fields: i
func (_m *PluginContext) IsCallable(i int) bool {
	ret := _m.Called(i)

	var r0 bool
	if rf, ok := ret.Get(0).(func(int) bool); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// JobMeta provides a mock function with given fields:
func (_m *PluginContext) JobMeta() sauron.JobMeta {
	ret := _m.Called()

	var r0 sauron.JobMeta
	if rf, ok := ret.Get(0).(func() sauron.JobMeta); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(sauron.JobMeta)
	}

	return r0
}

// Return provides a mock function with given fields: v
func (_m *PluginContext) Return(v interface{}) error {
	ret := _m.Called(v)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Store provides a mock function with given fields:
func (_m *PluginContext) Store() sauron.Store {
	ret := _m.Called()

	var r0 sauron.Store
	if rf, ok := ret.Get(0).(func() sauron.Store); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sauron.Store)
		}
	}

	return r0
}
