/*
Copyright 2024 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by mockery v2.40.1. DO NOT EDIT.

package spanner

import (
	mock "github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
)

// mockRowIterator is an autogenerated mock type for the mockRowIterator type
type mockRowIterator struct {
	mock.Mock
}

// Next provides a mock function with given fields:
func (_m *mockRowIterator) Next() (*Row, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Next")
	}

	var r0 *Row
	var r1 error
	if rf, ok := ret.Get(0).(func() (*Row, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *Row); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Row)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *mockRowIterator) Do(f func(r *Row) error) error {
	defer _m.Stop()
	for {
		row, err := _m.Next()
		switch err {
		case iterator.Done:
			return nil
		case nil:
			if err = f(row); err != nil {
				return err
			}
		default:
			return err
		}
	}
}

// Stop provides a mock function with given fields:
func (_m *mockRowIterator) Stop() {
	_m.Called()
}

// newRowIterator creates a new instance of mockRowIterator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newRowIterator(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockRowIterator {
	mock := &mockRowIterator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
