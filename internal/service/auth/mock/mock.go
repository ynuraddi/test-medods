// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/service/auth/auth.go

// Package mock_auth is a generated GoMock package.
package mock_auth

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// CreateSession mocks base method.
func (m *MockInterface) CreateSession(ctx context.Context, uid int, ip string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSession", ctx, uid, ip)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateSession indicates an expected call of CreateSession.
func (mr *MockInterfaceMockRecorder) CreateSession(ctx, uid, ip interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSession", reflect.TypeOf((*MockInterface)(nil).CreateSession), ctx, uid, ip)
}

// RefreshSession mocks base method.
func (m *MockInterface) RefreshSession(ctx context.Context, aT, rT string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshSession", ctx, aT, rT)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// RefreshSession indicates an expected call of RefreshSession.
func (mr *MockInterfaceMockRecorder) RefreshSession(ctx, aT, rT interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshSession", reflect.TypeOf((*MockInterface)(nil).RefreshSession), ctx, aT, rT)
}
