package camel

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestReify(t *testing.T) {

	b, bs, bc, err := Reify(
		cosv2.ManagedConnector{
			Spec: cosv2.ManagedConnectorSpec{
				ConnectorID: "cid",
			},
		},
		corev1.Secret{
			Data: map[string][]byte{
				"connector":      []byte("{\"azure_access_key\":{\"kind\":\"base64\",\"value\":\"YWFr\"},\"azure_account_name\":\"foo\",\"azure_container_name\":\"foo/csv/dir2\",\"data_shape\":{\"produces\":{\"format\":\"application/octet-stream\"}},\"error_handler\":{\"stop\":{}},\"kafka_topic\":\"techv-topic\"}"),
				"meta":           []byte("{\"annotations\":{\"trait.camel.apache.org/container.request-cpu\":\"0.20\",\"trait.camel.apache.org/container.request-memory\":\"128M\",\"trait.camel.apache.org/deployment.progress-deadline-seconds\":\"30\"},\"connector_image\":\"quay.io/foo/bar:89015f237880c5b81a2d3b4587f3ec0692a83cea\",\"connector_revision\":74,\"connector_type\":\"source\",\"consumes\":\"application/octet-stream\",\"error_handler_strategy\":\"stop\",\"kamelets\":{\"adapter\":{\"name\":\"cos-azure-storage-blob-source\",\"prefix\":\"azure\"},\"kafka\":{\"name\":\"cos-kafka-sink\",\"prefix\":\"kafka\"}},\"operators\":[{\"type\":\"camel-connector-operator\",\"version\":\"[1.0.0,2.0.0)\"}],\"produces\":\"application/octet-stream\"}"),
				"serviceAccount": []byte("{\"client_id\":\"225143db-c506-4f3c-9925-0772a2d825cb\",\"client_secret\":\"YWFr\"}"),
			},
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
