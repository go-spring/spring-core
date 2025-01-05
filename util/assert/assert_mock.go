// Code generated by MockGen. DO NOT EDIT.
// Source: assert.go
//
// Generated by this command:
//
//	mockgen -build_flags="-mod=mod" -package=assert -source=assert.go -destination=assert_mock.go
//

// Package assert is a generated GoMock package.
package assert

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockT is a mock of T interface.
type MockT struct {
	ctrl     *gomock.Controller
	recorder *MockTMockRecorder
	isgomock struct{}
}

// MockTMockRecorder is the mock recorder for MockT.
type MockTMockRecorder struct {
	mock *MockT
}

// NewMockT creates a new mock instance.
func NewMockT(ctrl *gomock.Controller) *MockT {
	mock := &MockT{ctrl: ctrl}
	mock.recorder = &MockTMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockT) EXPECT() *MockTMockRecorder {
	return m.recorder
}

// Error mocks base method.
func (m *MockT) Error(args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Error", varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockTMockRecorder) Error(args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockT)(nil).Error), args...)
}

// Helper mocks base method.
func (m *MockT) Helper() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Helper")
}

// Helper indicates an expected call of Helper.
func (mr *MockTMockRecorder) Helper() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Helper", reflect.TypeOf((*MockT)(nil).Helper))
}
