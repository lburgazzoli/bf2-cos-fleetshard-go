package configmaps

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func ComputeDigest(resource corev1.ConfigMap) (string, error) {
	hash := sha256.New()

	if _, err := hash.Write([]byte(resource.Namespace)); err != nil {
		return "", err
	}
	if _, err := hash.Write([]byte(resource.Name)); err != nil {
		return "", err
	}

	for k, v := range resource.Data {
		if _, err := hash.Write([]byte(k)); err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte(v)); err != nil {
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
