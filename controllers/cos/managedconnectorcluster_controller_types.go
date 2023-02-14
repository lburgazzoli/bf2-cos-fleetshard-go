package cos

import (
	v2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
	"net/url"
	"time"
)

type AddonParameters struct {
	BaseURL      *url.URL `mapstructure:"control-plane-base-url"`
	AuthURL      *url.URL `mapstructure:"mas-sso-base-url"`
	AuthRealm    string   `mapstructure:"mas-sso-realm"`
	ClientID     string   `mapstructure:"client-id"`
	ClientSecret string   `mapstructure:"client-secret"`
	ClusterID    string   `mapstructure:"cluster-id"`
}

type Cluster struct {
	fleetmanager.Client

	MCC        v2.ManagedConnectorCluster
	Parameters AddonParameters

	ResyncDelay time.Duration
	ResyncAt    time.Time
}
