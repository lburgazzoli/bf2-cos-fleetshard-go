package fleetmanager

import (
	"crypto/tls"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/logger"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
)

type Config struct {
	AccessToken  string
	ApiURL       *url.URL
	AuthURL      *url.URL
	UserAgent    string
	ClientID     string
	ClientSecret string
}

type Client struct {
	api *controlplane.APIClient
}

func NewClient(config *Config) (*Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: config.AccessToken,
		},
	)

	c := controlplane.NewConfiguration()
	c.Scheme = config.ApiURL.Scheme
	c.Host = config.ApiURL.Host
	c.UserAgent = config.UserAgent
	c.Debug = false
	c.HTTPClient = &http.Client{
		Transport: &oauth2.Transport{
			Base:   createTransport(config),
			Source: oauth2.ReuseTokenSource(nil, ts),
		},
	}

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
