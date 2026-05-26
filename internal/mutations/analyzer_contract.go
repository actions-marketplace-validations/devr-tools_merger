package mutations

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/pkg/identity"
)

func (openAPIAnalyzer) Name() string { return "openapi-analyzer" }

func (openAPIAnalyzer) Supports(file domain.ChangedFile) bool {
	lower := strings.ToLower(file.Path)
	return strings.HasSuffix(lower, ".proto") || strings.HasSuffix(lower, "openapi.yaml") || strings.HasSuffix(lower, "openapi.yml")
}

func (openAPIAnalyzer) Analyze(_ context.Context, input AnalysisInput) ([]domain.Mutation, error) {
	if strings.HasSuffix(strings.ToLower(input.File.Path), ".proto") {
		return []domain.Mutation{{
			ID:          identity.New("mut"),
			Kind:        domain.MutationAPIContract,
			Severity:    domain.SeverityHigh,
			Confidence:  0.91,
			Title:       "API contract mutation",
			Description: "Protocol buffer definition changed.",
			Files:       []string{input.File.Path},
			Signals:     []domain.MutationSignal{{Source: "proto", Value: "api:contract_surface", Weight: 3}},
			Detector:    "openapi-analyzer",
		}}, nil
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(input.Content, &parsed); err != nil {
		return nil, nil
	}
	if _, ok := parsed["openapi"]; !ok {
		return nil, nil
	}

	return []domain.Mutation{{
		ID:          identity.New("mut"),
		Kind:        domain.MutationAPIContract,
		Severity:    domain.SeverityHigh,
		Confidence:  0.93,
		Title:       "API contract mutation",
		Description: "OpenAPI schema changed.",
		Files:       []string{input.File.Path},
		Signals:     []domain.MutationSignal{{Source: "openapi", Value: "api:contract_surface", Weight: 4}},
		Detector:    "openapi-analyzer",
	}}, nil
}
