package v2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// StatusSpec ---
type StatusSpec struct {
	Phase      string             `json:"deployment"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
