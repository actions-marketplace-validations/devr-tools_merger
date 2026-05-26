package github

import (
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL        string
	httpClient     *http.Client
	authenticator  *AppAuthenticator
	installationID int64
	webhookSecret  string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	timeout := 10 * time.Second
	if cfg.Timeout != "" {
		parsed, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, err
		}
		timeout = parsed
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	client := &Client{
		baseURL:        strings.TrimRight(cfg.APIBaseURL, "/"),
		httpClient:     httpClient,
		installationID: cfg.InstallationID,
		webhookSecret:  cfg.WebhookSecret,
	}

	authenticator, err := NewAppAuthenticator(cfg.AppID, cfg.PrivateKeyPath, client)
	if err != nil {
		return nil, err
	}
	client.authenticator = authenticator

	return client, nil
}

func (c *Client) ForInstallation(installationID int64) Service {
	clone := *c
	clone.installationID = installationID
	return &clone
}
