package camel

import (
	"encoding/base64"
	"fmt"
	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	"github.com/apache/camel-k/pkg/apis/camel/v1/trait"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/camel/endpoints"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/configmaps"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/secrets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
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
) (camelv1alpha1.KameletBinding, corev1.Secret, corev1.ConfigMap, error) {

	binding := camelv1alpha1.KameletBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connector.Name,
			Namespace: connector.Namespace,
		},
	}
	bindingSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connector.Name,
			Namespace: connector.Namespace,
		},
	}
	bindingConfig := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connector.Name,
			Namespace: connector.Namespace,
		},
	}

	var sa ServiceAccount
	var meta ShardMetadata
	var config map[string]interface{}
	var cc ConnectorConfiguration

	if err := secrets.ExtractStructuredData(secret, SecretEntryServiceAccount, &sa); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding service account")
	}
	if err := secrets.ExtractStructuredData(secret, SecretEntryMeta, &meta); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding shard meta")
	}
	if err := secrets.ExtractStructuredData(secret, SecretEntryConnector, &config); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding config")
	}
	if err := secrets.ExtractStructuredData(secret, SecretEntryConnector, &cc); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding config")
	}

	// TODO: improve
	if cc.DataShape == nil {
		cc.DataShape = &DataShapeSpec{}
	}
	if cc.DataShape.Consumes == nil {
		cc.DataShape.Consumes = &DataShape{}
	}
	if cc.DataShape.Consumes.Format == "" {
		cc.DataShape.Consumes.Format = meta.Consumes
	}
	if cc.DataShape.Produces == nil {
		cc.DataShape.Produces = &DataShape{}
	}
	if cc.DataShape.Produces.Format == "" {
		cc.DataShape.Produces.Format = meta.Produces
	}

	if binding.Annotations == nil {
		binding.Annotations = make(map[string]string)
	}

	for k, v := range meta.Annotations {
		binding.Annotations[k] = v
	}

	// TODO: handle errors
	_ = setTrait(&binding, "container.image", meta.ConnectorImage)
	_ = setTrait(&binding, "kamelets.enabled", "false")
	_ = setTrait(&binding, "jvm.enabled", "false")
	_ = setTrait(&binding, "logging.json", "false")
	_ = setTrait(&binding, "prometheus.enabled", "true")
	_ = setTrait(&binding, "prometheus.pod-monitor", "false")
	_ = setTrait(&binding, "health.enabled", "true")
	_ = setTrait(&binding, "health.readiness-probe-enabled", "true")
	_ = setTrait(&binding, "health.liveness-probe-enabled", "true")
	_ = setTrait(&binding, "deployment.enabled", "true")
	_ = setTrait(&binding, "deployment.strategy", "Recreate")

	_ = setTrait(&binding,
		"owner.target-labels",
		"cos.bf2.org/operator.type",
		"cos.bf2.org/deployment.id",
		"cos.bf2.org/connector.id",
		"cos.bf2.org/connector.type.id")

	_ = setTrait(&binding,
		"owner.target-annotations",
		"")

	// TODO: must be configurable
	_ = setTrait(&binding, "health.readiness-success-threshold", "1")
	_ = setTrait(&binding, "health.readiness-failure-threshold", "3")
	_ = setTrait(&binding, "health.readiness-period", "10")
	_ = setTrait(&binding, "health.readiness-timeout", "1")
	_ = setTrait(&binding, "health.liveness-success-threshold", "1")
	_ = setTrait(&binding, "health.liveness-failure-threshold", "3")
	_ = setTrait(&binding, "health.liveness-period", "10")
	_ = setTrait(&binding, "health.liveness-timeout", "1")

	if bindingSecret.StringData == nil {
		bindingSecret.StringData = make(map[string]string)
	}

	sad, err := base64.StdEncoding.DecodeString(sa.ClientSecret)
	if err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding service account")
	}

	bindingSecret.StringData[ServiceAccountClientID] = sa.ClientID
	bindingSecret.StringData[ServiceAccountClientSecret] = string(sad)

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
			Property("valueSerializer", "org.bf2.cos.connector.camel.serdes.bytes.ByteArraySerializer").
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
			Property("valueDeserializer", "org.bf2.cos.connector.camel.serdes.bytes.ByteArrayDeserializer").
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

	if err := configureSteps(&binding, cc); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error configuring steps")
	}

	scs, err := secrets.ComputeDigest(bindingSecret)
	if err != nil {
		return binding, bindingSecret, bindingConfig, err
	}

	ccs, err := configmaps.ComputeDigest(bindingConfig)
	if err != nil {
		return binding, bindingSecret, bindingConfig, err
	}

	tcs, err := computeTraitsDigest(binding)
	if err != nil {
		return binding, bindingSecret, bindingConfig, err
	}

	binding.Spec.Integration = &camelv1.IntegrationSpec{
		Traits: camelv1.Traits{
			Environment: &trait.EnvironmentTrait{
				Vars: []string{
					"CONNECTOR_ID=" + connector.Spec.ConnectorID,
					"CONNECTOR_DEPLOYMENT_ID=" + connector.Spec.DeploymentID,
					"CONNECTOR_SECRET_NAME=" + bindingSecret.Name,
					"CONNECTOR_CONFIGMAP_NAME=" + bindingConfig.Name,
					"CONNECTOR_SECRET_CHECKSUM=" + scs,
					"CONNECTOR_CONFIGMAP_CHECKSUM=" + ccs,
					"CONNECTOR_TRAITS_CHECKSUM=" + tcs,
				},
			},
		},
	}

	sort.Strings(binding.Spec.Integration.Traits.Environment.Vars)

	return binding, bindingSecret, bindingConfig, nil
}

func configureSteps(binding *camelv1alpha1.KameletBinding, cc ConnectorConfiguration) error {

	switch cc.DataShape.Consumes.Format {
	case "":
		break
	case "application/json":
		step, err := endpoints.NewKameletBuilder("cos-decoder-json-action").Build()
		if err != nil {
			return errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Steps = append(binding.Spec.Steps, step)
	case "avro/binary":
		step, err := endpoints.NewKameletBuilder("cos-decoder-avro-action").Build()
		if err != nil {
			return errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Steps = append(binding.Spec.Steps, step)
	case "text/plain":
	case "application/octet-stream":
		break
	default:
		return fmt.Errorf("unsupported format %s", cc.DataShape.Consumes.Format)

	}

	switch cc.DataShape.Produces.Format {
	case "":
		break
	case "application/json":
		step, err := endpoints.NewKameletBuilder("cos-encoder-json-action").Build()
		if err != nil {
			return errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Steps = append(binding.Spec.Steps, step)
	case "avro/binary":
		step, err := endpoints.NewKameletBuilder("cos-encoder-avro-action").Build()
		if err != nil {
			return errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Steps = append(binding.Spec.Steps, step)
	case "text/plain":
		step, err := endpoints.NewKameletBuilder("cos-encoder-string-action").Build()
		if err != nil {
			return errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Steps = append(binding.Spec.Steps, step)
	case "application/octet-stream":
		step, err := endpoints.NewKameletBuilder("cos-encoder-bytearray-action").Build()
		if err != nil {
			return errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Steps = append(binding.Spec.Steps, step)
	default:
		return fmt.Errorf("unsupported format %s", cc.DataShape.Produces.Format)

	}

	return nil
}
