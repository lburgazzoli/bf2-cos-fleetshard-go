package v2

import (
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&ManagedConnector{}, &ManagedConnectorList{})
}

type DesiredStateType string

const (
	DesiredStateReady   DesiredStateType = "ready"
	DesiredStateDeleted DesiredStateType = "deleted"
	DesiredStateStopped DesiredStateType = "stopped"
)

// KafkaSpec ---
type KafkaSpec struct {
	ID string `json:"id"`

	// +required
	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// ServiceRegistrySpec ---
type ServiceRegistrySpec struct {
	ID string `json:"id"`

	// +required
	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// DeploymentSpec ---
type DeploymentSpec struct {
	// +required
	// +kubebuilder:validation:Required
	ConnectorTypeID string `json:"connectorTypeId"`

	// +required
	// +kubebuilder:validation:Required
	ConnectorResourceVersion int64 `json:"connectorResourceVersion"`

	// +required
	// +kubebuilder:validation:Required
	DeploymentResourceVersion int64 `json:"deploymentResourceVersion"`

	// +required
	// +kubebuilder:validation:Required
	DesiredState DesiredStateType `json:"desiredState"`

	// +required
	// +kubebuilder:validation:Required
	Kafka KafkaSpec `json:"kafka"`

	ServiceRegistry *ServiceRegistrySpec `json:"serviceRegistry,omitempty"`
}

// ManagedConnectorSpec defines the desired state of ManagedConnector
type ManagedConnectorSpec struct {
	// +required
	// +kubebuilder:validation:Required
	ClusterID string `json:"clusterId"`

	// +required
	// +kubebuilder:validation:Required
	ConnectorID string `json:"connectorId"`

	// +required
	// +kubebuilder:validation:Required
	DeploymentID string `json:"deploymentId"`

	// +required
	// +kubebuilder:validation:Required
	OperatorID string `json:"operatorId"`

	// +required
	// +kubebuilder:validation:Required
	Deployment DeploymentSpec `json:"deployment"`

	// +required
	// +kubebuilder:validation:Required
	Config ConnectorConfig `json:"config"`
}

// ConnectorConfig is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
// +kubebuilder:validation:Type=object
// +kubebuilder:validation:Format=""
// +kubebuilder:pruning:PreserveUnknownFields
type ConnectorConfig []byte

// MarshalJSON returns m as the JSON encoding of m.
func (m ConnectorConfig) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *ConnectorConfig) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

// ManagedConnectorStatus defines the observed state of ManagedConnector
type ManagedConnectorStatus struct {
	Phase              string      `json:"phase"`
	Conditions         []Condition `json:"conditions,omitempty"`
	ObservedGeneration int64       `json:"observedGeneration,omitempty"`
	OperatorID         string      `json:"operatorId"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=managedconnectors,scope=Namespaced,shortName=mctr,categories=cos;mas
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="CLUSTER_ID",type=string,JSONPath=`.spec.clusterId`,description="The Cluster ID"
//+kubebuilder:printcolumn:name="CONNECTOR_ID",type=string,JSONPath=`.spec.connectorId`,description="The Connector ID"
//+kubebuilder:printcolumn:name="DEPLOYMENT_ID",type=string,JSONPath=`.spec.connectorId`,description="The Deployment ID"
//+kubebuilder:printcolumn:name="OPERATOR_ID",type=string,JSONPath=`.spec.operatorId`,description="The Operator ID"
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="Ready"
//+kubebuilder:printcolumn:name="REASON",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`,description="Reason"
//+kubebuilder:printcolumn:name="MESSAGE",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].message`,description="Message"

// ManagedConnector is the Schema for the managedconnectors API
type ManagedConnector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedConnectorSpec   `json:"spec,omitempty"`
	Status ManagedConnectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManagedConnectorList contains a list of ManagedConnector
type ManagedConnectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedConnector `json:"items"`
}
