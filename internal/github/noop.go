package github

import (
	"context"
	"errors"
)

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
