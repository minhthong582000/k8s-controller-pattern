package k8s

import (
	"path/filepath"
	"testing"

	"github.com/minhthong582000/k8s-controller-pattern/gitops/common"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynclientfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_GenerateManifests(t *testing.T) {
	var testCases = []struct {
		name        string
		testPath    string
		expectedErr string
	}{
		{
			name:        "Should generate manifests",
			testPath:    filepath.Join(".", "testdata"),
			expectedErr: "",
		},
		{
			name:        "Should return error when the path is invalid",
			testPath:    "invalid-path",
			expectedErr: "lstat invalid-path: no such file or directory",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			k8sUtil := NewK8s(nil, nil)
			objs, err := k8sUtil.GenerateManifests(tt.testPath)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}

			assert.NotEmpty(t, objs)
			assert.Len(t, objs, 4) // 4 is the number of resources in the /testdata directory
			for _, obj := range objs {
				assert.NotEmpty(t, obj)
			}
		})
	}
}

func Test_DiffResources(t *testing.T) {
	var testCases = []struct {
		name           string
		current        []*unstructured.Unstructured
		new            []*unstructured.Unstructured
		expectedResult bool
	}{
		{
			name: "Should return false when there are no differences between the resources",
			current: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
					},
				},
			},
			new: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "Should return true when there are differences between the resources",
			current: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.26",
								},
							},
						},
					},
				},
			},
			new: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:latest",
								},
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when resources exist in the current but not in the new",
			current: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.26",
								},
							},
						},
					},
				},
			},
			new: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "apache",
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "apache",
									"image": "httpd:1.1",
								},
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when resources exist in the current but new is empty",
			current: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.26",
								},
							},
						},
					},
				},
			},
			new:            []*unstructured.Unstructured{},
			expectedResult: true,
		},
		{
			name:    "Should return true when resources exist in new but not in the current",
			current: []*unstructured.Unstructured{},
			new: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.26",
								},
							},
						},
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			k8sUtil := NewK8s(nil, nil)
			result, err := k8sUtil.DiffResources(tt.current, tt.new)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func Test_SetLabelsForResources(t *testing.T) {
	var testCases = []struct {
		name           string
		resources      []*unstructured.Unstructured
		labels         map[string]string
		expectedOutput []*unstructured.Unstructured
		expectedErr    string
	}{
		{
			name: "Should set labels for resources",
			resources: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
						},
					},
				},
			},
			labels: map[string]string{
				"app":                      "nginx",
				common.LabelKeyAppInstance: "example-app",
			},
			expectedOutput: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name": "nginx",
							"labels": map[string]interface{}{
								"app":                      "nginx",
								common.LabelKeyAppInstance: "example-app",
							},
						},
					},
				},
			},
			expectedErr: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			k8sUtil := NewK8s(nil, nil)
			err := k8sUtil.SetLabelsForResources(tt.resources, tt.labels)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}

			assert.Equal(t, tt.expectedOutput, tt.resources)
		})
	}
}

func Test_GetResourceWithLabel(t *testing.T) {
	var testCases = []struct {
		name              string
		label             map[string]string
		expectedResources []*unstructured.Unstructured
		expectedErr       string
	}{
		{
			name:        "Should return error when the label is empty",
			label:       map[string]string{},
			expectedErr: "label is empty",
		},
		{
			name:        "Should return error when the label is missing",
			label:       nil,
			expectedErr: "label is empty",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			clientSet := fake.NewSimpleClientset()
			discoveryClient := clientSet.Discovery()
			dynClientSet := dynclientfake.NewSimpleDynamicClient(runtime.NewScheme())
			k8sUtil := NewK8s(discoveryClient, dynClientSet)

			resources, err := k8sUtil.GetResourceWithLabel(tt.label)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}

			assert.Equal(t, tt.expectedResources, resources)
		})
	}
}
