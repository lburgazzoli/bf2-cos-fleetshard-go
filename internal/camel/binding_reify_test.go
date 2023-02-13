package camel

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReifySimple(t *testing.T) {

	var err error

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mctr-foo-secret",
			Namespace: "mctr-baz",
		},
		Data: map[string][]byte{
			cosmeta.ServiceAccountClientID:     []byte("AD828C28-34F9-4DCA-97FF-C6AD60E78CD9"),
			cosmeta.ServiceAccountClientSecret: []byte("AD828C28-34F9-4DCA-97FF-C6AD60E78CD9"),
		},
	}
	configmap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mctr-foo-configmap",
			Namespace: "mctr-baz",
		},
	}

	connector := cosv2.ManagedConnector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mctr-foo",
			Namespace: "mctr-baz",
		},
		Spec: cosv2.ManagedConnectorSpec{
			ConnectorID:  "cid",
			DeploymentID: "did",
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

	b, bs, bc, err := reify(
		&controller.ReconciliationContext{
			Connector: &connector,
			Secret:    &secret,
			ConfigMap: &configmap,
		},
	)

	assert.Nil(t, err)

	d, err := json.MarshalIndent(&b, "", "  ")
	assert.Nil(t, err)

	ds, err := json.MarshalIndent(&bs, "", "  ")
	assert.Nil(t, err)

	dc, err := json.MarshalIndent(&bc, "", "  ")
	assert.Nil(t, err)

	fmt.Println(string(d))
	fmt.Println(string(ds))
	fmt.Println(string(dc))

}
