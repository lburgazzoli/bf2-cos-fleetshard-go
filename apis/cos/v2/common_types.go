package v2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Condition ---
type Condition struct {
	metav1.Condition `json:",inline"`
	ResourceRevision int64 `json:"resourceRevision,omitempty"`
}

// AuthSpec ---
type AuthSpec struct {
	// +required
	// +kubebuilder:validation:Required
	AuthURL string `json:"authURL"`
	// +required
	// +kubebuilder:validation:Required
	AuthRealm string `json:"authRealm"`
	// +required
	// +kubebuilder:validation:Required
	ClientID corev1.EnvFromSource `json:"clientId"`
	// +required
	// +kubebuilder:validation:Required
	ClientSecret corev1.EnvFromSource `json:"clientSecret"`
}
