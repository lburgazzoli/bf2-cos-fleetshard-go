package camel

import (
	"context"
	"encoding/json"
	"fmt"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ConnectorType string

const (
	ConnectorTypeSource ConnectorType = "source"
	ConnectorTypeSink   ConnectorType = "sink"

	TraitGroup string = "trait.camel.apache.org"

	ContentTypeBinary     = "application/octet-stream"
	ContentTypeJSON       = "application/json"
	ContentTypeText       = "text/plain"
	ContentTypeAvroBinary = "avro/binary"
)

type ServiceAccount struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Operator struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type EndpointKamelet struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

type Kamelets struct {
	Adapter     EndpointKamelet   `json:"adapter"`
	Kafka       EndpointKamelet   `json:"kafka"`
	Annotations map[string]string `json:"processors,omitempty"`
}

type ShardMetadata struct {
	ConnectorImage       string            `json:"connector_image"`
	ConnectorType        ConnectorType     `json:"connector_type"`
	ConnectorRevision    int64             `json:"connector_revision"`
	Annotations          map[string]string `json:"annotations,omitempty"`
	Operators            []Operator        `json:"operators,omitempty"`
	Kamelets             Kamelets          `json:"kamelets"`
	Consumes             string            `json:"consumes"`
	ConsumesClass        string            `json:"consumes_class"`
	Produces             string            `json:"produces"`
	ProducesClass        string            `json:"produces_class"`
	ErrorHandlerStrategy string            `json:"error_handler_strategy"`
}

type DataShape struct {
	Format string `json:"format"`
}

type DataShapeSpec struct {
	Consumes *DataShape `json:"consumes,omitempty"`
	Produces *DataShape `json:"produces,omitempty"`
}

type StopErrorHandler struct {
}
type LogErrorHandler struct {
}
type DLQErrorHandler struct {
}

type ErrorHandlerSpec struct {
	Stop *StopErrorHandler `json:"stop,omitempty"`
	Log  *LogErrorHandler  `json:"log,omitempty"`
	DLQ  *DLQErrorHandler  `json:"dead_letter_queue,omitempty"`
}

type ConnectorConfiguration struct {
	DataShape    DataShapeSpec          `json:"data_shape,omitempty"`
	ErrorHandler ErrorHandlerSpec       `json:"error_handler,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

func (c *ConnectorConfiguration) UnmarshalJSON(data []byte) error {

	tmp := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if d, ok := tmp["data_shape"]; ok {
		if err := json.Unmarshal(d, &c.DataShape); err != nil {
			return err
		}

		delete(tmp, "data_shape")
	}

	if d, ok := tmp["error_handler"]; ok {
		if err := json.Unmarshal(d, &c.ErrorHandler); err != nil {
			return err
		}

		delete(tmp, "error_handler")
	}

	c.Properties = make(map[string]interface{})

	for k, v := range tmp {
		var val interface{}

		if err := json.Unmarshal(v, &val); err != nil {
			return err
		}

		c.Properties[k] = val
	}

	return nil
}

type ReconciliationContext struct {
	client.Client
	types.NamespacedName

	M manager.Manager
	C context.Context

	Connector *cos.ManagedConnector
	Secret    *corev1.Secret
}

func (rc *ReconciliationContext) PatchDependant(source client.Object, target client.Object) error {
	target.GetAnnotations()[meta.MetaConnectorRevision] = fmt.Sprintf("%d", rc.Connector.Spec.Deployment.ConnectorResourceVersion)
	target.GetAnnotations()[meta.MetaDeploymentRevision] = fmt.Sprintf("%d", rc.Connector.Spec.Deployment.DeploymentResourceVersion)

	return controller.Patch(rc.C, rc.Client, source, target)
}

func (rc *ReconciliationContext) GetDependant(obj client.Object, opts ...client.GetOption) error {
	return rc.Client.Get(rc.C, rc.NamespacedName, obj, opts...)
}

func (rc *ReconciliationContext) DeleteDependant(obj client.Object, opts ...client.DeleteOption) error {
	return rc.Client.Delete(rc.C, obj, opts...)
}
