package camel

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"

	camelv1lapha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/stoewer/go-strcase"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
)

func extractSecrets(
	config map[string]interface{},
	secret *corev1.Secret) error {

	for k, v := range config {
		if k == "data_shape" || k == "error_handler" {
			continue
		}

		switch t := v.(type) {
		case map[string]interface{}:
			kind := t["kind"]
			value := t["value"]

			if kind != "base64" {
				return fmt.Errorf("unsupported kind: %s", kind)
			}

			encoded, ok := value.(string)
			if !ok {
				return fmt.Errorf("error decoding key %s", k)
			}

			decoded, err := base64.StdEncoding.DecodeString(encoded)
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

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

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

func setTrait(target *camelv1lapha1.KameletBinding, key string, vals ...string) error {
	if len(vals) == 0 {
		return nil
	}

	if !strings.HasPrefix("trait.camel.apache.org/", key) {
		key = "trait.camel.apache.org/" + key
	}

	if len(vals) == 1 {
		resources.SetAnnotation(&target.ObjectMeta, key, vals[0])
	} else {
		data, err := json.Marshal(vals)
		if err != nil {
			return err
		}

		resources.SetAnnotation(&target.ObjectMeta, key, string(data))
	}

	return nil
}

func computeTraitsDigest(resource camelv1lapha1.KameletBinding) (string, error) {
	hash := sha256.New()

	if _, err := hash.Write([]byte(resource.Namespace)); err != nil {
		return "", err
	}
	if _, err := hash.Write([]byte(resource.Name)); err != nil {
		return "", err
	}

	keys := make([]string, 0, len(resource.Annotations))

	for k := range resource.Annotations {
		if !strings.HasPrefix(k, "trait.camel.apache.org/") {
			continue
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := resource.Annotations[k]

		if _, err := hash.Write([]byte(k)); err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte(v)); err != nil {
			return "", err
		}
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}
