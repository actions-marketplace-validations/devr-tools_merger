package github

import (
	"context"
	"errors"
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
	CheckRunPublisher
}

type NoopService struct{}

func (NoopService) GetPullRequest(_ context.Context, _, _ string, _ int) (PullRequest, error) {
	return PullRequest{}, errors.New("github pull request fetch not implemented")
}

func (NoopService) GetPullRequestDiff(_ context.Context, _, _ string, _ int) (string, error) {
	return "", errors.New("github diff fetch not implemented")
}

func (NoopService) PublishCheckRun(_ context.Context, _ CheckRunInput) error {
	return nil
}
