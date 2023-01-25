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
	SchemeBuilder.Register(&ManagedConnectorCluster{}, &ManagedConnectorClusterList{})
}

// ManagedConnectorClusterSpec ---
type ManagedConnectorClusterSpec struct {
	// +required
	// +kubebuilder:validation:Required
	ControlPlaneURL string `json:"controlPlaneURL"`
	// +required
	// +kubebuilder:validation:Required
	Auth AuthSpec `json:"auth"`
}

// ManagedConnectorClusterStatus ---
type ManagedConnectorClusterStatus struct {
	Phase              string             `json:"phase"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ManagedConnectorCluster is the Schema for the managedconnectorclusters API
type ManagedConnectorCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedConnectorClusterSpec   `json:"spec,omitempty"`
	Status ManagedConnectorClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManagedConnectorClusterList contains a list of ManagedConnectorCluster
type ManagedConnectorClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedConnectorCluster `json:"items"`
}
