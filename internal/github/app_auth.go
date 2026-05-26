package github

import "context"

type InstallationTokenSource interface {
	Token(context.Context, int64) (string, error)
}

type StaticTokenSource struct {
	Value string
}

func (s StaticTokenSource) Token(_ context.Context, _ int64) (string, error) {
	return s.Value, nil
}
