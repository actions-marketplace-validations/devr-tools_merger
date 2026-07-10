package mutations

import (
	"context"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/pkg/identity"
)

func (runtimeConfigAnalyzer) Name() string { return "runtime-config-analyzer" }

func (runtimeConfigAnalyzer) Supports(file domain.ChangedFile) bool {
	ext := strings.ToLower(filepath.Ext(file.Path))
	if ext != ".yaml" && ext != ".yml" {
		return false
	}
	lower := strings.ToLower(file.Path)
	return containsAny(lower, "helm/", "k8s/", "deploy/", "config/")
}

func (runtimeConfigAnalyzer) Analyze(_ context.Context, input AnalysisInput) ([]domain.Mutation, error) {
	var parsed map[string]any
	if err := yaml.Unmarshal(input.Content, &parsed); err != nil {
		return nil, nil
	}

	keys := flattenKeys(parsed, "")
	for _, key := range keys {
		if containsAny(strings.ToLower(key), "replicas", "resources", "image", "env", "ingress", "rollout") {
			return []domain.Mutation{{
				ID:          identity.New("mut"),
				Kind:        domain.MutationRuntimeConfig,
				Severity:    domain.SeverityMedium,
				Confidence:  0.86,
				Title:       "runtime configuration mutation",
				Description: "Structured config analysis detected deploy/runtime keys.",
				Files:       []string{input.File.Path},
				Signals:     []domain.MutationSignal{{Source: "yaml_structured", Value: "yaml:runtime_key", Weight: 3}},
				Detector:    "runtime-config-analyzer",
			}}, nil
		}
	}

	return nil, nil
}
