// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git (interfaces: GitClient)
//
// Generated by this command:
//
//	mockgen -destination=mock_kube.go -package=mock github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git GitClient
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockGitClient is a mock of GitClient interface.
type MockGitClient struct {
	ctrl     *gomock.Controller
	recorder *MockGitClientMockRecorder
}

// MockGitClientMockRecorder is the mock recorder for MockGitClient.
type MockGitClientMockRecorder struct {
	mock *MockGitClient
}

// NewMockGitClient creates a new mock instance.
func NewMockGitClient(ctrl *gomock.Controller) *MockGitClient {
	mock := &MockGitClient{ctrl: ctrl}
	mock.recorder = &MockGitClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitClient) EXPECT() *MockGitClientMockRecorder {
	return m.recorder
}

// Checkout mocks base method.
func (m *MockGitClient) Checkout(arg0, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Checkout", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Checkout indicates an expected call of Checkout.
func (mr *MockGitClientMockRecorder) Checkout(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Checkout", reflect.TypeOf((*MockGitClient)(nil).Checkout), arg0, arg1)
}

// CleanUp mocks base method.
func (m *MockGitClient) CleanUp(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CleanUp", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CleanUp indicates an expected call of CleanUp.
func (mr *MockGitClientMockRecorder) CleanUp(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanUp", reflect.TypeOf((*MockGitClient)(nil).CleanUp), arg0)
}

// CloneOrFetch mocks base method.
func (m *MockGitClient) CloneOrFetch(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloneOrFetch", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CloneOrFetch indicates an expected call of CloneOrFetch.
func (mr *MockGitClientMockRecorder) CloneOrFetch(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloneOrFetch", reflect.TypeOf((*MockGitClient)(nil).CloneOrFetch), arg0, arg1)
}
