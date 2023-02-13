package camel

import (
	"context"
	"encoding/json"
	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestReifyReady(t *testing.T) {
	var err error

	connectorName := xid.New().String()
	connectorNamespace := xid.New().String()

	connector := cosv2.ManagedConnector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connectorName,
			Namespace: connectorNamespace,
		},
		Spec: cosv2.ManagedConnectorSpec{
			ConnectorID:  xid.New().String(),
			DeploymentID: xid.New().String(),
			DesiredState: cosv2.DesiredStateReady,
			Kafka: cosv2.KafkaSpec{
				ID:  xid.New().String(),
				URL: "kafka.acme.com:443",
			},
		},
	}

	err = connector.Spec.DeploymentMeta.Set(ShardMetadata{
		Annotations: map[string]string{
			"trait.camel.apache.org/container.request-cpu":                "0.20",
			"trait.camel.apache.org/container.request-memory":             "128M",
			"trait.camel.apache.org/deployment.progress-deadline-seconds": "30",
		},
		ConnectorImage:       "quay.io/foo/bar:1.0",
		ConnectorType:        "source",
		ConnectorRevision:    1,
		Consumes:             "application/octet-stream",
		Produces:             "application/octet-stream",
		ErrorHandlerStrategy: "stop",
		Kamelets: Kamelets{
			Adapter: EndpointKamelet{
				Name:   "cos-azure-storage-blob-source",
				Prefix: "azure",
			},
			Kafka: EndpointKamelet{
				Name:   "cos-kafka-sink",
				Prefix: "kafka",
			},
		},
		Operators: []Operator{
			{
				Type:    "camel-connector-operator",
				Version: "[1.0.0,2.0.0)",
			},
		},
	})

	assert.Nil(t, err)

	err = connector.Spec.DeploymentConfig.Set(map[string]interface{}{
		"azure_access_key":     "{{azure_access_key}}",
		"azure_account_name":   "foo",
		"azure_container_name": "foo/csv/1",
		"data_shape": map[string]interface{}{
			"produces": map[string]interface{}{
				"format": "application/octet-stream",
			},
		},
		"error_handler": map[string]interface{}{
			"stop": map[string]interface{}{},
		},
		"kafka_topic": "bar",
	})

	assert.Nil(t, err)

	s := runtime.NewScheme()
	assert.Nil(t, cosv2.AddToScheme(s))
	assert.Nil(t, camelv1alpha1.AddToScheme(s))
	assert.Nil(t, camelv1.AddToScheme(s))
	assert.Nil(t, clientgoscheme.AddToScheme(s))

	c := fake.NewClientBuilder().
		WithScheme(s).
		Build()

	rc := controller.ReconciliationContext{
		C:      context.TODO(),
		Client: c,
		NamespacedName: types.NamespacedName{
			Name:      connectorName,
			Namespace: connectorNamespace,
		},
		Connector: &connector,
		Secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectorName + "-secret",
				Namespace: connectorNamespace,
			},
			Data: map[string][]byte{
				cosmeta.ServiceAccountClientID:     []byte(xid.New().String()),
				cosmeta.ServiceAccountClientSecret: []byte(xid.New().String()),
			},
		},
		ConfigMap: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectorName + "config",
				Namespace: connectorNamespace,
			},
		},
	}

	_, err = Apply(rc)
	assert.Nil(t, err)

	kb := camelv1alpha1.KameletBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connectorName,
			Namespace: connectorNamespace,
		},
	}

	assert.Nil(t, resources.Get(rc.C, rc.Client, &kb))

	assert.NotNil(t, kb.Spec.Replicas)
	assert.Equal(t, int32(1), *kb.Spec.Replicas)
	assert.Equal(t, camelv1.TraitProfileOpenShift, kb.Spec.Integration.Profile)
	assert.Contains(t, kb.Spec.Integration.Traits.Environment.Vars, "CONNECTOR_ID="+rc.Connector.Spec.ConnectorID)

	sourceProps := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal(kb.Spec.Source.Properties.RawMessage, &sourceProps))

	assert.Equal(t, "Kamelet", kb.Spec.Source.Ref.Kind)
	assert.Equal(t, "cos-azure-storage-blob-source", kb.Spec.Source.Ref.Name)
	assert.Equal(t, "camel.apache.org/v1alpha1", kb.Spec.Source.Ref.APIVersion)
	assert.Equal(t, "{{azure_access_key}}", sourceProps["accessKey"])
	assert.Equal(t, "foo", sourceProps["accountName"])
	assert.Equal(t, "foo/csv/1", sourceProps["containerName"])
	assert.Contains(t, sourceProps, "id")

	assert.Equal(t, "Kamelet", kb.Spec.Sink.Ref.Kind)
	assert.Equal(t, "cos-kafka-sink", kb.Spec.Sink.Ref.Name)
	assert.Equal(t, "camel.apache.org/v1alpha1", kb.Spec.Sink.Ref.APIVersion)

	sinkProps := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal(kb.Spec.Sink.Properties.RawMessage, &sinkProps))

	assert.Equal(t, "kafka.acme.com:443", sinkProps["bootstrapServers"])
	assert.Equal(t, "{{sa_client_secret}}", sinkProps["password"])
	assert.Equal(t, "{{sa_client_id}}", sinkProps["user"])
	assert.Equal(t, "org.bf2.cos.connector.camel.serdes.bytes.ByteArraySerializer", sinkProps["valueSerializer"])
	assert.Equal(t, "bar", sinkProps["topic"])
	assert.Contains(t, sinkProps, "id")

	assert.Len(t, kb.Spec.Steps, 1)
	assert.Equal(t, "Kamelet", kb.Spec.Steps[0].Ref.Kind)
	assert.Equal(t, "cos-encoder-bytearray-action", kb.Spec.Steps[0].Ref.Name)
	assert.Equal(t, "camel.apache.org/v1alpha1", kb.Spec.Steps[0].Ref.APIVersion)

	/*
		d, err := json.MarshalIndent(&kb, "", "  ")
		assert.Nil(t, err)

		fmt.Println(string(d))
	*/

}
