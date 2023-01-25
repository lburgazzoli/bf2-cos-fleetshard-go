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

// ConnectorDeploymentSpec Holds the deployment specification of a connector
type ConnectorDeploymentSpec struct {
	ServiceAccount           *ServiceAccount                   `json:"service_account,omitempty"`
	Kafka                    *KafkaConnectionSettings          `json:"kafka,omitempty"`
	SchemaRegistry           *SchemaRegistryConnectionSettings `json:"schema_registry,omitempty"`
	ConnectorId              *string                           `json:"connector_id,omitempty"`
	ConnectorResourceVersion *int64                            `json:"connector_resource_version,omitempty"`
	ConnectorTypeId          *string                           `json:"connector_type_id,omitempty"`
	NamespaceId              *string                           `json:"namespace_id,omitempty"`
	ConnectorSpec            map[string]interface{}            `json:"connector_spec,omitempty"`
	// an optional operator id that the connector should be run under.
	OperatorId    *string                `json:"operator_id,omitempty"`
	DesiredState  *ConnectorDesiredState `json:"desired_state,omitempty"`
	ShardMetadata map[string]interface{} `json:"shard_metadata,omitempty"`
}

// NewConnectorDeploymentSpec instantiates a new ConnectorDeploymentSpec object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConnectorDeploymentSpec() *ConnectorDeploymentSpec {
	this := ConnectorDeploymentSpec{}
	return &this
}

// NewConnectorDeploymentSpecWithDefaults instantiates a new ConnectorDeploymentSpec object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConnectorDeploymentSpecWithDefaults() *ConnectorDeploymentSpec {
	this := ConnectorDeploymentSpec{}
	return &this
}

// GetServiceAccount returns the ServiceAccount field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetServiceAccount() ServiceAccount {
	if o == nil || isNil(o.ServiceAccount) {
		var ret ServiceAccount
		return ret
	}
	return *o.ServiceAccount
}

// GetServiceAccountOk returns a tuple with the ServiceAccount field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetServiceAccountOk() (*ServiceAccount, bool) {
	if o == nil || isNil(o.ServiceAccount) {
		return nil, false
	}
	return o.ServiceAccount, true
}

// HasServiceAccount returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasServiceAccount() bool {
	if o != nil && !isNil(o.ServiceAccount) {
		return true
	}

	return false
}

// SetServiceAccount gets a reference to the given ServiceAccount and assigns it to the ServiceAccount field.
func (o *ConnectorDeploymentSpec) SetServiceAccount(v ServiceAccount) {
	o.ServiceAccount = &v
}

// GetKafka returns the Kafka field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetKafka() KafkaConnectionSettings {
	if o == nil || isNil(o.Kafka) {
		var ret KafkaConnectionSettings
		return ret
	}
	return *o.Kafka
}

// GetKafkaOk returns a tuple with the Kafka field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetKafkaOk() (*KafkaConnectionSettings, bool) {
	if o == nil || isNil(o.Kafka) {
		return nil, false
	}
	return o.Kafka, true
}

// HasKafka returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasKafka() bool {
	if o != nil && !isNil(o.Kafka) {
		return true
	}

	return false
}

// SetKafka gets a reference to the given KafkaConnectionSettings and assigns it to the Kafka field.
func (o *ConnectorDeploymentSpec) SetKafka(v KafkaConnectionSettings) {
	o.Kafka = &v
}

// GetSchemaRegistry returns the SchemaRegistry field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetSchemaRegistry() SchemaRegistryConnectionSettings {
	if o == nil || isNil(o.SchemaRegistry) {
		var ret SchemaRegistryConnectionSettings
		return ret
	}
	return *o.SchemaRegistry
}

// GetSchemaRegistryOk returns a tuple with the SchemaRegistry field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetSchemaRegistryOk() (*SchemaRegistryConnectionSettings, bool) {
	if o == nil || isNil(o.SchemaRegistry) {
		return nil, false
	}
	return o.SchemaRegistry, true
}

// HasSchemaRegistry returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasSchemaRegistry() bool {
	if o != nil && !isNil(o.SchemaRegistry) {
		return true
	}

	return false
}

// SetSchemaRegistry gets a reference to the given SchemaRegistryConnectionSettings and assigns it to the SchemaRegistry field.
func (o *ConnectorDeploymentSpec) SetSchemaRegistry(v SchemaRegistryConnectionSettings) {
	o.SchemaRegistry = &v
}

// GetConnectorId returns the ConnectorId field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetConnectorId() string {
	if o == nil || isNil(o.ConnectorId) {
		var ret string
		return ret
	}
	return *o.ConnectorId
}

// GetConnectorIdOk returns a tuple with the ConnectorId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetConnectorIdOk() (*string, bool) {
	if o == nil || isNil(o.ConnectorId) {
		return nil, false
	}
	return o.ConnectorId, true
}

// HasConnectorId returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasConnectorId() bool {
	if o != nil && !isNil(o.ConnectorId) {
		return true
	}

	return false
}

// SetConnectorId gets a reference to the given string and assigns it to the ConnectorId field.
func (o *ConnectorDeploymentSpec) SetConnectorId(v string) {
	o.ConnectorId = &v
}

// GetConnectorResourceVersion returns the ConnectorResourceVersion field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetConnectorResourceVersion() int64 {
	if o == nil || isNil(o.ConnectorResourceVersion) {
		var ret int64
		return ret
	}
	return *o.ConnectorResourceVersion
}

// GetConnectorResourceVersionOk returns a tuple with the ConnectorResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetConnectorResourceVersionOk() (*int64, bool) {
	if o == nil || isNil(o.ConnectorResourceVersion) {
		return nil, false
	}
	return o.ConnectorResourceVersion, true
}

// HasConnectorResourceVersion returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasConnectorResourceVersion() bool {
	if o != nil && !isNil(o.ConnectorResourceVersion) {
		return true
	}

	return false
}

// SetConnectorResourceVersion gets a reference to the given int64 and assigns it to the ConnectorResourceVersion field.
func (o *ConnectorDeploymentSpec) SetConnectorResourceVersion(v int64) {
	o.ConnectorResourceVersion = &v
}

// GetConnectorTypeId returns the ConnectorTypeId field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetConnectorTypeId() string {
	if o == nil || isNil(o.ConnectorTypeId) {
		var ret string
		return ret
	}
	return *o.ConnectorTypeId
}

// GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetConnectorTypeIdOk() (*string, bool) {
	if o == nil || isNil(o.ConnectorTypeId) {
		return nil, false
	}
	return o.ConnectorTypeId, true
}

// HasConnectorTypeId returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasConnectorTypeId() bool {
	if o != nil && !isNil(o.ConnectorTypeId) {
		return true
	}

	return false
}

// SetConnectorTypeId gets a reference to the given string and assigns it to the ConnectorTypeId field.
func (o *ConnectorDeploymentSpec) SetConnectorTypeId(v string) {
	o.ConnectorTypeId = &v
}

// GetNamespaceId returns the NamespaceId field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetNamespaceId() string {
	if o == nil || isNil(o.NamespaceId) {
		var ret string
		return ret
	}
	return *o.NamespaceId
}

// GetNamespaceIdOk returns a tuple with the NamespaceId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetNamespaceIdOk() (*string, bool) {
	if o == nil || isNil(o.NamespaceId) {
		return nil, false
	}
	return o.NamespaceId, true
}

// HasNamespaceId returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasNamespaceId() bool {
	if o != nil && !isNil(o.NamespaceId) {
		return true
	}

	return false
}

// SetNamespaceId gets a reference to the given string and assigns it to the NamespaceId field.
func (o *ConnectorDeploymentSpec) SetNamespaceId(v string) {
	o.NamespaceId = &v
}

// GetConnectorSpec returns the ConnectorSpec field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetConnectorSpec() map[string]interface{} {
	if o == nil || isNil(o.ConnectorSpec) {
		var ret map[string]interface{}
		return ret
	}
	return o.ConnectorSpec
}

// GetConnectorSpecOk returns a tuple with the ConnectorSpec field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetConnectorSpecOk() (map[string]interface{}, bool) {
	if o == nil || isNil(o.ConnectorSpec) {
		return map[string]interface{}{}, false
	}
	return o.ConnectorSpec, true
}

// HasConnectorSpec returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasConnectorSpec() bool {
	if o != nil && !isNil(o.ConnectorSpec) {
		return true
	}

	return false
}

// SetConnectorSpec gets a reference to the given map[string]interface{} and assigns it to the ConnectorSpec field.
func (o *ConnectorDeploymentSpec) SetConnectorSpec(v map[string]interface{}) {
	o.ConnectorSpec = v
}

// GetOperatorId returns the OperatorId field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetOperatorId() string {
	if o == nil || isNil(o.OperatorId) {
		var ret string
		return ret
	}
	return *o.OperatorId
}

// GetOperatorIdOk returns a tuple with the OperatorId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetOperatorIdOk() (*string, bool) {
	if o == nil || isNil(o.OperatorId) {
		return nil, false
	}
	return o.OperatorId, true
}

// HasOperatorId returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasOperatorId() bool {
	if o != nil && !isNil(o.OperatorId) {
		return true
	}

	return false
}

// SetOperatorId gets a reference to the given string and assigns it to the OperatorId field.
func (o *ConnectorDeploymentSpec) SetOperatorId(v string) {
	o.OperatorId = &v
}

// GetDesiredState returns the DesiredState field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetDesiredState() ConnectorDesiredState {
	if o == nil || isNil(o.DesiredState) {
		var ret ConnectorDesiredState
		return ret
	}
	return *o.DesiredState
}

// GetDesiredStateOk returns a tuple with the DesiredState field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetDesiredStateOk() (*ConnectorDesiredState, bool) {
	if o == nil || isNil(o.DesiredState) {
		return nil, false
	}
	return o.DesiredState, true
}

// HasDesiredState returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasDesiredState() bool {
	if o != nil && !isNil(o.DesiredState) {
		return true
	}

	return false
}

// SetDesiredState gets a reference to the given ConnectorDesiredState and assigns it to the DesiredState field.
func (o *ConnectorDeploymentSpec) SetDesiredState(v ConnectorDesiredState) {
	o.DesiredState = &v
}

// GetShardMetadata returns the ShardMetadata field value if set, zero value otherwise.
func (o *ConnectorDeploymentSpec) GetShardMetadata() map[string]interface{} {
	if o == nil || isNil(o.ShardMetadata) {
		var ret map[string]interface{}
		return ret
	}
	return o.ShardMetadata
}

// GetShardMetadataOk returns a tuple with the ShardMetadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorDeploymentSpec) GetShardMetadataOk() (map[string]interface{}, bool) {
	if o == nil || isNil(o.ShardMetadata) {
		return map[string]interface{}{}, false
	}
	return o.ShardMetadata, true
}

// HasShardMetadata returns a boolean if a field has been set.
func (o *ConnectorDeploymentSpec) HasShardMetadata() bool {
	if o != nil && !isNil(o.ShardMetadata) {
		return true
	}

	return false
}

// SetShardMetadata gets a reference to the given map[string]interface{} and assigns it to the ShardMetadata field.
func (o *ConnectorDeploymentSpec) SetShardMetadata(v map[string]interface{}) {
	o.ShardMetadata = v
}

func (o ConnectorDeploymentSpec) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.ServiceAccount) {
		toSerialize["service_account"] = o.ServiceAccount
	}
	if !isNil(o.Kafka) {
		toSerialize["kafka"] = o.Kafka
	}
	if !isNil(o.SchemaRegistry) {
		toSerialize["schema_registry"] = o.SchemaRegistry
	}
	if !isNil(o.ConnectorId) {
		toSerialize["connector_id"] = o.ConnectorId
	}
	if !isNil(o.ConnectorResourceVersion) {
		toSerialize["connector_resource_version"] = o.ConnectorResourceVersion
	}
	if !isNil(o.ConnectorTypeId) {
		toSerialize["connector_type_id"] = o.ConnectorTypeId
	}
	if !isNil(o.NamespaceId) {
		toSerialize["namespace_id"] = o.NamespaceId
	}
	if !isNil(o.ConnectorSpec) {
		toSerialize["connector_spec"] = o.ConnectorSpec
	}
	if !isNil(o.OperatorId) {
		toSerialize["operator_id"] = o.OperatorId
	}
	if !isNil(o.DesiredState) {
		toSerialize["desired_state"] = o.DesiredState
	}
	if !isNil(o.ShardMetadata) {
		toSerialize["shard_metadata"] = o.ShardMetadata
	}
	return json.Marshal(toSerialize)
}

type NullableConnectorDeploymentSpec struct {
	value *ConnectorDeploymentSpec
	isSet bool
}

func (v NullableConnectorDeploymentSpec) Get() *ConnectorDeploymentSpec {
	return v.value
}

func (v *NullableConnectorDeploymentSpec) Set(val *ConnectorDeploymentSpec) {
	v.value = val
	v.isSet = true
}

func (v NullableConnectorDeploymentSpec) IsSet() bool {
	return v.isSet
}

func (v *NullableConnectorDeploymentSpec) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConnectorDeploymentSpec(val *ConnectorDeploymentSpec) *NullableConnectorDeploymentSpec {
	return &NullableConnectorDeploymentSpec{value: val, isSet: true}
}

func (v NullableConnectorDeploymentSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConnectorDeploymentSpec) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
