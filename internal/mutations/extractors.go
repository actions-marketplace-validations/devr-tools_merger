package mutations

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/devr-tools/merger/internal/domain"
)

type goPatchSignalExtractor struct{}
type sqlSignalExtractor struct{}
type yamlSignalExtractor struct{}
type manifestSignalExtractor struct{}

func DefaultExtractors() []SignalExtractor {
	return []SignalExtractor{
		goPatchSignalExtractor{},
		sqlSignalExtractor{},
		yamlSignalExtractor{},
		manifestSignalExtractor{},
	}
}

func (goPatchSignalExtractor) Name() string { return "go-patch-signals" }

func (goPatchSignalExtractor) Supports(file domain.ChangedFile) bool {
	return strings.EqualFold(filepath.Ext(file.Path), ".go")
}

func (goPatchSignalExtractor) Extract(_ context.Context, file domain.ChangedFile) ([]domain.MutationSignal, error) {
	patch := strings.ToLower(file.Patch)
	signals := make([]domain.MutationSignal, 0)

	if strings.Contains(patch, "func ") {
		signals = append(signals, domain.MutationSignal{Source: "go_patch", Value: "go:function_change", Weight: 1})
	}
	if strings.Contains(patch, "jwt") || strings.Contains(patch, "token") || strings.Contains(patch, "authorize") {
		signals = append(signals, domain.MutationSignal{Source: "go_patch", Value: "go:auth_symbol", Weight: 2})
	}

	return signals, nil
}

func (sqlSignalExtractor) Name() string { return "sql-signals" }

func (sqlSignalExtractor) Supports(file domain.ChangedFile) bool {
	return strings.EqualFold(filepath.Ext(file.Path), ".sql")
}

func (sqlSignalExtractor) Extract(_ context.Context, file domain.ChangedFile) ([]domain.MutationSignal, error) {
	patch := strings.ToLower(file.Patch)
	if strings.Contains(patch, "alter table") || strings.Contains(patch, "create table") || strings.Contains(patch, "drop table") {
		return []domain.MutationSignal{{Source: "sql_patch", Value: "sql:ddl", Weight: 3}}, nil
	}
	return nil, nil
}

func (yamlSignalExtractor) Name() string { return "yaml-signals" }

func (yamlSignalExtractor) Supports(file domain.ChangedFile) bool {
	ext := strings.ToLower(filepath.Ext(file.Path))
	return ext == ".yaml" || ext == ".yml"
}

func (yamlSignalExtractor) Extract(_ context.Context, file domain.ChangedFile) ([]domain.MutationSignal, error) {
	patch := strings.ToLower(file.Patch)
	signals := make([]domain.MutationSignal, 0)

	for _, key := range []string{"replicas:", "resources:", "image:", "env:", "ingress:", "rollout:"} {
		if strings.Contains(patch, key) {
			signals = append(signals, domain.MutationSignal{Source: "yaml_patch", Value: "yaml:runtime_key", Weight: 2})
			break
		}
	}

	if strings.Contains(strings.ToLower(file.Path), "openapi") {
		signals = append(signals, domain.MutationSignal{Source: "yaml_path", Value: "api:contract_surface", Weight: 3})
	}

	return signals, nil
}

func (manifestSignalExtractor) Name() string { return "manifest-signals" }

func (manifestSignalExtractor) Supports(file domain.ChangedFile) bool {
	base := strings.ToLower(filepath.Base(file.Path))
	switch base {
	case "go.mod", "go.sum", "package.json", "package-lock.json":
		return true
	default:
		return false
	}
}

func (manifestSignalExtractor) Extract(_ context.Context, file domain.ChangedFile) ([]domain.MutationSignal, error) {
	return []domain.MutationSignal{{Source: "manifest", Value: "manifest:dependency", Weight: 2}}, nil
}
