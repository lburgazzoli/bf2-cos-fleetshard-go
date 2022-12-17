package v2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// StatusSpec ---
type StatusSpec struct {
	Phase      string             `json:"phase"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
