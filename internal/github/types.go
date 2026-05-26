package github

import (
	"context"
	"net/http"
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
	HTTPClient     *http.Client
}
