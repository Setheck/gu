// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Producer is an autogenerated mock type for the Producer type
type Producer struct {
	mock.Mock
}

// Produce provides a mock function with given fields: message
func (_m *Producer) Produce(message []byte) error {
	ret := _m.Called(message)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ProducerHealth provides a mock function with given fields:
func (_m *Producer) ProducerHealth() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
