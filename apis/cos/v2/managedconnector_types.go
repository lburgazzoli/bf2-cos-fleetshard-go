/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&ManagedConnector{}, &ManagedConnectorList{})
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
	DesiredState string `json:"desiredState"`
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
}

// ManagedConnectorStatus defines the observed state of ManagedConnector
type ManagedConnectorStatus struct {
	StatusSpec `json:",inline"`

	ObservedGeneration int64          `json:"observedGeneration,omitempty"`
	Deployment         DeploymentSpec `json:"deployment"`
	Connector          StatusSpec     `json:"connector"`
}

//+kubebuilder:object:root=true
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
