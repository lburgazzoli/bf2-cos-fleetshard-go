package camel

import (
	"encoding/base64"
	"fmt"
	"github.com/stoewer/go-strcase"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

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
