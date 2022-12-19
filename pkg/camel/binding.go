package camel

import (
	"encoding/base64"
	"fmt"
	camel "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stoewer/go-strcase"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/secrets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"strings"
)

const (
	SecretEntryConnector       string = "connector"
	SecretEntryServiceAccount  string = "serviceAccount"
	SecretEntryMeta            string = "meta"
	ServiceAccountClientID     string = "sa_client_id"
	ServiceAccountClientSecret string = "sa_client_secret"
)

func Reify(
	connector cos.ManagedConnector,
	secret corev1.Secret,
) (camel.KameletBinding, corev1.Secret, corev1.ConfigMap, error) {
	binding := camel.KameletBinding{}
	bindingSecret := corev1.Secret{}
	bindingConfig := corev1.ConfigMap{}

	var sa ServiceAccount
	var meta ShardMetadata
	var config map[string]interface{}

	if err := secrets.ExtractStructuredData(secret, SecretEntryServiceAccount, &sa); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding service account")
	}
	if err := secrets.ExtractStructuredData(secret, SecretEntryMeta, &meta); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding shard meta")
	}
	if err := secrets.ExtractStructuredData(secret, SecretEntryConnector, &config); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding config")
	}

	bindingSecret.StringData[ServiceAccountClientID] = sa.ClientID
	bindingSecret.StringData[ServiceAccountClientSecret] = sa.ClientSecret

	if err := extractSecrets(config, &bindingSecret); err != nil {
		return binding, bindingSecret, bindingConfig, err
	}
	if err := extractConfig(config, &bindingConfig); err != nil {
		return binding, bindingSecret, bindingConfig, err
	}

	switch meta.ConnectorType {
	case ConnectorTypeSource:
		src, err := NewEndpointBuilder().
			ApiVersion("camel.apache.org/v1alpha1").
			Kind("Kamelet").
			Name(meta.Kamelets.Adapter.Name).
			Property("id", connector.Spec.ConnectorID+"-source").
			PropertiesFrom(config, meta.Kamelets.Adapter.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error setting source properties")
		}

		binding.Spec.Source = src

		binding.Spec.Sink.Ref = &corev1.ObjectReference{
			APIVersion: "camel.apache.org/v1alpha1",
			Kind:       "Kamelet",
			Name:       meta.Kamelets.Kafka.Name,
		}

		sinkProperties := map[string]interface{}{
			"id":               connector.Spec.ConnectorID + "-sink",
			"bootstrapServers": connector.Spec.Deployment.Kafka.URL,
			"user":             "{{" + ServiceAccountClientID + "}}",
			"password":         "{{" + ServiceAccountClientSecret + "}}",
		}

		setEndpointProperties(sinkProperties, config, meta.Kamelets.Kafka.Prefix)

		if err := setProperties(&binding.Spec.Sink, sinkProperties); err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error setting sink properties")
		}

		break
	case ConnectorTypeSink:

		binding.Spec.Sink.Ref = &corev1.ObjectReference{
			APIVersion: "camel.apache.org/v1alpha1",
			Kind:       "Kamelet",
			Name:       meta.Kamelets.Adapter.Name,
		}

		sinkProperties := map[string]interface{}{
			"id": connector.Spec.ConnectorID + "-sink",
		}

		setEndpointProperties(sinkProperties, config, meta.Kamelets.Adapter.Prefix)

		if err := setProperties(&binding.Spec.Sink, sinkProperties); err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error setting sink properties")
		}

		binding.Spec.Source.Ref = &corev1.ObjectReference{
			APIVersion: "camel.apache.org/v1alpha1",
			Kind:       "Kamelet",
			Name:       meta.Kamelets.Kafka.Name,
		}

		sourceProperties := map[string]interface{}{
			"id":               connector.Spec.ConnectorID + "-source",
			"bootstrapServers": connector.Spec.Deployment.Kafka.URL,
			"consumerGroup":    connector.Spec.ConnectorID,
			"user":             "{{" + ServiceAccountClientID + "}}",
			"password":         "{{" + ServiceAccountClientSecret + "}}",
		}

		setEndpointProperties(sourceProperties, config, meta.Kamelets.Kafka.Prefix)

		if err := setProperties(&binding.Spec.Source, sourceProperties); err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error setting source properties")
		}

		break
	}

	return binding, bindingSecret, bindingConfig, nil
}

func extractSecrets(
	config map[string]interface{},
	secret *corev1.Secret) error {

	for k, v := range config {
		switch t := v.(type) {
		case map[string]interface{}:
			kind := t["kind"]
			value := t["value"]

			if kind != "base64" {
				return fmt.Errorf("unsupported kind: %s", kind)
			}

			decoded, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", value))
			if err != nil {
				return fmt.Errorf("error decoding secret: %s", k)
			}

			secret.StringData[k] = string(decoded)

			break
		default:
			break
		}
	}

	return nil
}

func extractConfig(
	config map[string]interface{},
	configMap *corev1.ConfigMap) error {

	configMap.Data["camel.main.route-controller-supervise-enabled"] = "true"
	configMap.Data["camel.main.route-controller-unhealthy-on-exhausted"] = "true"

	configMap.Data["camel.main.load-health-checks"] = "true"
	configMap.Data["camel.health.routesEnabled"] = "true"
	configMap.Data["camel.health.consumersEnabled"] = "true"
	configMap.Data["camel.health.registryEnabled"] = "true"

	// TODO: must be configurable
	configMap.Data["camel.main.route-controller-backoff-delay"] = "10s"
	configMap.Data["camel.main.route-controller-initial-delay"] = "0s"
	configMap.Data["camel.main.route-controller-backoff-multiplier"] = "1"
	configMap.Data["camel.main.route-controller-backoff-max-attempts"] = "6"

	// TODO: must be configurable
	configMap.Data["camel.main.exchange-factory"] = "prototype"
	configMap.Data["camel.main.exchange-factory-capacity"] = "100"
	configMap.Data["camel.main.exchange-factory-statistics-enabled"] = "false"

	return nil
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

func setEndpointProperties(properties map[string]interface{}, config map[string]interface{}, prefix string) {
	for k, v := range config {
		// rude check, it should be enhanced
		if _, ok := v.(map[string]interface{}); ok {
			continue
		}

		if strings.HasPrefix(k, prefix+"_") {
			k = strings.TrimPrefix(k, prefix)
			k = strcase.LowerCamelCase(k)

			properties[k] = v
		}
	}
}

func configureEndpoint(
	endpoint *camel.Endpoint,
	ke EndpointKamelet,
	ID string,
	config map[string]interface{},
) error {
	endpoint.Ref = &corev1.ObjectReference{
		APIVersion: "camel.apache.org/v1alpha1",
		Kind:       "Kamelet",
		Name:       ke.Name,
	}

	sourceProperties := make(map[string]interface{})
	sourceProperties["id"] = ID

	setEndpointProperties(sourceProperties, config, ke.Prefix)

	if err := setProperties(endpoint, sourceProperties); err != nil {
		return errors.Wrap(err, "error setting source properties")
	}

	return nil
}
