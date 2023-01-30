package configmaps

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/decoder"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"sort"
)

func ComputeDigest(resource corev1.ConfigMap) (string, error) {
	hash := sha256.New()

	if _, err := hash.Write([]byte(resource.Namespace)); err != nil {
		return "", err
	}
	if _, err := hash.Write([]byte(resource.Name)); err != nil {
		return "", err
	}

	keys := make([]string, 0, len(resource.Data))

	for k := range resource.Data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		if _, err := hash.Write([]byte(k)); err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte(resource.Data[k])); err != nil {
			return "", err
		}
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

func ExtractStructuredData[T any](resource corev1.ConfigMap, key string, target *T) error {
	data, ok := resource.Data[key]
	if !ok {
		return nil
	}

	err := json.Unmarshal([]byte(data), target)
	if err != nil {
		return errors.Wrap(err, "unable to decode content")
	}

	return nil
}

func SetStructuredData(resource *corev1.ConfigMap, key string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return errors.Wrap(err, "unable to marshal content")
	}

	if resource.Data == nil {
		resource.Data = make(map[string]string)
	}

	resource.Data[key] = string(data)

	return nil
}

// Decode decodes configmap's data using mapstructure
// TODO: add constraints to T
func Decode[T any](resource corev1.ConfigMap) (T, error) {
	var result T

	if resource.Data == nil {
		return result, nil
	}

	cfg := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			decoder.StringToURLHookFunc(),
			decoder.BytesToURLHookFunc(),
		),
	}

	decoder, err := mapstructure.NewDecoder(&cfg)
	if err != nil {
		return result, err
	}

	if err := decoder.Decode(resource.Data); err != nil {
		return result, err
	}

	return result, nil
}
