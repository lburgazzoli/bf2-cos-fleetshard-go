package secrets

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

func ComputeDigest(resource corev1.Secret) (string, error) {
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
		if _, err := hash.Write(resource.Data[k]); err != nil {
			return "", err
		}
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

func ExtractStructuredData[T any](resource corev1.Secret, key string, target *T) error {
	if resource.Data == nil {
		return nil
	}

	data, ok := resource.Data[key]
	if !ok {
		return nil
	}

	if err := json.Unmarshal(data, target); err != nil {
		return errors.Wrap(err, "unable to extract content")
	}

	return nil
}

// Decode decodes secret's data using mapstructure
func Decode[T any](resource corev1.Secret) (T, error) {
	var result T

	if resource.Data == nil {
		return result, nil
	}

	cfg := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &result,
		DecodeHook:       decoder.StringToURLHookFunc(),
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

func SetStructuredData(resource *corev1.Secret, key string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return errors.Wrap(err, "unable to marshal content")
	}

	if resource.Data == nil {
		resource.Data = make(map[string][]byte)
	}

	resource.Data[key] = data

	return nil
}
