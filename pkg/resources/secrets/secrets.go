package secrets

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func ComputeDigest(resource corev1.Secret) (string, error) {
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
		if _, err := hash.Write(v); err != nil {
			return "", err
		}
	}

	// Add a letter at the beginning and use URL safe encoding
	digest := "v" + base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

	return digest, nil
}

func ExtractStructuredData[T any](resource corev1.Secret, key string, target *T) error {
	data, ok := resource.Data[key]
	if !ok {
		return nil
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		return errors.Wrap(err, "unable to decode content")
	}

	return nil
}
