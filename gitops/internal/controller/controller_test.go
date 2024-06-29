package controller

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/apis/application/v1alpha1"
	appclientset "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned/fake"
	appinformers "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/informers/externalversions"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git"
	gitMock "github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git/mock"
	k8sUtil "github.com/minhthong582000/k8s-controller-pattern/gitops/utils/kube"
	k8sUtilMock "github.com/minhthong582000/k8s-controller-pattern/gitops/utils/kube/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/fake"
)

func newFakeApp(appString string) *v1alpha1.Application {
	var app v1alpha1.Application
	dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(appString)), 1000)
	if err := dec.Decode(&app); err != nil {
		return nil
	}

	return &app
}

func newFakeController(gitClient git.GitClient, k8sUtil k8sUtil.K8s, apps ...runtime.Object) *Controller {
	kubeClientSet := fake.NewSimpleClientset()
	appClientSet := appclientset.NewSimpleClientset(apps...)
	appInformerFactory := appinformers.NewSharedInformerFactory(appClientSet, time.Second*30)

	return NewController(
		kubeClientSet,
		appClientSet,
		appInformerFactory.Thongdepzai().V1alpha1().Applications(),
		gitClient,
		k8sUtil,
	)
}

func Test_CreateResources(t *testing.T) {
	ctrl := gomock.NewController(t)

	createResourcesTestCases := []struct {
		name           string
		app            string
		mockGitClient  git.GitClient
		mockk8sUtil    k8sUtil.K8s
		expectedOut    string
		expectedStatus v1alpha1.HealthStatusCode
		expectedErr    string
	}{
		{
			name: "Should create resources successfully if the repository is valid",
			app: `
kind: Application
apiVersion: thongdepzai.cloud/v1alpha1
metadata:
  name: test-example-application-one
spec:
  repository: https://github.com/minhthong582000/k8s-controller-pattern.git
  revision: main
  path: k8s-controller-pattern/gitops
`,
			mockGitClient: func() git.GitClient {
				mock := gitMock.NewMockGitClient(ctrl)
				mock.EXPECT().CloneOrFetch(gomock.Any(), gomock.Any()).Return(nil)
				mock.EXPECT().Checkout(gomock.Any(), gomock.Any()).Return("randomsha", nil)
				return mock
			}(),
			mockk8sUtil: func() k8sUtil.K8s {
				mock := k8sUtilMock.NewMockK8s(ctrl)
				mock.EXPECT().GenerateManifests(gomock.Any()).Return(nil, nil)
				mock.EXPECT().GetResourceWithLabel(gomock.Any()).Return(nil, nil)
				mock.EXPECT().DiffResources(gomock.Any(), gomock.Any()).Return(true, nil)
				mock.EXPECT().SetLabelsForResources(gomock.Any(), gomock.Any()).Return(nil)
				mock.EXPECT().CreateResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return mock
			}(),
			expectedStatus: v1alpha1.HealthStatusCode(v1alpha1.HealthStatusHealthy),
		},
		{
			name: "Should create resources successfully even if there is no diff between the old and new resources",
			app: `
kind: Application
apiVersion: thongdepzai.cloud/v1alpha1
metadata:
  name: test-example-application-one
spec:
  repository: https://github.com/minhthong582000/k8s-controller-pattern.git
  revision: main
  path: k8s-controller-pattern/gitops
`,
			mockGitClient: func() git.GitClient {
				mock := gitMock.NewMockGitClient(ctrl)
				mock.EXPECT().CloneOrFetch(gomock.Any(), gomock.Any()).Return(nil)
				mock.EXPECT().Checkout(gomock.Any(), gomock.Any()).Return("randomsha", nil)
				return mock
			}(),
			mockk8sUtil: func() k8sUtil.K8s {
				mock := k8sUtilMock.NewMockK8s(ctrl)
				mock.EXPECT().GenerateManifests(gomock.Any()).Return(nil, nil)
				mock.EXPECT().GetResourceWithLabel(gomock.Any()).Return(nil, nil)
				mock.EXPECT().SetLabelsForResources(gomock.Any(), gomock.Any()).Return(nil)
				mock.EXPECT().DiffResources(gomock.Any(), gomock.Any()).Return(false, nil)
				return mock
			}(),
			expectedStatus: v1alpha1.HealthStatusCode(v1alpha1.HealthStatusHealthy),
		},
		{
			name: "Should return error if the application has invalid repository",
			app: `
kind: Application
apiVersion: thongdepzai.cloud/v1alpha1
metadata:
  name: example-application
spec:
  repository: https://github.com/kubernetes/kubernetes-but-not-exist.git
  revision: main
  path: k8s-controller-pattern/gitops
`,
			mockGitClient: func() git.GitClient {
				mock := gitMock.NewMockGitClient(ctrl)
				mock.EXPECT().CloneOrFetch(gomock.Any(), gomock.Any()).Return(
					fmt.Errorf("failed to clone repository: authentication required"),
				)
				return mock
			}(),
			expectedStatus: v1alpha1.HealthStatusCode(v1alpha1.HealthStatusProgressing),
			expectedErr:    "error cloning repository: failed to clone repository: authentication required",
		},
	}

	for _, tt := range createResourcesTestCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			app := newFakeApp(tt.app)
			controller := newFakeController(tt.mockGitClient, tt.mockk8sUtil, app)

			err := controller.createResources(ctx, app)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
			}

			// Check the status of the application
			queryApp, err := controller.appClientSet.ThongdepzaiV1alpha1().Applications(app.Namespace).Get(ctx, app.Name, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, queryApp.Status.HealthStatus)

			// Check the last sync time
			assert.NotNil(t, queryApp.Status.LastSyncAt)
		})
	}
}

func Test_DeleteResources(t *testing.T) {
	ctrl := gomock.NewController(t)

	testCases := []struct {
		name           string
		app            string
		mockGitClient  git.GitClient
		mockk8sUtil    k8sUtil.K8s
		expectedOut    string
		expectedStatus string
		expectedErr    string
	}{
		{
			name: "Should delete resources successfully if the application is valid",
			app: `
kind: Application
apiVersion: thongdepzai.cloud/v1alpha1
metadata:
  name: test-another-example-application-one
spec:
  repository: https://github.com/minhthong582000/k8s-controller-pattern.git
  revision: main
  path: k8s-controller-pattern/gitops
`,
			mockGitClient: func() git.GitClient {
				mock := gitMock.NewMockGitClient(ctrl)
				mock.EXPECT().CleanUp(gomock.Any()).Return(nil)
				return mock
			}(),
			mockk8sUtil: func() k8sUtil.K8s {
				mock := k8sUtilMock.NewMockK8s(ctrl)
				mock.EXPECT().GetResourceWithLabel(gomock.Any()).Return(nil, nil)
				mock.EXPECT().DeleteResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return mock
			}(),
		},
		{
			name: "Should delete resources successfully even if the application has invalid repository",
			app: `
kind: Application
apiVersion: thongdepzai.cloud/v1alpha1
metadata:
  name: test-another-example-application-two
spec:
  repository: https://github.com/kubernetes/kubernetes-but-not-exist.git
  revision: main
  path: k8s-controller-pattern/gitops
`,
			mockGitClient: func() git.GitClient {
				mock := gitMock.NewMockGitClient(ctrl)
				mock.EXPECT().CleanUp(gomock.Any()).Return(nil)
				return mock
			}(),
			mockk8sUtil: func() k8sUtil.K8s {
				mock := k8sUtilMock.NewMockK8s(ctrl)
				mock.EXPECT().GetResourceWithLabel(gomock.Any()).Return(nil, nil)
				mock.EXPECT().DeleteResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return mock
			}(),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := newFakeApp(tt.app)
			controller := newFakeController(tt.mockGitClient, tt.mockk8sUtil, app)

			// Delete resources
			err := controller.deleteResources(app)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
			}
		})
	}
}
