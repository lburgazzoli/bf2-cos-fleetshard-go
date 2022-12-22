package camel

import (
	"encoding/base64"
	"sort"

	kamelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	"github.com/apache/camel-k/pkg/apis/camel/v1/trait"
	kamelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"

	"github.com/pkg/errors"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/camel/endpoints"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/configmaps"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/secrets"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SecretEntryConnector       string = "connector"
	SecretEntryServiceAccount  string = "serviceAccount"
	SecretEntryMeta            string = "meta"
	ServiceAccountClientID     string = "sa_client_id"
	ServiceAccountClientSecret string = "sa_client_secret"
)

func reify(rc *controller.ReconciliationContext) (kamelv1alpha1.KameletBinding, corev1.Secret, corev1.ConfigMap, error) {

	binding := kamelv1alpha1.KameletBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Connector.Name,
			Namespace: rc.Connector.Namespace,
		},
	}
	bindingSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Connector.Name,
			Namespace: rc.Connector.Namespace,
		},
	}
	bindingConfig := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Connector.Name,
			Namespace: rc.Connector.Namespace,
		},
	}

	var sa ServiceAccount
	var meta ShardMetadata
	var config ConnectorConfiguration

	if err := secrets.ExtractStructuredData(*rc.Secret, SecretEntryServiceAccount, &sa); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding service account")
	}
	if err := secrets.ExtractStructuredData(*rc.Secret, SecretEntryMeta, &meta); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding shard meta")
	}
	if err := secrets.ExtractStructuredData(*rc.Secret, SecretEntryConnector, &config); err != nil {
		return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error decoding config")
	}

	// TODO: improve
	if config.DataShape.Consumes == nil {
		config.DataShape.Consumes = &DataShape{}
	}
	if config.DataShape.Consumes.Format == "" {
		config.DataShape.Consumes.Format = meta.Consumes
	}
	if config.DataShape.Produces == nil {
		config.DataShape.Produces = &DataShape{}
	}
	if config.DataShape.Produces.Format == "" {
		config.DataShape.Produces.Format = meta.Produces
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
		cosmeta.MetaOperatorType,
		cosmeta.MetaDeploymentID,
		cosmeta.MetaConnectorID,
		cosmeta.MetaConnectorTypeID)

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

	if err := extractSecrets(config.Properties, &bindingSecret); err != nil {
		return binding, bindingSecret, bindingConfig, err
	}
	if err := extractConfig(config.Properties, &bindingConfig); err != nil {
		return binding, bindingSecret, bindingConfig, err
	}

	switch meta.ConnectorType {
	case ConnectorTypeSource:
		src, err := endpoints.NewKameletBuilder(meta.Kamelets.Adapter.Name).
			Property("id", rc.Connector.Spec.ConnectorID+"-source").
			PropertiesFrom(config.Properties, meta.Kamelets.Adapter.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating source")
		}

		sink, err := endpoints.NewKameletBuilder(meta.Kamelets.Kafka.Name).
			Property("id", rc.Connector.Spec.ConnectorID+"-sink").
			Property("bootstrapServers", rc.Connector.Spec.Deployment.Kafka.URL).
			Property("valueSerializer", "org.bf2.cos.connector.camel.serdes.bytes.ByteArraySerializer").
			PropertyPlaceholder("user", ServiceAccountClientID).
			PropertyPlaceholder("password", ServiceAccountClientSecret).
			PropertiesFrom(config.Properties, meta.Kamelets.Kafka.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Source = src
		binding.Spec.Sink = sink

		break

	case ConnectorTypeSink:
		src, err := endpoints.NewKameletBuilder(meta.Kamelets.Kafka.Name).
			Property("id", rc.Connector.Spec.ConnectorID+"-source").
			Property("bootstrapServers", rc.Connector.Spec.Deployment.Kafka.URL).
			Property("consumerGroup", rc.Connector.Spec.ConnectorID).
			Property("valueDeserializer", "org.bf2.cos.connector.camel.serdes.bytes.ByteArrayDeserializer").
			PropertyPlaceholder("user", ServiceAccountClientID).
			PropertyPlaceholder("password", ServiceAccountClientSecret).
			PropertiesFrom(config.Properties, meta.Kamelets.Kafka.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating source")
		}

		sink, err := endpoints.NewKameletBuilder(meta.Kamelets.Adapter.Name).
			Property("id", rc.Connector.Spec.ConnectorID+"-sink").
			PropertiesFrom(config.Properties, meta.Kamelets.Adapter.Prefix).
			Build()

		if err != nil {
			return binding, bindingSecret, bindingConfig, errors.Wrap(err, "error creating sink")
		}

		binding.Spec.Source = src
		binding.Spec.Sink = sink

		break
	}

	if err := configureSteps(&binding, config); err != nil {
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

	binding.Spec.Integration = &kamelv1.IntegrationSpec{
		Profile: kamelv1.TraitProfileOpenShift,
		Traits: kamelv1.Traits{
			Environment: &trait.EnvironmentTrait{
				Vars: []string{
					"CONNECTOR_ID=" + rc.Connector.Spec.ConnectorID,
					"CONNECTOR_DEPLOYMENT_ID=" + rc.Connector.Spec.DeploymentID,
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
