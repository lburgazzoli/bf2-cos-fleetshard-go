/*
Connector Service Fleet Manager Private APIs

Connector Service Fleet Manager apis that are used by internal services.

API version: 0.0.3
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package controlplane

import (
	"encoding/json"
	"time"
)

// ObjectMeta struct for ObjectMeta
type ObjectMeta struct {
	Owner      *string    `json:"owner,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	ModifiedAt *time.Time `json:"modified_at,omitempty"`
}

// NewObjectMeta instantiates a new ObjectMeta object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewObjectMeta() *ObjectMeta {
	this := ObjectMeta{}
	return &this
}

// NewObjectMetaWithDefaults instantiates a new ObjectMeta object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewObjectMetaWithDefaults() *ObjectMeta {
	this := ObjectMeta{}
	return &this
}

// GetOwner returns the Owner field value if set, zero value otherwise.
func (o *ObjectMeta) GetOwner() string {
	if o == nil || isNil(o.Owner) {
		var ret string
		return ret
	}
	return *o.Owner
}

// GetOwnerOk returns a tuple with the Owner field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ObjectMeta) GetOwnerOk() (*string, bool) {
	if o == nil || isNil(o.Owner) {
		return nil, false
	}
	return o.Owner, true
}

// HasOwner returns a boolean if a field has been set.
func (o *ObjectMeta) HasOwner() bool {
	if o != nil && !isNil(o.Owner) {
		return true
	}

	return false
}

// SetOwner gets a reference to the given string and assigns it to the Owner field.
func (o *ObjectMeta) SetOwner(v string) {
	o.Owner = &v
}

// GetCreatedAt returns the CreatedAt field value if set, zero value otherwise.
func (o *ObjectMeta) GetCreatedAt() time.Time {
	if o == nil || isNil(o.CreatedAt) {
		var ret time.Time
		return ret
	}
	return *o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ObjectMeta) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil || isNil(o.CreatedAt) {
		return nil, false
	}
	return o.CreatedAt, true
}

// HasCreatedAt returns a boolean if a field has been set.
func (o *ObjectMeta) HasCreatedAt() bool {
	if o != nil && !isNil(o.CreatedAt) {
		return true
	}

	return false
}

// SetCreatedAt gets a reference to the given time.Time and assigns it to the CreatedAt field.
func (o *ObjectMeta) SetCreatedAt(v time.Time) {
	o.CreatedAt = &v
}

// GetModifiedAt returns the ModifiedAt field value if set, zero value otherwise.
func (o *ObjectMeta) GetModifiedAt() time.Time {
	if o == nil || isNil(o.ModifiedAt) {
		var ret time.Time
		return ret
	}
	return *o.ModifiedAt
}

// GetModifiedAtOk returns a tuple with the ModifiedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ObjectMeta) GetModifiedAtOk() (*time.Time, bool) {
	if o == nil || isNil(o.ModifiedAt) {
		return nil, false
	}
	return o.ModifiedAt, true
}

// HasModifiedAt returns a boolean if a field has been set.
func (o *ObjectMeta) HasModifiedAt() bool {
	if o != nil && !isNil(o.ModifiedAt) {
		return true
	}

	return false
}

// SetModifiedAt gets a reference to the given time.Time and assigns it to the ModifiedAt field.
func (o *ObjectMeta) SetModifiedAt(v time.Time) {
	o.ModifiedAt = &v
}

func (o ObjectMeta) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Owner) {
		toSerialize["owner"] = o.Owner
	}
	if !isNil(o.CreatedAt) {
		toSerialize["created_at"] = o.CreatedAt
	}
	if !isNil(o.ModifiedAt) {
		toSerialize["modified_at"] = o.ModifiedAt
	}
	return json.Marshal(toSerialize)
}

type NullableObjectMeta struct {
	value *ObjectMeta
	isSet bool
}

func (v NullableObjectMeta) Get() *ObjectMeta {
	return v.value
}

func (v *NullableObjectMeta) Set(val *ObjectMeta) {
	v.value = val
	v.isSet = true
}

func (v NullableObjectMeta) IsSet() bool {
	return v.isSet
}

func (v *NullableObjectMeta) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableObjectMeta(val *ObjectMeta) *NullableObjectMeta {
	return &NullableObjectMeta{value: val, isSet: true}
}

func (v NullableObjectMeta) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableObjectMeta) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}