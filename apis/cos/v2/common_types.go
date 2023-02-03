package v2

import (
	"errors"
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
	ClientID corev1.EnvVarSource `json:"clientId"`
	// +required
	// +kubebuilder:validation:Required
	ClientSecret corev1.EnvVarSource `json:"clientSecret"`
}

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
// +kubebuilder:validation:Type=object
// +kubebuilder:validation:Format=""
// +kubebuilder:pruning:PreserveUnknownFields
type RawMessage []byte

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}
