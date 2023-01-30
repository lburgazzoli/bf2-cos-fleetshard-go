package cos

import (
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
)

type AddonParameters struct {
	BaseURL      string `mapstructure:"control-plane-base-url"`
	AuthURL      string `mapstructure:"mas-sso-base-url"`
	AuthRealm    string `mapstructure:"mas-sso-realm"`
	ClientID     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
	ClusterID    string `mapstructure:"cluster-id"`
}

type Cluster struct {
	fleetmanager.Client

	Parameters AddonParameters
}
