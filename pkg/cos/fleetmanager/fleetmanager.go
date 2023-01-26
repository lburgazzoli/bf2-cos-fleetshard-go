package fleetmanager

import (
	"context"
	"crypto/tls"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/logger"
	"golang.org/x/oauth2"
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

	oauthConfig := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   config.AuthURL.String(),
			TokenURL:  config.AuthTokenURL.String(),
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	apiConfig := controlplane.NewConfiguration()
	apiConfig.Scheme = config.ApiURL.Scheme
	apiConfig.Host = config.ApiURL.Host
	apiConfig.UserAgent = config.UserAgent
	apiConfig.Debug = false
	apiConfig.HTTPClient = oauthConfig.Client(context.WithValue(ctx, oauth2.HTTPClient, http.Client{Transport: t}), nil)

	client := defaultClient{
		api: controlplane.NewAPIClient(apiConfig),
	}

	return &client, nil
}
