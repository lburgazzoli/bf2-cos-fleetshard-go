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

type Client struct {
	api *controlplane.APIClient
}

func NewClient(ctx context.Context, config *Config) (*Client, error) {
	ts := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   config.AuthURL.String(),
			TokenURL:  config.AuthTokenURL.Scheme,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	c := controlplane.NewConfiguration()
	c.Scheme = config.ApiURL.Scheme
	c.Host = config.ApiURL.Host
	c.UserAgent = config.UserAgent
	c.Debug = false
	c.HTTPClient = ts.Client(ctx, nil)

	s := Client{
		api: controlplane.NewAPIClient(c),
	}

	return &s, nil
}

func createTransport(config *Config) (transport http.RoundTripper) {
	// Create the raw transport:
	// #nosec 402
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            nil,
		},
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: false,
	}

	return logger.LoggingRoundTripper{
		Proxied: transport,
	}
}
