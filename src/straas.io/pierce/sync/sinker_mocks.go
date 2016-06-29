package sync

import "github.com/stretchr/testify/mock"

import "straas.io/pierce"

// sinkerMock is an autogenerated mock type for the sinker type
type sinkerMock struct {
	mock.Mock
}

// Diff provides a mock function with given fields: namespace, index, size
func (_m *sinkerMock) Diff(namespace string, index uint64, size int) ([]Record, error) {
	ret := _m.Called(namespace, index, size)

	var r0 []Record
	if rf, ok := ret.Get(0).(func(string, uint64, int) []Record); ok {
		r0 = rf(namespace, index, size)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Record)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, uint64, int) error); ok {
		r1 = rf(namespace, index, size)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatestVersion provides a mock function with given fields:
func (_m *sinkerMock) LatestVersion() (uint64, error) {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Sink provides a mock function with given fields: roomMeta, data, version
func (_m *sinkerMock) Sink(roomMeta pierce.RoomMeta, data interface{}, version uint64) error {
	ret := _m.Called(roomMeta, data, version)

	var r0 error
	if rf, ok := ret.Get(0).(func(pierce.RoomMeta, interface{}, uint64) error); ok {
		r0 = rf(roomMeta, data, version)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
