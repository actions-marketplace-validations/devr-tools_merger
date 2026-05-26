package runtimegraph_test

import (
	"context"
	"fmt"
)

type stubLoader struct {
	files map[string][]byte
}

func (s stubLoader) Load(_ context.Context, path string) ([]byte, error) {
	content, ok := s.files[path]
	if !ok {
		return nil, fmt.Errorf("missing file %q", path)
	}
	return content, nil
}
