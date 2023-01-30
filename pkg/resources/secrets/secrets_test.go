package secrets

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

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Data: map[string][]byte{
			"control-plane-base-url": []byte("https://api.acme.com"),
			"sso-base-url":           []byte("https://sso.acme.com"),
			"client-id":              []byte("foo"),
			"client-secret":          []byte("bar"),
		},
	}

	params, err := Decode[AddonParameters](secret)

	assert.Nil(t, err)

	assert.NotNil(t, params.BaseURL)
	assert.NotNil(t, params.AuthURL)

	assert.Equal(t, "https://api.acme.com", params.BaseURL.String())
	assert.Equal(t, "https://sso.acme.com", params.AuthURL.String())
	assert.Equal(t, "foo", params.ClientID)
	assert.Equal(t, "bar", params.ClientSecret)
}
