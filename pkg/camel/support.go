package camel

import (
	camel "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stoewer/go-strcase"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

type EndpointBuilder struct {
	ref        corev1.ObjectReference
	properties map[string]interface{}
}

func (builder *EndpointBuilder) ApiVersion(val string) *EndpointBuilder {
	builder.ref.APIVersion = val
	return builder
}
func (builder *EndpointBuilder) Kind(val string) *EndpointBuilder {
	builder.ref.Kind = val
	return builder
}
func (builder *EndpointBuilder) Name(val string) *EndpointBuilder {
	builder.ref.Name = val
	return builder
}
func (builder *EndpointBuilder) Property(key string, val interface{}) *EndpointBuilder {
	builder.properties[key] = val
	return builder
}
func (builder *EndpointBuilder) Properties(properties map[string]interface{}) *EndpointBuilder {
	for k, v := range properties {
		// rude check, it should be enhanced
		if _, ok := v.(map[string]interface{}); ok {
			continue
		}

		k = strcase.LowerCamelCase(k)

		builder.properties[k] = v
	}

	return builder
}

func (builder *EndpointBuilder) PropertiesFrom(properties map[string]interface{}, prefix string) *EndpointBuilder {
	for k, v := range properties {
		// rude check, it should be enhanced
		if _, ok := v.(map[string]interface{}); ok {
			continue
		}

		if strings.HasPrefix(k, prefix+"_") {
			k = strings.TrimPrefix(k, prefix)
			k = strcase.LowerCamelCase(k)

			builder.properties[k] = v
		}
	}

	return builder
}
func (builder *EndpointBuilder) Build() (camel.Endpoint, error) {
	ref := builder.ref

	answer := camel.Endpoint{
		Ref: &ref,
	}

	if err := setProperties(&answer, builder.properties); err != nil {
		return answer, errors.Wrap(err, "error setting source properties")
	}

	return answer, nil
}

func NewEndpointBuilder() *EndpointBuilder {
	return &EndpointBuilder{
		ref:        corev1.ObjectReference{},
		properties: make(map[string]interface{}),
	}
}
