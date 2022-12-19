package endpoints

import (
	"encoding/json"
	camel "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stoewer/go-strcase"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

type Builder struct {
	ref        corev1.ObjectReference
	properties map[string]interface{}
}

func (builder *Builder) GroupVersion(gv schema.GroupVersion) *Builder {
	builder.ref.APIVersion = gv.String()
	return builder
}

func (builder *Builder) ApiVersion(val string) *Builder {
	builder.ref.APIVersion = val
	return builder
}

func (builder *Builder) Kind(val string) *Builder {
	builder.ref.Kind = val
	return builder
}

func (builder *Builder) Name(val string) *Builder {
	builder.ref.Name = val
	return builder
}

func (builder *Builder) Property(key string, val interface{}) *Builder {
	builder.properties[key] = val
	return builder
}

func (builder *Builder) PropertyPlaceholder(key string, val string) *Builder {
	if !strings.HasPrefix(val, "{{") {
		val = "{{" + val
	}
	if !strings.HasSuffix(val, "}}") {
		val = val + "}}"
	}

	return builder.Property(key, val)
}

func (builder *Builder) Properties(properties map[string]interface{}) *Builder {
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

func (builder *Builder) PropertiesFrom(properties map[string]interface{}, prefix string) *Builder {
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
func (builder *Builder) Build() (camel.Endpoint, error) {
	ref := builder.ref

	answer := camel.Endpoint{
		Ref: &ref,
	}

	if err := setProperties(&answer, builder.properties); err != nil {
		return answer, errors.Wrap(err, "error setting source properties")
	}

	return answer, nil
}

func NewBuilder() *Builder {
	return &Builder{
		ref:        corev1.ObjectReference{},
		properties: make(map[string]interface{}),
	}
}

func NewKameletBuilder(name string) *Builder {
	return &Builder{
		ref: corev1.ObjectReference{
			APIVersion: camel.SchemeGroupVersion.String(),
			Kind:       "Kamelet",
		},
		properties: make(map[string]interface{}),
	}
}

func setProperties(endpoint *camel.Endpoint, properties map[string]interface{}) error {
	data, err := json.Marshal(properties)
	if err != nil {
		return errors.Wrap(err, "unable to encode endpoint properties")
	}

	endpoint.Properties = &camel.EndpointProperties{
		RawMessage: data,
	}

	return nil
}
