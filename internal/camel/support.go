package camel

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	conditions2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/conditions"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	kamelv1lapha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
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

type TraitVal interface {
	string | []string
}

func setTrait[T TraitVal](target *kamelv1lapha1.KameletBinding, key string, val T) error {
	if !strings.HasPrefix(TraitGroup+"/", key) {
		key = TraitGroup + "/" + key
	}

	switch v := any(val).(type) {
	case string:
		resources.SetAnnotation(&target.ObjectMeta, key, v)
	case []string:
		if len(v) == 0 {
			return nil
		}

		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		resources.SetAnnotation(&target.ObjectMeta, key, string(data))
	default:
		return nil
	}

	return nil
}

func computeTraitsDigest(resource kamelv1lapha1.KameletBinding) (string, error) {
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

func extractConditions(conditions *[]cosv2.Condition, binding kamelv1lapha1.KameletBinding) error {

	var gen int64
	var err error

	rev := binding.Annotations[cosmeta.MetaDeploymentRevision]
	if rev != "" {
		gen, err = strconv.ParseInt(rev, 10, 64)
		if err != nil {
			return errors.Wrap(err, "unable to determine revision")
		}
	}

	// TODO: conditions must be filtered out
	for i := range binding.Status.Conditions {
		c := binding.Status.Conditions[i]

		wc := cosv2.Condition{
			Condition: metav1.Condition{
				Type:               "Workload" + string(c.Type),
				Status:             metav1.ConditionStatus(c.Status),
				LastTransitionTime: c.LastTransitionTime,
				Reason:             c.Reason,
				Message:            c.Message,
			},
			ResourceRevision: gen,
		}

		if len(wc.Reason) == 0 {
			wc.Reason = "Unknown"
		}
		if len(wc.Message) == 0 {
			wc.Message = "Unknown"
		}

		conditions2.Set(conditions, wc)
	}

	if len(binding.Status.Conditions) == 0 {
		conditions2.Set(conditions, cosv2.Condition{
			Condition: metav1.Condition{
				Type:    "WorkloadReady",
				Status:  metav1.ConditionFalse,
				Reason:  "Unknown",
				Message: "Unknown",
			},
			ResourceRevision: gen,
		})
	}

	return nil
}

func patchDependant(rc controller.ReconciliationContext, source client.Object, target client.Object) error {

	if err := controllerutil.SetControllerReference(rc.Connector, target, rc.M.GetScheme()); err != nil {
		return errors.Wrapf(err, "unable to set binding config controller to: %s", target.GetObjectKind().GroupVersionKind().String())
	}

	ok, err := rc.PatchDependant(source, target)
	if err != nil {
		return errors.Wrapf(err, "unable to patch %s", target.GetObjectKind().GroupVersionKind().String())
	}
	if ok {
		patchDependantCount.With(prometheus.Labels{
			"connector_id":   rc.Connector.Spec.ConnectorID,
			"dependant_name": target.GetName(),
			"dependant_kind": target.GetObjectKind().GroupVersionKind().String(),
		})
	}

	return nil
}
