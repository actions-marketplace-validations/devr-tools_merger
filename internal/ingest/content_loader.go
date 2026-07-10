package ingest

import (
	"context"

	"github.com/devr-tools/merger/internal/github"
)

type githubContentLoader struct {
	service github.Service
	owner   string
	repo    string
	ref     string
}

func newGitHubContentLoader(service github.Service, owner, repo, ref string) githubContentLoader {
	return githubContentLoader{
		service: service,
		owner:   owner,
		repo:    repo,
		ref:     ref,
	}
}

func (l githubContentLoader) Load(ctx context.Context, path string) ([]byte, error) {
	return l.service.GetFileContent(ctx, l.owner, l.repo, path, l.ref)
}
