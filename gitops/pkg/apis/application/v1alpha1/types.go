package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="HealthStatus",type=string,JSONPath=`.status.healthStatus`
// +kubebuilder:printcolumn:name="LastSync",type=string,JSONPath=`.status.lastSyncAt`
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {
	Repository string `json:"repository,omitempty"`
	Revision   string `json:"revision,omitempty"`
	Path       string `json:"path,omitempty"`
}

type ApplicationStatus struct {
	HealthStatus HealthStatusCode `json:"healthStatus,omitempty"`
	Revision     string           `json:"revision,omitempty"`
	LastSyncAt   metav1.Time      `json:"lastSyncAt,omitempty"`
}

type HealthStatusCode string

const (
	HealthStatusProgressing = "Progressing"
	HealthStatusHealthy     = "Healthy"
	HealthStatusDegraded    = "Degraded"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Application `json:"items"`
}
