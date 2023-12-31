// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cihub/seelog (interfaces: CustomReceiver)

// Package mock_seelog is a generated GoMock package.
package mock_seelog

import (
	reflect "reflect"

	seelog "github.com/cihub/seelog"
	gomock "github.com/golang/mock/gomock"
)

// MockCustomReceiver is a mock of CustomReceiver interface.
type MockCustomReceiver struct {
	ctrl     *gomock.Controller
	recorder *MockCustomReceiverMockRecorder
}

// MockCustomReceiverMockRecorder is the mock recorder for MockCustomReceiver.
type MockCustomReceiverMockRecorder struct {
	mock *MockCustomReceiver
}

// NewMockCustomReceiver creates a new mock instance.
func NewMockCustomReceiver(ctrl *gomock.Controller) *MockCustomReceiver {
	mock := &MockCustomReceiver{ctrl: ctrl}
	mock.recorder = &MockCustomReceiverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCustomReceiver) EXPECT() *MockCustomReceiverMockRecorder {
	return m.recorder
}

// AfterParse mocks base method.
func (m *MockCustomReceiver) AfterParse(arg0 seelog.CustomReceiverInitArgs) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterParse", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// AfterParse indicates an expected call of AfterParse.
func (mr *MockCustomReceiverMockRecorder) AfterParse(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterParse", reflect.TypeOf((*MockCustomReceiver)(nil).AfterParse), arg0)
}

// Close mocks base method.
func (m *MockCustomReceiver) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockCustomReceiverMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockCustomReceiver)(nil).Close))
}

// Flush mocks base method.
func (m *MockCustomReceiver) Flush() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Flush")
}

// Flush indicates an expected call of Flush.
func (mr *MockCustomReceiverMockRecorder) Flush() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockCustomReceiver)(nil).Flush))
}

// ReceiveMessage mocks base method.
func (m *MockCustomReceiver) ReceiveMessage(arg0 string, arg1 seelog.LogLevel, arg2 seelog.LogContextInterface) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReceiveMessage", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReceiveMessage indicates an expected call of ReceiveMessage.
func (mr *MockCustomReceiverMockRecorder) ReceiveMessage(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReceiveMessage", reflect.TypeOf((*MockCustomReceiver)(nil).ReceiveMessage), arg0, arg1, arg2)
}
