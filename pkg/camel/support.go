package camel

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/conditions"
	meta2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/meta"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strconv"
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

	if !strings.HasPrefix(TraitGroup+"/", key) {
		key = TraitGroup + "/" + key
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
		if !strings.HasPrefix(k, TraitGroup+"/") {
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

func ExtractConditions(conditions *[]metav1.Condition, binding camelv1lapha1.KameletBinding) error {

	var gen int64
	var err error

	rev := binding.Annotations[meta2.MetaDeploymentRevision]
	if rev != "" {
		gen, err = strconv.ParseInt(rev, 10, 64)
		if err != nil {
			return errors.Wrap(err, "unable to determine revision")
		}
	}

	// TODO: conditions must be filtered out
	for i := range binding.Status.Conditions {
		c := binding.Status.Conditions[i]

		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               "Workload" + string(c.Type),
			Status:             metav1.ConditionStatus(c.Status),
			LastTransitionTime: c.LastTransitionTime,
			Reason:             c.Reason,
			Message:            c.Message,

			// use ObservedGeneration to reference the deployment revision the
			// condition is about
			ObservedGeneration: gen,
		})
	}

	if len(binding.Status.Conditions) == 0 {
		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:    "WorkloadReady",
			Status:  metav1.ConditionFalse,
			Reason:  "Unknown",
			Message: "Unknown",

			// use ObservedGeneration to reference the deployment revision the
			// condition is about
			ObservedGeneration: gen,
		})
	}

	return nil
}

func ReadyCondition(connector cos.ManagedConnector) metav1.Condition {
	ready := metav1.Condition{
		Type:               conditions.ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		Reason:             conditions.ConditionReasonUnknown,
		Message:            conditions.ConditionMessageUnknown,
		ObservedGeneration: connector.Spec.Deployment.DeploymentResourceVersion,
	}

	if connector.Generation != connector.Status.ObservedGeneration {
		ready.Reason = conditions.ConditionMessageProvisioning
		ready.Message = conditions.ConditionReasonProvisioning
	}

	return ready
}

func SetReadyCondition(connector *cos.ManagedConnector, status metav1.ConditionStatus, reason string, message string) {
	controller.UpdateStatusCondition(&connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *metav1.Condition) {
		condition.Status = status
		condition.Reason = reason
		condition.Message = message
	})
}
