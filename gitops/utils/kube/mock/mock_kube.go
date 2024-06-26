// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/minhthong582000/k8s-controller-pattern/gitops/utils/kube (interfaces: K8s)
//
// Generated by this command:
//
//	mockgen -destination=mock_kube.go -package=mock github.com/minhthong582000/k8s-controller-pattern/gitops/utils/kube K8s
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MockK8s is a mock of K8s interface.
type MockK8s struct {
	ctrl     *gomock.Controller
	recorder *MockK8sMockRecorder
}

// MockK8sMockRecorder is the mock recorder for MockK8s.
type MockK8sMockRecorder struct {
	mock *MockK8s
}

// NewMockK8s creates a new mock instance.
func NewMockK8s(ctrl *gomock.Controller) *MockK8s {
	mock := &MockK8s{ctrl: ctrl}
	mock.recorder = &MockK8sMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockK8s) EXPECT() *MockK8sMockRecorder {
	return m.recorder
}

// CreateResource mocks base method.
func (m *MockK8s) CreateResource(arg0 context.Context, arg1 *unstructured.Unstructured, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateResource", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateResource indicates an expected call of CreateResource.
func (mr *MockK8sMockRecorder) CreateResource(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResource", reflect.TypeOf((*MockK8s)(nil).CreateResource), arg0, arg1, arg2)
}

// DeleteResource mocks base method.
func (m *MockK8s) DeleteResource(arg0 context.Context, arg1 *unstructured.Unstructured, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteResource", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteResource indicates an expected call of DeleteResource.
func (mr *MockK8sMockRecorder) DeleteResource(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteResource", reflect.TypeOf((*MockK8s)(nil).DeleteResource), arg0, arg1, arg2)
}

// DiffResources mocks base method.
func (m *MockK8s) DiffResources(arg0, arg1 []*unstructured.Unstructured) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DiffResources", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DiffResources indicates an expected call of DiffResources.
func (mr *MockK8sMockRecorder) DiffResources(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DiffResources", reflect.TypeOf((*MockK8s)(nil).DiffResources), arg0, arg1)
}

// GenerateManifests mocks base method.
func (m *MockK8s) GenerateManifests(arg0 string) ([]*unstructured.Unstructured, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateManifests", arg0)
	ret0, _ := ret[0].([]*unstructured.Unstructured)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateManifests indicates an expected call of GenerateManifests.
func (mr *MockK8sMockRecorder) GenerateManifests(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateManifests", reflect.TypeOf((*MockK8s)(nil).GenerateManifests), arg0)
}

// GetResourceWithLabel mocks base method.
func (m *MockK8s) GetResourceWithLabel(arg0 map[string]string) ([]*unstructured.Unstructured, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResourceWithLabel", arg0)
	ret0, _ := ret[0].([]*unstructured.Unstructured)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResourceWithLabel indicates an expected call of GetResourceWithLabel.
func (mr *MockK8sMockRecorder) GetResourceWithLabel(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResourceWithLabel", reflect.TypeOf((*MockK8s)(nil).GetResourceWithLabel), arg0)
}

// PatchResource mocks base method.
func (m *MockK8s) PatchResource(arg0 context.Context, arg1 *unstructured.Unstructured, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PatchResource", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// PatchResource indicates an expected call of PatchResource.
func (mr *MockK8sMockRecorder) PatchResource(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchResource", reflect.TypeOf((*MockK8s)(nil).PatchResource), arg0, arg1, arg2)
}

// SetLabelsForResources mocks base method.
func (m *MockK8s) SetLabelsForResources(arg0 []*unstructured.Unstructured, arg1 map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetLabelsForResources", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetLabelsForResources indicates an expected call of SetLabelsForResources.
func (mr *MockK8sMockRecorder) SetLabelsForResources(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLabelsForResources", reflect.TypeOf((*MockK8s)(nil).SetLabelsForResources), arg0, arg1)
}
