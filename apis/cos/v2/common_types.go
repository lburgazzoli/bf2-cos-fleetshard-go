package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Condition ---
type Condition struct {
	metav1.Condition `json:",inline"`
	ResourceRevision int64 `json:"resourceRevision,omitempty"`
}
