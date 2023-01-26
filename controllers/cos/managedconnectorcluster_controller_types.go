package cos

import "github.com/mitchellh/mapstructure"

type AddonParameters struct {
	BaseURL      string `mapstructure:"control-plane-base-url"`
	AuthURL      string `mapstructure:"mas-sso-base-url"`
	AuthRealm    string `mapstructure:"mas-sso-realm"`
	ClientID     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
	ClusterID    string `mapstructure:"cluster-id"`
}

func DecodeAddonsParams(in interface{}) (AddonParameters, error) {
	var params AddonParameters

	err := mapstructure.Decode(in, &params)
	if err != nil {
		return params, err
	}

	return params, nil
}
