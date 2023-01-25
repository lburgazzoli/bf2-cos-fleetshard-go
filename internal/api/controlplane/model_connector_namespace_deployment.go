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

// ConnectorNamespaceDeployment A connector namespace deployment
type ConnectorNamespaceDeployment struct {
	Id         string     `json:"id"`
	Kind       *string    `json:"kind,omitempty"`
	Href       *string    `json:"href,omitempty"`
	Owner      *string    `json:"owner,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	ModifiedAt *time.Time `json:"modified_at,omitempty"`
	Name       string     `json:"name"`
	// Name-value string annotations for resource
	Annotations     *map[string]string       `json:"annotations,omitempty"`
	ResourceVersion int64                    `json:"resource_version"`
	Quota           *ConnectorNamespaceQuota `json:"quota,omitempty"`
	ClusterId       string                   `json:"cluster_id"`
	// Namespace expiration timestamp in RFC 3339 format
	Expiration *string                  `json:"expiration,omitempty"`
	Tenant     ConnectorNamespaceTenant `json:"tenant"`
	Status     ConnectorNamespaceStatus `json:"status"`
}

// NewConnectorNamespaceDeployment instantiates a new ConnectorNamespaceDeployment object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConnectorNamespaceDeployment(id string, name string, resourceVersion int64, clusterId string, tenant ConnectorNamespaceTenant, status ConnectorNamespaceStatus) *ConnectorNamespaceDeployment {
	this := ConnectorNamespaceDeployment{}
	this.Id = id
	this.Name = name
	this.ResourceVersion = resourceVersion
	this.ClusterId = clusterId
	this.Tenant = tenant
	this.Status = status
	return &this
}

// NewConnectorNamespaceDeploymentWithDefaults instantiates a new ConnectorNamespaceDeployment object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConnectorNamespaceDeploymentWithDefaults() *ConnectorNamespaceDeployment {
	this := ConnectorNamespaceDeployment{}
	return &this
}

// GetId returns the Id field value
func (o *ConnectorNamespaceDeployment) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *ConnectorNamespaceDeployment) SetId(v string) {
	o.Id = v
}

// GetKind returns the Kind field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetKind() string {
	if o == nil || isNil(o.Kind) {
		var ret string
		return ret
	}
	return *o.Kind
}

// GetKindOk returns a tuple with the Kind field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetKindOk() (*string, bool) {
	if o == nil || isNil(o.Kind) {
		return nil, false
	}
	return o.Kind, true
}

// HasKind returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasKind() bool {
	if o != nil && !isNil(o.Kind) {
		return true
	}

	return false
}

// SetKind gets a reference to the given string and assigns it to the Kind field.
func (o *ConnectorNamespaceDeployment) SetKind(v string) {
	o.Kind = &v
}

// GetHref returns the Href field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetHref() string {
	if o == nil || isNil(o.Href) {
		var ret string
		return ret
	}
	return *o.Href
}

// GetHrefOk returns a tuple with the Href field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetHrefOk() (*string, bool) {
	if o == nil || isNil(o.Href) {
		return nil, false
	}
	return o.Href, true
}

// HasHref returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasHref() bool {
	if o != nil && !isNil(o.Href) {
		return true
	}

	return false
}

// SetHref gets a reference to the given string and assigns it to the Href field.
func (o *ConnectorNamespaceDeployment) SetHref(v string) {
	o.Href = &v
}

// GetOwner returns the Owner field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetOwner() string {
	if o == nil || isNil(o.Owner) {
		var ret string
		return ret
	}
	return *o.Owner
}

// GetOwnerOk returns a tuple with the Owner field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetOwnerOk() (*string, bool) {
	if o == nil || isNil(o.Owner) {
		return nil, false
	}
	return o.Owner, true
}

// HasOwner returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasOwner() bool {
	if o != nil && !isNil(o.Owner) {
		return true
	}

	return false
}

// SetOwner gets a reference to the given string and assigns it to the Owner field.
func (o *ConnectorNamespaceDeployment) SetOwner(v string) {
	o.Owner = &v
}

// GetCreatedAt returns the CreatedAt field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetCreatedAt() time.Time {
	if o == nil || isNil(o.CreatedAt) {
		var ret time.Time
		return ret
	}
	return *o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil || isNil(o.CreatedAt) {
		return nil, false
	}
	return o.CreatedAt, true
}

// HasCreatedAt returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasCreatedAt() bool {
	if o != nil && !isNil(o.CreatedAt) {
		return true
	}

	return false
}

// SetCreatedAt gets a reference to the given time.Time and assigns it to the CreatedAt field.
func (o *ConnectorNamespaceDeployment) SetCreatedAt(v time.Time) {
	o.CreatedAt = &v
}

// GetModifiedAt returns the ModifiedAt field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetModifiedAt() time.Time {
	if o == nil || isNil(o.ModifiedAt) {
		var ret time.Time
		return ret
	}
	return *o.ModifiedAt
}

// GetModifiedAtOk returns a tuple with the ModifiedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetModifiedAtOk() (*time.Time, bool) {
	if o == nil || isNil(o.ModifiedAt) {
		return nil, false
	}
	return o.ModifiedAt, true
}

// HasModifiedAt returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasModifiedAt() bool {
	if o != nil && !isNil(o.ModifiedAt) {
		return true
	}

	return false
}

// SetModifiedAt gets a reference to the given time.Time and assigns it to the ModifiedAt field.
func (o *ConnectorNamespaceDeployment) SetModifiedAt(v time.Time) {
	o.ModifiedAt = &v
}

// GetName returns the Name field value
func (o *ConnectorNamespaceDeployment) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *ConnectorNamespaceDeployment) SetName(v string) {
	o.Name = v
}

// GetAnnotations returns the Annotations field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetAnnotations() map[string]string {
	if o == nil || isNil(o.Annotations) {
		var ret map[string]string
		return ret
	}
	return *o.Annotations
}

// GetAnnotationsOk returns a tuple with the Annotations field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetAnnotationsOk() (*map[string]string, bool) {
	if o == nil || isNil(o.Annotations) {
		return nil, false
	}
	return o.Annotations, true
}

// HasAnnotations returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasAnnotations() bool {
	if o != nil && !isNil(o.Annotations) {
		return true
	}

	return false
}

// SetAnnotations gets a reference to the given map[string]string and assigns it to the Annotations field.
func (o *ConnectorNamespaceDeployment) SetAnnotations(v map[string]string) {
	o.Annotations = &v
}

// GetResourceVersion returns the ResourceVersion field value
func (o *ConnectorNamespaceDeployment) GetResourceVersion() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetResourceVersionOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ResourceVersion, true
}

// SetResourceVersion sets field value
func (o *ConnectorNamespaceDeployment) SetResourceVersion(v int64) {
	o.ResourceVersion = v
}

// GetQuota returns the Quota field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetQuota() ConnectorNamespaceQuota {
	if o == nil || isNil(o.Quota) {
		var ret ConnectorNamespaceQuota
		return ret
	}
	return *o.Quota
}

// GetQuotaOk returns a tuple with the Quota field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetQuotaOk() (*ConnectorNamespaceQuota, bool) {
	if o == nil || isNil(o.Quota) {
		return nil, false
	}
	return o.Quota, true
}

// HasQuota returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasQuota() bool {
	if o != nil && !isNil(o.Quota) {
		return true
	}

	return false
}

// SetQuota gets a reference to the given ConnectorNamespaceQuota and assigns it to the Quota field.
func (o *ConnectorNamespaceDeployment) SetQuota(v ConnectorNamespaceQuota) {
	o.Quota = &v
}

// GetClusterId returns the ClusterId field value
func (o *ConnectorNamespaceDeployment) GetClusterId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ClusterId
}

// GetClusterIdOk returns a tuple with the ClusterId field value
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetClusterIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ClusterId, true
}

// SetClusterId sets field value
func (o *ConnectorNamespaceDeployment) SetClusterId(v string) {
	o.ClusterId = v
}

// GetExpiration returns the Expiration field value if set, zero value otherwise.
func (o *ConnectorNamespaceDeployment) GetExpiration() string {
	if o == nil || isNil(o.Expiration) {
		var ret string
		return ret
	}
	return *o.Expiration
}

// GetExpirationOk returns a tuple with the Expiration field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetExpirationOk() (*string, bool) {
	if o == nil || isNil(o.Expiration) {
		return nil, false
	}
	return o.Expiration, true
}

// HasExpiration returns a boolean if a field has been set.
func (o *ConnectorNamespaceDeployment) HasExpiration() bool {
	if o != nil && !isNil(o.Expiration) {
		return true
	}

	return false
}

// SetExpiration gets a reference to the given string and assigns it to the Expiration field.
func (o *ConnectorNamespaceDeployment) SetExpiration(v string) {
	o.Expiration = &v
}

// GetTenant returns the Tenant field value
func (o *ConnectorNamespaceDeployment) GetTenant() ConnectorNamespaceTenant {
	if o == nil {
		var ret ConnectorNamespaceTenant
		return ret
	}

	return o.Tenant
}

// GetTenantOk returns a tuple with the Tenant field value
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetTenantOk() (*ConnectorNamespaceTenant, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Tenant, true
}

// SetTenant sets field value
func (o *ConnectorNamespaceDeployment) SetTenant(v ConnectorNamespaceTenant) {
	o.Tenant = v
}

// GetStatus returns the Status field value
func (o *ConnectorNamespaceDeployment) GetStatus() ConnectorNamespaceStatus {
	if o == nil {
		var ret ConnectorNamespaceStatus
		return ret
	}

	return o.Status
}

// GetStatusOk returns a tuple with the Status field value
// and a boolean to check if the value has been set.
func (o *ConnectorNamespaceDeployment) GetStatusOk() (*ConnectorNamespaceStatus, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Status, true
}

// SetStatus sets field value
func (o *ConnectorNamespaceDeployment) SetStatus(v ConnectorNamespaceStatus) {
	o.Status = v
}

func (o ConnectorNamespaceDeployment) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["id"] = o.Id
	}
	if !isNil(o.Kind) {
		toSerialize["kind"] = o.Kind
	}
	if !isNil(o.Href) {
		toSerialize["href"] = o.Href
	}
	if !isNil(o.Owner) {
		toSerialize["owner"] = o.Owner
	}
	if !isNil(o.CreatedAt) {
		toSerialize["created_at"] = o.CreatedAt
	}
	if !isNil(o.ModifiedAt) {
		toSerialize["modified_at"] = o.ModifiedAt
	}
	if true {
		toSerialize["name"] = o.Name
	}
	if !isNil(o.Annotations) {
		toSerialize["annotations"] = o.Annotations
	}
	if true {
		toSerialize["resource_version"] = o.ResourceVersion
	}
	if !isNil(o.Quota) {
		toSerialize["quota"] = o.Quota
	}
	if true {
		toSerialize["cluster_id"] = o.ClusterId
	}
	if !isNil(o.Expiration) {
		toSerialize["expiration"] = o.Expiration
	}
	if true {
		toSerialize["tenant"] = o.Tenant
	}
	if true {
		toSerialize["status"] = o.Status
	}
	return json.Marshal(toSerialize)
}

type NullableConnectorNamespaceDeployment struct {
	value *ConnectorNamespaceDeployment
	isSet bool
}

func (v NullableConnectorNamespaceDeployment) Get() *ConnectorNamespaceDeployment {
	return v.value
}

func (v *NullableConnectorNamespaceDeployment) Set(val *ConnectorNamespaceDeployment) {
	v.value = val
	v.isSet = true
}

func (v NullableConnectorNamespaceDeployment) IsSet() bool {
	return v.isSet
}

func (v *NullableConnectorNamespaceDeployment) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConnectorNamespaceDeployment(val *ConnectorNamespaceDeployment) *NullableConnectorNamespaceDeployment {
	return &NullableConnectorNamespaceDeployment{value: val, isSet: true}
}

func (v NullableConnectorNamespaceDeployment) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConnectorNamespaceDeployment) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
