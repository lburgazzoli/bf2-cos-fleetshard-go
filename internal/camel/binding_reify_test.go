package camel

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/secrets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReify(t *testing.T) {
	t.Skip("skipping testing")

	var err error

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mctr-foo-secret",
			Namespace: "mctr-baz",
		},
	}
	configmap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mctr-foo-configmap",
			Namespace: "mctr-baz",
		},
	}

	err = secrets.SetStructuredData(&secret, "serviceAccount", ServiceAccount{
		ClientID:     "225143db-c506-4f3c-9925-0772a2d825cb",
		ClientSecret: "YWFr",
	})

	assert.Nil(t, err)

	err = secrets.SetStructuredData(&secret, "meta", ShardMetadata{
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

	err = secrets.SetStructuredData(&secret, "connector", map[string]interface{}{
		"azure_access_key": map[string]interface{}{
			"kind":  "base64",
			"value": "YWFr",
		},
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
			Connector: &cosv2.ManagedConnector{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mctr-foo",
					Namespace: "mctr-baz",
				},
				Spec: cosv2.ManagedConnectorSpec{
					ConnectorID:  "cid",
					DeploymentID: "did",
				},
			},
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
