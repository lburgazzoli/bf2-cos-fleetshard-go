package fleetmanager

import (
	"context"
	"crypto/tls"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/url"
)

type Config struct {
	ApiURL       *url.URL
	AuthURL      *url.URL
	AuthTokenURL *url.URL
	UserAgent    string
	ClientID     string
	ClientSecret string
	ClusterID    string
	Debug        bool
}

func NewClient(ctx context.Context, config Config) (Client, error) {
	t := logger.LoggingRoundTripper{
		Proxied: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            nil,
			},
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: false,
		},
	}

	oauthConfig := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     config.AuthTokenURL.String(),
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}

	ct := context.WithValue(ctx, oauth2.HTTPClient, http.Client{Transport: t})

	apiConfig := controlplane.NewConfiguration()
	apiConfig.Scheme = config.ApiURL.Scheme
	apiConfig.Host = config.ApiURL.Host
	apiConfig.UserAgent = config.UserAgent
	apiConfig.Debug = config.Debug
	apiConfig.HTTPClient = oauthConfig.Client(ct)

	client := defaultClient{
		api:       controlplane.NewAPIClient(apiConfig),
		clusterId: config.ClusterID,
	}

	return &client, nil
}
