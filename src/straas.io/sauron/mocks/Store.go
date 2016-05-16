package mocks

import "github.com/stretchr/testify/mock"

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// Get provides a mock function with given fields: ns, key, v
func (_m *Store) Get(ns string, key string, v interface{}) (bool, error) {
	ret := _m.Called(ns, key, v)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, interface{}) bool); ok {
		r0 = rf(ns, key, v)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, interface{}) error); ok {
		r1 = rf(ns, key, v)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Set provides a mock function with given fields: ns, key, v
func (_m *Store) Set(ns string, key string, v interface{}) error {
	ret := _m.Called(ns, key, v)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, interface{}) error); ok {
		r0 = rf(ns, key, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ns, key, v, action
func (_m *Store) Update(ns string, key string, v interface{}, action func(interface{}) (interface{}, error)) error {
	ret := _m.Called(ns, key, v, action)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, interface{}, func(interface{}) (interface{}, error)) error); ok {
		r0 = rf(ns, key, v, action)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
