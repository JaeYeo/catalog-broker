// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import internal "github.com/kyma-project/helm-broker/internal"
import mock "github.com/stretchr/testify/mock"

// instanceBindDataRemover is an autogenerated mock type for the instanceBindDataRemover type
type instanceBindDataRemover struct {
	mock.Mock
}

// Remove provides a mock function with given fields: _a0
func (_m *instanceBindDataRemover) Remove(_a0 internal.InstanceID) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(internal.InstanceID) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}