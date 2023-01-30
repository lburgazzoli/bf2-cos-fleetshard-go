package configmaps

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"testing"
)

func TestDecodeURL(t *testing.T) {
	type AddonParameters struct {
		BaseURL      *url.URL `mapstructure:"control-plane-base-url"`
		AuthURL      *url.URL `mapstructure:"sso-base-url"`
		ClientID     string   `mapstructure:"client-id"`
		ClientSecret string   `mapstructure:"client-secret"`
	}

	config := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Data: map[string]string{
			"control-plane-base-url": "https://api.acme.com",
			"sso-base-url":           "https://sso.acme.com",
			"client-id":              "foo",
			"client-secret":          "bar",
		},
	}

	params, err := Decode[AddonParameters](config)

	assert.Nil(t, err)

	assert.NotNil(t, params.BaseURL)
	assert.NotNil(t, params.AuthURL)

	assert.Equal(t, "https://api.acme.com", params.BaseURL.String())
	assert.Equal(t, "https://sso.acme.com", params.AuthURL.String())
	assert.Equal(t, "foo", params.ClientID)
	assert.Equal(t, "bar", params.ClientSecret)
}
