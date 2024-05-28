package controller

import (
	"bytes"
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/apis/application/v1alpha1"
	appclientset "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned/fake"
	appinformers "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/informers/externalversions"
	"github.com/minhthong582000/k8s-controller-pattern/gitops/utils/git"
	k8sutil "github.com/minhthong582000/k8s-controller-pattern/gitops/utils/kube"
	"github.com/stretchr/testify/assert"
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

func newFakeController(apps ...runtime.Object) *Controller {
	kubeClientSet := fake.NewSimpleClientset()
	appClientSet := appclientset.NewSimpleClientset(apps...)
	gitClient := git.NewGitClient("")
	k8sUtil := k8sutil.NewK8s(nil, nil)
	appInformerFactory := appinformers.NewSharedInformerFactory(appClientSet, time.Second*30)

	return NewController(
		kubeClientSet,
		appClientSet,
		appInformerFactory.Thongdepzai().V1alpha1().Applications(),
		gitClient,
		k8sUtil,
	)
}

var (
	createResourcesTestCases = []struct {
		name           string
		app            string
		expectedOut    string
		expectedStatus v1alpha1.HealthStatusCode
		expectedErr    string
	}{
		{
			name: "Normal application 1",
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
			expectedStatus: v1alpha1.HealthStatusCode(v1alpha1.HealthStatusHealthy),
		},
		{
			name: "Normal application 2",
			app: `
kind: Application
apiVersion: thongdepzai.cloud/v1alpha1
metadata:
  name: test-example-application-two
spec:
  repository: https://github.com/minhthong582000/k8s-controller-pattern.git
  revision: main
  path: k8s-controller-pattern/gitops
`,
			expectedStatus: v1alpha1.HealthStatusCode(v1alpha1.HealthStatusHealthy),
		},
		{
			name: "Application with invalid repository",
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
			expectedStatus: v1alpha1.HealthStatusCode(v1alpha1.HealthStatusProgressing),
			expectedErr:    "error cloning repository: failed to clone repository: authentication required",
		},
	}
)

func Test_CreateResources(t *testing.T) {
	for _, tt := range createResourcesTestCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			app := newFakeApp(tt.app)
			controller := newFakeController(app)

			err := controller.createResources(ctx, app)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				repoPath := path.Join(os.TempDir(), app.Name, strings.Replace(app.Spec.Repository, "/", "_", -1))

				// Check if git repository is cloned
				assert.DirExists(t, repoPath)

				// TODO: check if the revision is checked out

				// Delete the git repository
				err = os.RemoveAll(repoPath)
				assert.NoError(t, err)
				assert.NoDirExists(t, repoPath)
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

var (
	deleteResourcesTestCases = []struct {
		name           string
		app            string
		expectedOut    string
		expectedStatus string
		expectedErr    string
	}{
		{
			name: "Normal application",
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
		},
		{
			name: "Application with invalid repository",
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
		},
	}
)

func Test_DeleteResources(t *testing.T) {
	for _, tt := range deleteResourcesTestCases {
		t.Run(tt.name, func(t *testing.T) {
			app := newFakeApp(tt.app)
			controller := newFakeController(app)

			// Create a fake git repository
			repoPath := path.Join(os.TempDir(), app.Name, strings.Replace(app.Spec.Repository, "/", "_", -1))
			err := os.MkdirAll(repoPath, os.ModePerm)
			assert.NoError(t, err)
			assert.DirExists(t, repoPath)

			// Delete resources
			err = controller.deleteResources(app)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
			}

			// Check if the git repository is deleted
			assert.NoDirExists(t, repoPath)
		})
	}
}
