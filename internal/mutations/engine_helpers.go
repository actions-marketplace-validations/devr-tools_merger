package mutations

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/devr-tools/merger/internal/domain"
)

func (e *RuleBasedEngine) extractSignals(ctx context.Context, file domain.ChangedFile) ([]domain.MutationSignal, error) {
	signals := make([]domain.MutationSignal, 0, 1)
	signals = append(signals, domain.MutationSignal{
		Source: "path",
		Value:  strings.ToLower(filepath.Base(file.Path)),
		Weight: 1,
	})

	for _, extractor := range e.extractors {
		if !extractor.Supports(file) {
			continue
		}

		extracted, err := extractor.Extract(ctx, file)
		if err != nil {
			return nil, err
		}
		signals = append(signals, extracted...)
	}

	return signals, nil
}

func loadContent(ctx context.Context, loader ContentLoader, path string) ([]byte, error) {
	if loader == nil {
		return nil, nil
	}
	return loader.Load(ctx, path)
}
