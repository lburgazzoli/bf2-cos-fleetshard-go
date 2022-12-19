package camel

import (
	camel "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/camel/endpoints"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/secrets"
	corev1 "k8s.io/api/core/v1"
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

	binding.Annotations["trait.camel.apache.org/container.image"] = meta.ConnectorImage
	binding.Annotations["trait.camel.apache.org/kamelets.enabled"] = "false"
	binding.Annotations["trait.camel.apache.org/jvm.enabled"] = "false"
	binding.Annotations["trait.camel.apache.org/logging.json"] = "false"
	binding.Annotations["trait.camel.apache.org/prometheus.enabled"] = "true"
	binding.Annotations["trait.camel.apache.org/prometheus.pod-monitor"] = "false"
	binding.Annotations["trait.camel.apache.org /health.enabled"] = "true"
	binding.Annotations["trait.camel.apache.org/health.readiness-probe-enabled"] = "true"
	binding.Annotations["trait.camel.apache.org/health.liveness-probe-enabled"] = "true"
	binding.Annotations["trait.camel.apache.org/deployment.enabled"] = "true"
	binding.Annotations["trait.camel.apache.org/deployment.strategy"] = "Recreate"

	binding.Annotations["trait.camel.apache.org/owner.target-labels"] = "[ \"cos.bf2.org/operator.type\", \"cos.bf2.org/deployment.id\", \"cos.bf2.org/connector.id\", \"cos.bf2.org/connector.type.id\" ]"
	binding.Annotations["trait.camel.apache.org/owner.target-annotations"] = ""

	// TODO: must be configurable
	binding.Annotations["trait.camel.apache.org/health.readiness-success-threshold"] = "1"
	binding.Annotations["trait.camel.apache.org/health.readiness-failure-threshold"] = "3"
	binding.Annotations["trait.camel.apache.org/health.readiness-period"] = "10"
	binding.Annotations["trait.camel.apache.org/health.readiness-timeout"] = "1"
	binding.Annotations["trait.camel.apache.org/health.liveness-success-threshold"] = "1"
	binding.Annotations["trait.camel.apache.org/health.liveness-failure-threshold"] = "3"
	binding.Annotations["trait.camel.apache.org/health.liveness-period"] = "10"
	binding.Annotations["trait.camel.apache.org/health.liveness-timeout"] = "1"

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
		src, err := endpoints.NewKameletBuilder(meta.Kamelets.Adapter.Name).
			Property("id", connector.Spec.ConnectorID+"-source").
			PropertiesFrom(config, meta.Kamelets.Adapter.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating source")
		}

		sink, err := endpoints.NewKameletBuilder(meta.Kamelets.Kafka.Name).
			Property("id", connector.Spec.ConnectorID+"-sink").
			Property("bootstrapServers", connector.Spec.Deployment.Kafka.URL).
			PropertyPlaceholder("user", ServiceAccountClientID).
			PropertyPlaceholder("password", ServiceAccountClientSecret).
			PropertiesFrom(config, meta.Kamelets.Kafka.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Source = src
		binding.Spec.Sink = sink

		break

	case ConnectorTypeSink:
		src, err := endpoints.NewKameletBuilder(meta.Kamelets.Kafka.Name).
			Property("id", connector.Spec.ConnectorID+"-source").
			Property("bootstrapServers", connector.Spec.Deployment.Kafka.URL).
			Property("consumerGroup", connector.Spec.ConnectorID).
			PropertyPlaceholder("user", ServiceAccountClientID).
			PropertyPlaceholder("password", ServiceAccountClientSecret).
			PropertiesFrom(config, meta.Kamelets.Kafka.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating source")
		}

		sink, err := endpoints.NewKameletBuilder(meta.Kamelets.Adapter.Name).
			Property("id", connector.Spec.ConnectorID+"-sink").
			PropertiesFrom(config, meta.Kamelets.Adapter.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Source = src
		binding.Spec.Sink = sink

		break
	}

	return binding, bindingSecret, bindingConfig, nil
}
