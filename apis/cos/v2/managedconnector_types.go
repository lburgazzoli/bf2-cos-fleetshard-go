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

// ManagedConnectorSpec defines the desired state of ManagedConnector
type ManagedConnectorSpec struct {
}

// ManagedConnectorStatus defines the observed state of ManagedConnector
type ManagedConnectorStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
