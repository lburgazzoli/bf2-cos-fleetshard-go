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

// ConnectorDeploymentWatchEvent struct for ConnectorDeploymentWatchEvent
type ConnectorDeploymentWatchEvent struct {
	Type   string               `json:"type"`
	Error  NullableError        `json:"error,omitempty"`
	Object *ConnectorDeployment `json:"object,omitempty"`
}

// NewConnectorDeploymentWatchEvent instantiates a new ConnectorDeploymentWatchEvent object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConnectorDeploymentWatchEvent(type_ string) *ConnectorDeploymentWatchEvent {
	this := ConnectorDeploymentWatchEvent{}
	this.Type = type_
	return &this
}

// NewConnectorDeploymentWatchEventWithDefaults instantiates a new ConnectorDeploymentWatchEvent object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConnectorDeploymentWatchEventWithDefaults() *ConnectorDeploymentWatchEvent {
	this := ConnectorDeploymentWatchEvent{}
	return &this
}

// GetType returns the Type field value
func (o *ConnectorDeploymentWatchEvent) GetType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentWatchEvent) GetTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *ConnectorDeploymentWatchEvent) SetType(v string) {
	o.Type = v
}

// GetError returns the Error field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *ConnectorDeploymentWatchEvent) GetError() Error {
	if o == nil || isNil(o.Error.Get()) {
		var ret Error
		return ret
	}
	return *o.Error.Get()
}

// GetErrorOk returns a tuple with the Error field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *ConnectorDeploymentWatchEvent) GetErrorOk() (*Error, bool) {
	if o == nil {
		return nil, false
	}
	return o.Error.Get(), o.Error.IsSet()
}

// HasError returns a boolean if a field has been set.
func (o *ConnectorDeploymentWatchEvent) HasError() bool {
	if o != nil && o.Error.IsSet() {
		return true
	}

	return false
}

// SetError gets a reference to the given NullableError and assigns it to the Error field.
func (o *ConnectorDeploymentWatchEvent) SetError(v Error) {
	o.Error.Set(&v)
}

// SetErrorNil sets the value for Error to be an explicit nil
func (o *ConnectorDeploymentWatchEvent) SetErrorNil() {
	o.Error.Set(nil)
}

// UnsetError ensures that no value is present for Error, not even an explicit nil
func (o *ConnectorDeploymentWatchEvent) UnsetError() {
	o.Error.Unset()
}

// GetObject returns the Object field value if set, zero value otherwise.
func (o *ConnectorDeploymentWatchEvent) GetObject() ConnectorDeployment {
	if o == nil || isNil(o.Object) {
		var ret ConnectorDeployment
		return ret
	}
	return *o.Object
}

// GetObjectOk returns a tuple with the Object field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentWatchEvent) GetObjectOk() (*ConnectorDeployment, bool) {
	if o == nil || isNil(o.Object) {
		return nil, false
	}
	return o.Object, true
}

// HasObject returns a boolean if a field has been set.
func (o *ConnectorDeploymentWatchEvent) HasObject() bool {
	if o != nil && !isNil(o.Object) {
		return true
	}

	return false
}

// SetObject gets a reference to the given ConnectorDeployment and assigns it to the Object field.
func (o *ConnectorDeploymentWatchEvent) SetObject(v ConnectorDeployment) {
	o.Object = &v
}

func (o ConnectorDeploymentWatchEvent) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["type"] = o.Type
	}
	if o.Error.IsSet() {
		toSerialize["error"] = o.Error.Get()
	}
	if !isNil(o.Object) {
		toSerialize["object"] = o.Object
	}
	return json.Marshal(toSerialize)
}

type NullableConnectorDeploymentWatchEvent struct {
	value *ConnectorDeploymentWatchEvent
	isSet bool
}

func (v NullableConnectorDeploymentWatchEvent) Get() *ConnectorDeploymentWatchEvent {
	return v.value
}

func (v *NullableConnectorDeploymentWatchEvent) Set(val *ConnectorDeploymentWatchEvent) {
	v.value = val
	v.isSet = true
}

func (v NullableConnectorDeploymentWatchEvent) IsSet() bool {
	return v.isSet
}

func (v *NullableConnectorDeploymentWatchEvent) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConnectorDeploymentWatchEvent(val *ConnectorDeploymentWatchEvent) *NullableConnectorDeploymentWatchEvent {
	return &NullableConnectorDeploymentWatchEvent{value: val, isSet: true}
}

func (v NullableConnectorDeploymentWatchEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConnectorDeploymentWatchEvent) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
