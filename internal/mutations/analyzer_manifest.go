package mutations

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/pkg/identity"
)

func (manifestAnalyzer) Name() string { return "manifest-analyzer" }

func (manifestAnalyzer) Supports(file domain.ChangedFile) bool {
	switch strings.ToLower(filepath.Base(file.Path)) {
	case "go.mod", "go.sum", "package.json", "package-lock.json":
		return true
	default:
		return false
	}
}

func (manifestAnalyzer) Analyze(_ context.Context, input AnalysisInput) ([]domain.Mutation, error) {
	return []domain.Mutation{{
		ID:          identity.New("mut"),
		Kind:        domain.MutationDependency,
		Severity:    domain.SeverityMedium,
		Confidence:  0.97,
		Title:       "dependency graph mutation",
		Description: "Dependency manifest changed.",
		Files:       []string{input.File.Path},
		Signals:     []domain.MutationSignal{{Source: "manifest", Value: "manifest:dependency", Weight: 3}},
		Detector:    "manifest-analyzer",
	}}, nil
}
