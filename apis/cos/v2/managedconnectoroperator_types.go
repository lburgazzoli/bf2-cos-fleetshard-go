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
	SchemeBuilder.Register(&ManagedConnectorOperator{}, &ManagedConnectorOperatorList{})
}

// ManagedConnectorOperatorSpec defines the desired state of ManagedConnectorOperator
type ManagedConnectorOperatorSpec struct {
}

// ManagedConnectorOperatorStatus defines the observed state of ManagedConnectorOperator
type ManagedConnectorOperatorStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=managedconnectoroperators,scope=Namespaced,shortName=mco,categories=cos;mas

// ManagedConnectorOperator is the Schema for the managedconnectoroperators API
type ManagedConnectorOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedConnectorOperatorSpec   `json:"spec,omitempty"`
	Status ManagedConnectorOperatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ManagedConnectorOperatorList contains a list of ManagedConnectorOperator
type ManagedConnectorOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedConnectorOperator `json:"items"`
}
