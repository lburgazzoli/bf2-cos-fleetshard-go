/*
Connector Service Fleet Manager Private APIs

Connector Service Fleet Manager apis that are used by internal services.

API version: 0.0.3
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package controlplane

import (
	"encoding/json"
)

// ConnectorDeploymentStatus The status of connector deployment
type ConnectorDeploymentStatus struct {
	Phase           *ConnectorState                     `json:"phase,omitempty"`
	ResourceVersion *int64                              `json:"resource_version,omitempty"`
	Operators       *ConnectorDeploymentStatusOperators `json:"operators,omitempty"`
	Conditions      []MetaV1Condition                   `json:"conditions,omitempty"`
}

// NewConnectorDeploymentStatus instantiates a new ConnectorDeploymentStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConnectorDeploymentStatus() *ConnectorDeploymentStatus {
	this := ConnectorDeploymentStatus{}
	return &this
}

// NewConnectorDeploymentStatusWithDefaults instantiates a new ConnectorDeploymentStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConnectorDeploymentStatusWithDefaults() *ConnectorDeploymentStatus {
	this := ConnectorDeploymentStatus{}
	return &this
}

// GetPhase returns the Phase field value if set, zero value otherwise.
func (o *ConnectorDeploymentStatus) GetPhase() ConnectorState {
	if o == nil || isNil(o.Phase) {
		var ret ConnectorState
		return ret
	}
	return *o.Phase
}

// GetPhaseOk returns a tuple with the Phase field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentStatus) GetPhaseOk() (*ConnectorState, bool) {
	if o == nil || isNil(o.Phase) {
		return nil, false
	}
	return o.Phase, true
}

// HasPhase returns a boolean if a field has been set.
func (o *ConnectorDeploymentStatus) HasPhase() bool {
	if o != nil && !isNil(o.Phase) {
		return true
	}

	return false
}

// SetPhase gets a reference to the given ConnectorState and assigns it to the Phase field.
func (o *ConnectorDeploymentStatus) SetPhase(v ConnectorState) {
	o.Phase = &v
}

// GetResourceVersion returns the ResourceVersion field value if set, zero value otherwise.
func (o *ConnectorDeploymentStatus) GetResourceVersion() int64 {
	if o == nil || isNil(o.ResourceVersion) {
		var ret int64
		return ret
	}
	return *o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentStatus) GetResourceVersionOk() (*int64, bool) {
	if o == nil || isNil(o.ResourceVersion) {
		return nil, false
	}
	return o.ResourceVersion, true
}

// HasResourceVersion returns a boolean if a field has been set.
func (o *ConnectorDeploymentStatus) HasResourceVersion() bool {
	if o != nil && !isNil(o.ResourceVersion) {
		return true
	}

	return false
}

// SetResourceVersion gets a reference to the given int64 and assigns it to the ResourceVersion field.
func (o *ConnectorDeploymentStatus) SetResourceVersion(v int64) {
	o.ResourceVersion = &v
}

// GetOperators returns the Operators field value if set, zero value otherwise.
func (o *ConnectorDeploymentStatus) GetOperators() ConnectorDeploymentStatusOperators {
	if o == nil || isNil(o.Operators) {
		var ret ConnectorDeploymentStatusOperators
		return ret
	}
	return *o.Operators
}

// GetOperatorsOk returns a tuple with the Operators field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentStatus) GetOperatorsOk() (*ConnectorDeploymentStatusOperators, bool) {
	if o == nil || isNil(o.Operators) {
		return nil, false
	}
	return o.Operators, true
}

// HasOperators returns a boolean if a field has been set.
func (o *ConnectorDeploymentStatus) HasOperators() bool {
	if o != nil && !isNil(o.Operators) {
		return true
	}

	return false
}

// SetOperators gets a reference to the given ConnectorDeploymentStatusOperators and assigns it to the Operators field.
func (o *ConnectorDeploymentStatus) SetOperators(v ConnectorDeploymentStatusOperators) {
	o.Operators = &v
}

// GetConditions returns the Conditions field value if set, zero value otherwise.
func (o *ConnectorDeploymentStatus) GetConditions() []MetaV1Condition {
	if o == nil || isNil(o.Conditions) {
		var ret []MetaV1Condition
		return ret
	}
	return o.Conditions
}

// GetConditionsOk returns a tuple with the Conditions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentStatus) GetConditionsOk() ([]MetaV1Condition, bool) {
	if o == nil || isNil(o.Conditions) {
		return nil, false
	}
	return o.Conditions, true
}

// HasConditions returns a boolean if a field has been set.
func (o *ConnectorDeploymentStatus) HasConditions() bool {
	if o != nil && !isNil(o.Conditions) {
		return true
	}

	return false
}

// SetConditions gets a reference to the given []MetaV1Condition and assigns it to the Conditions field.
func (o *ConnectorDeploymentStatus) SetConditions(v []MetaV1Condition) {
	o.Conditions = v
}

func (o ConnectorDeploymentStatus) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Phase) {
		toSerialize["phase"] = o.Phase
	}
	if !isNil(o.ResourceVersion) {
		toSerialize["resource_version"] = o.ResourceVersion
	}
	if !isNil(o.Operators) {
		toSerialize["operators"] = o.Operators
	}
	if !isNil(o.Conditions) {
		toSerialize["conditions"] = o.Conditions
	}
	return json.Marshal(toSerialize)
}

type NullableConnectorDeploymentStatus struct {
	value *ConnectorDeploymentStatus
	isSet bool
}

func (v NullableConnectorDeploymentStatus) Get() *ConnectorDeploymentStatus {
	return v.value
}

func (v *NullableConnectorDeploymentStatus) Set(val *ConnectorDeploymentStatus) {
	v.value = val
	v.isSet = true
}

func (v NullableConnectorDeploymentStatus) IsSet() bool {
	return v.isSet
}

func (v *NullableConnectorDeploymentStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConnectorDeploymentStatus(val *ConnectorDeploymentStatus) *NullableConnectorDeploymentStatus {
	return &NullableConnectorDeploymentStatus{value: val, isSet: true}
}

func (v NullableConnectorDeploymentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConnectorDeploymentStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
