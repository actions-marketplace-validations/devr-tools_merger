package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type PullRequest struct {
	Owner   string
	Repo    string
	Number  int
	Title   string
	Body    string
	Author  string
	URL     string
	HeadSHA string
	BaseSHA string
}

type PullRequestFetcher interface {
	GetPullRequest(context.Context, string, string, int) (PullRequest, error)
}

type DiffFetcher interface {
	GetPullRequestDiff(context.Context, string, string, int) (string, error)
}

type FileContentFetcher interface {
	GetFileContent(context.Context, string, string, string, string) ([]byte, error)
}

type CheckRunInput struct {
	RepoOwner  string
	RepoName   string
	HeadSHA    string
	Name       string
	Status     string
	Conclusion string
	Summary    string
	DetailsURL string
}

type CheckRunPublisher interface {
	PublishCheckRun(context.Context, CheckRunInput) error
}

type Service interface {
	PullRequestFetcher
	DiffFetcher
	FileContentFetcher
	CheckRunPublisher
}

type InstallationBinder interface {
	ForInstallation(int64) Service
}

type ClientConfig struct {
	AppID          string
	InstallationID int64
	PrivateKeyPath string
	WebhookSecret  string
	APIBaseURL     string
	Timeout        string
}

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

	client := &Client{
		baseURL:        strings.TrimRight(cfg.APIBaseURL, "/"),
		httpClient:     &http.Client{Timeout: timeout},
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

func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (PullRequest, error) {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return PullRequest{}, err
	}

	body, err := c.getWithBearer(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), token, "application/vnd.github+json")
	if err != nil {
		return PullRequest{}, err
	}

	var response struct {
		HTMLURL string `json:"html_url"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		Head    struct {
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			SHA string `json:"sha"`
		} `json:"base"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return PullRequest{}, err
	}

	return PullRequest{
		Owner:   owner,
		Repo:    repo,
		Number:  number,
		Title:   response.Title,
		Body:    response.Body,
		Author:  response.User.Login,
		URL:     response.HTMLURL,
		HeadSHA: response.Head.SHA,
		BaseSHA: response.Base.SHA,
	}, nil
}

func (c *Client) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return "", err
	}

	body, err := c.getWithBearer(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), token, "application/vnd.github.diff")
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) GetFileContent(ctx context.Context, owner, repo, filePath, ref string) ([]byte, error) {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s", owner, repo, path.Clean(filePath), url.QueryEscape(ref))
	body, err := c.getWithBearer(ctx, endpoint, token, "application/vnd.github+json")
	if err != nil {
		return nil, err
	}

	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !strings.EqualFold(response.Encoding, "base64") {
		return nil, fmt.Errorf("unsupported GitHub content encoding: %s", response.Encoding)
	}

	normalized := strings.ReplaceAll(response.Content, "\n", "")
	return base64.StdEncoding.DecodeString(normalized)
}

func (c *Client) PublishCheckRun(ctx context.Context, input CheckRunInput) error {
	token, err := c.authenticator.Token(ctx, c.installationID)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"name":       input.Name,
		"head_sha":   input.HeadSHA,
		"status":     input.Status,
		"conclusion": input.Conclusion,
		"output": map[string]string{
			"title":   input.Name,
			"summary": input.Summary,
		},
	}
	if input.DetailsURL != "" {
		payload["details_url"] = input.DetailsURL
	}

	_, err = c.postWithBearer(ctx, fmt.Sprintf("/repos/%s/%s/check-runs", input.RepoOwner, input.RepoName), token, payload)
	return err
}

func (c *Client) getWithBearer(ctx context.Context, endpoint, bearerToken, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return c.do(req)
}

func (c *Client) postWithBearer(ctx context.Context, endpoint, bearerToken string, payload any) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.do(req)
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("github api %s %s failed with %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(body))
	}

	return body, nil
}

type NoopService struct{}

func (NoopService) GetPullRequest(_ context.Context, _, _ string, _ int) (PullRequest, error) {
	return PullRequest{}, errors.New("github pull request fetch not implemented")
}

func (NoopService) GetPullRequestDiff(_ context.Context, _, _ string, _ int) (string, error) {
	return "", errors.New("github diff fetch not implemented")
}

func (NoopService) GetFileContent(_ context.Context, _, _, _, _ string) ([]byte, error) {
	return nil, errors.New("github file content fetch not implemented")
}

func (NoopService) PublishCheckRun(_ context.Context, _ CheckRunInput) error {
	return nil
}
