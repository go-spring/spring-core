// Code generated by MockGen. DO NOT EDIT.
// Source: cond.go
//
// Generated by this command:
//
//	mockgen -build_flags="-mod=mod" -package=cond -source=cond.go -destination=cond_mock.go
//

// Package cond is a generated GoMock package.
package gs_cond

import (
	"reflect"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"go.uber.org/mock/gomock"
)

// MockContext is a mock of Context interface.
type MockContext struct {
	ctrl     *gomock.Controller
	recorder *MockContextMockRecorder
	isgomock struct{}
}

// MockContextMockRecorder is the mock recorder for MockContext.
type MockContextMockRecorder struct {
	mock *MockContext
}

// NewMockContext creates a new mock instance.
func NewMockContext(ctrl *gomock.Controller) *MockContext {
	mock := &MockContext{ctrl: ctrl}
	mock.recorder = &MockContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContext) EXPECT() *MockContextMockRecorder {
	return m.recorder
}

// Find mocks base method.
func (m *MockContext) Find(selector gs.BeanSelector) ([]gs.BeanDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Find", selector)
	ret0, _ := ret[0].([]gs.BeanDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Find indicates an expected call of Find.
func (mr *MockContextMockRecorder) Find(selector any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockContext)(nil).Find), selector)
}

// Has mocks base method.
func (m *MockContext) Has(key string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Has", key)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Has indicates an expected call of Has.
func (mr *MockContextMockRecorder) Has(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Has", reflect.TypeOf((*MockContext)(nil).Has), key)
}

// Prop mocks base method.
func (m *MockContext) Prop(key string, opts ...conf.GetOption) string {
	m.ctrl.T.Helper()
	varargs := []any{key}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Prop", varargs...)
	ret0, _ := ret[0].(string)
	return ret0
}

// Prop indicates an expected call of Prop.
func (mr *MockContextMockRecorder) Prop(key any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{key}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prop", reflect.TypeOf((*MockContext)(nil).Prop), varargs...)
}

// MockCondition is a mock of Condition interface.
type MockCondition struct {
	ctrl     *gomock.Controller
	recorder *MockConditionMockRecorder
	isgomock struct{}
}

// MockConditionMockRecorder is the mock recorder for MockCondition.
type MockConditionMockRecorder struct {
	mock *MockCondition
}

// NewMockCondition creates a new mock instance.
func NewMockCondition(ctrl *gomock.Controller) *MockCondition {
	mock := &MockCondition{ctrl: ctrl}
	mock.recorder = &MockConditionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCondition) EXPECT() *MockConditionMockRecorder {
	return m.recorder
}

// Matches mocks base method.
func (m *MockCondition) Matches(ctx gs.ConditionContext) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Matches", ctx)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Matches indicates an expected call of Matches.
func (mr *MockConditionMockRecorder) Matches(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Matches", reflect.TypeOf((*MockCondition)(nil).Matches), ctx)
}
