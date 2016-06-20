package mocks

import "github.com/stretchr/testify/mock"

// Fluent is an autogenerated mock type for the Fluent type
type Fluent struct {
	mock.Mock
}

// Post provides a mock function with given fields: tag, v
func (_m *Fluent) Post(tag string, v interface{}) {
	_m.Called(tag, v)
}
