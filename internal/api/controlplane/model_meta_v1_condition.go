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

// MetaV1Condition struct for MetaV1Condition
type MetaV1Condition struct {
	Type               *string `json:"type,omitempty"`
	Reason             *string `json:"reason,omitempty"`
	Message            *string `json:"message,omitempty"`
	Status             *string `json:"status,omitempty"`
	LastTransitionTime *string `json:"last_transition_time,omitempty"`
}

// NewMetaV1Condition instantiates a new MetaV1Condition object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewMetaV1Condition() *MetaV1Condition {
	this := MetaV1Condition{}
	return &this
}

// NewMetaV1ConditionWithDefaults instantiates a new MetaV1Condition object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewMetaV1ConditionWithDefaults() *MetaV1Condition {
	this := MetaV1Condition{}
	return &this
}

// GetType returns the Type field value if set, zero value otherwise.
func (o *MetaV1Condition) GetType() string {
	if o == nil || isNil(o.Type) {
		var ret string
		return ret
	}
	return *o.Type
}

// GetTypeOk returns a tuple with the Type field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetaV1Condition) GetTypeOk() (*string, bool) {
	if o == nil || isNil(o.Type) {
		return nil, false
	}
	return o.Type, true
}

// HasType returns a boolean if a field has been set.
func (o *MetaV1Condition) HasType() bool {
	if o != nil && !isNil(o.Type) {
		return true
	}

	return false
}

// SetType gets a reference to the given string and assigns it to the Type field.
func (o *MetaV1Condition) SetType(v string) {
	o.Type = &v
}

// GetReason returns the Reason field value if set, zero value otherwise.
func (o *MetaV1Condition) GetReason() string {
	if o == nil || isNil(o.Reason) {
		var ret string
		return ret
	}
	return *o.Reason
}

// GetReasonOk returns a tuple with the Reason field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetaV1Condition) GetReasonOk() (*string, bool) {
	if o == nil || isNil(o.Reason) {
		return nil, false
	}
	return o.Reason, true
}

// HasReason returns a boolean if a field has been set.
func (o *MetaV1Condition) HasReason() bool {
	if o != nil && !isNil(o.Reason) {
		return true
	}

	return false
}

// SetReason gets a reference to the given string and assigns it to the Reason field.
func (o *MetaV1Condition) SetReason(v string) {
	o.Reason = &v
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *MetaV1Condition) GetMessage() string {
	if o == nil || isNil(o.Message) {
		var ret string
		return ret
	}
	return *o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetaV1Condition) GetMessageOk() (*string, bool) {
	if o == nil || isNil(o.Message) {
		return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *MetaV1Condition) HasMessage() bool {
	if o != nil && !isNil(o.Message) {
		return true
	}

	return false
}

// SetMessage gets a reference to the given string and assigns it to the Message field.
func (o *MetaV1Condition) SetMessage(v string) {
	o.Message = &v
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *MetaV1Condition) GetStatus() string {
	if o == nil || isNil(o.Status) {
		var ret string
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetaV1Condition) GetStatusOk() (*string, bool) {
	if o == nil || isNil(o.Status) {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *MetaV1Condition) HasStatus() bool {
	if o != nil && !isNil(o.Status) {
		return true
	}

	return false
}

// SetStatus gets a reference to the given string and assigns it to the Status field.
func (o *MetaV1Condition) SetStatus(v string) {
	o.Status = &v
}

// GetLastTransitionTime returns the LastTransitionTime field value if set, zero value otherwise.
func (o *MetaV1Condition) GetLastTransitionTime() string {
	if o == nil || isNil(o.LastTransitionTime) {
		var ret string
		return ret
	}
	return *o.LastTransitionTime
}

// GetLastTransitionTimeOk returns a tuple with the LastTransitionTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *MetaV1Condition) GetLastTransitionTimeOk() (*string, bool) {
	if o == nil || isNil(o.LastTransitionTime) {
		return nil, false
	}
	return o.LastTransitionTime, true
}

// HasLastTransitionTime returns a boolean if a field has been set.
func (o *MetaV1Condition) HasLastTransitionTime() bool {
	if o != nil && !isNil(o.LastTransitionTime) {
		return true
	}

	return false
}

// SetLastTransitionTime gets a reference to the given string and assigns it to the LastTransitionTime field.
func (o *MetaV1Condition) SetLastTransitionTime(v string) {
	o.LastTransitionTime = &v
}

func (o MetaV1Condition) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Type) {
		toSerialize["type"] = o.Type
	}
	if !isNil(o.Reason) {
		toSerialize["reason"] = o.Reason
	}
	if !isNil(o.Message) {
		toSerialize["message"] = o.Message
	}
	if !isNil(o.Status) {
		toSerialize["status"] = o.Status
	}
	if !isNil(o.LastTransitionTime) {
		toSerialize["last_transition_time"] = o.LastTransitionTime
	}
	return json.Marshal(toSerialize)
}

type NullableMetaV1Condition struct {
	value *MetaV1Condition
	isSet bool
}

func (v NullableMetaV1Condition) Get() *MetaV1Condition {
	return v.value
}

func (v *NullableMetaV1Condition) Set(val *MetaV1Condition) {
	v.value = val
	v.isSet = true
}

func (v NullableMetaV1Condition) IsSet() bool {
	return v.isSet
}

func (v *NullableMetaV1Condition) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableMetaV1Condition(val *MetaV1Condition) *NullableMetaV1Condition {
	return &NullableMetaV1Condition{value: val, isSet: true}
}

func (v NullableMetaV1Condition) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableMetaV1Condition) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}