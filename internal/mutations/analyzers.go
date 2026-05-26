package mutations

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/pkg/identity"
)

type goASTAnalyzer struct{}
type sqlDDLAnalyzer struct{}
type openAPIAnalyzer struct{}
type manifestAnalyzer struct{}
type runtimeConfigAnalyzer struct{}

func DefaultAnalyzers() []Analyzer {
	return []Analyzer{
		goASTAnalyzer{},
		sqlDDLAnalyzer{},
		openAPIAnalyzer{},
		manifestAnalyzer{},
		runtimeConfigAnalyzer{},
	}
}

func (goASTAnalyzer) Name() string { return "go-ast-analyzer" }

func (goASTAnalyzer) Supports(file domain.ChangedFile) bool {
	return strings.EqualFold(filepath.Ext(file.Path), ".go")
}

func (goASTAnalyzer) Analyze(_ context.Context, input AnalysisInput) ([]domain.Mutation, error) {
	if len(input.Content) == 0 {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, input.File.Path, input.Content, parser.ParseComments)
	if err != nil {
		return nil, nil
	}

	var authSignals []domain.MutationSignal
	var dataSignals []domain.MutationSignal

	for _, imp := range parsed.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		lower := strings.ToLower(importPath)
		switch {
		case strings.Contains(lower, "auth"), strings.Contains(lower, "jwt"), strings.Contains(lower, "oauth"):
			authSignals = append(authSignals, domain.MutationSignal{Source: "go_ast_import", Value: "go:auth_symbol", Weight: 3})
		case strings.Contains(lower, "database/sql"), strings.Contains(lower, "pgx"), strings.Contains(lower, "sqlx"):
			dataSignals = append(dataSignals, domain.MutationSignal{Source: "go_ast_import", Value: "go:data_access", Weight: 2})
		}
	}

	ast.Inspect(parsed, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.FuncDecl:
			name := strings.ToLower(typed.Name.Name)
			if containsAny(name, "auth", "token", "jwt", "authorize", "permission") {
				authSignals = append(authSignals, domain.MutationSignal{Source: "go_ast_function", Value: "go:auth_symbol", Weight: 3})
			}
		case *ast.SelectorExpr:
			name := strings.ToLower(typed.Sel.Name)
			if containsAny(name, "query", "exec", "scan", "transaction", "begin") {
				dataSignals = append(dataSignals, domain.MutationSignal{Source: "go_ast_selector", Value: "go:data_access", Weight: 2})
			}
		}
		return true
	})

	var mutations []domain.Mutation
	if len(authSignals) > 0 {
		mutations = append(mutations, domain.Mutation{
			ID:          identity.New("mut"),
			Kind:        domain.MutationAuthBehaviorChange,
			Severity:    domain.SeverityHigh,
			Confidence:  0.92,
			Title:       "authentication or authorization behavior change",
			Description: "Go AST analysis detected auth-sensitive code paths or imports.",
			Files:       []string{input.File.Path},
			Signals:     authSignals,
			Detector:    "go-ast-analyzer",
		})
	}
	if len(dataSignals) > 0 {
		mutations = append(mutations, domain.Mutation{
			ID:          identity.New("mut"),
			Kind:        domain.MutationDataAccess,
			Severity:    domain.SeverityMedium,
			Confidence:  0.76,
			Title:       "data access behavior change",
			Description: "Go AST analysis detected database interaction surfaces.",
			Files:       []string{input.File.Path},
			Signals:     dataSignals,
			Detector:    "go-ast-analyzer",
		})
	}

	return mutations, nil
}

func (sqlDDLAnalyzer) Name() string { return "sql-ddl-analyzer" }

func (sqlDDLAnalyzer) Supports(file domain.ChangedFile) bool {
	return strings.EqualFold(filepath.Ext(file.Path), ".sql")
}

func (sqlDDLAnalyzer) Analyze(_ context.Context, input AnalysisInput) ([]domain.Mutation, error) {
	content := strings.ToLower(string(bytes.TrimSpace(input.Content)))
	if !containsAny(content, "alter table", "create table", "drop table", "create index", "drop index") {
		return nil, nil
	}

	return []domain.Mutation{{
		ID:          identity.New("mut"),
		Kind:        domain.MutationDatabaseSchema,
		Severity:    domain.SeverityHigh,
		Confidence:  0.96,
		Title:       "database schema mutation",
		Description: "DDL statements were detected in SQL content.",
		Files:       []string{input.File.Path},
		Signals:     []domain.MutationSignal{{Source: "sql_ast", Value: "sql:ddl", Weight: 4}},
		Detector:    "sql-ddl-analyzer",
	}}, nil
}

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

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func flattenKeys(values map[string]any, prefix string) []string {
	keys := make([]string, 0)
	for key, value := range values {
		full := key
		if prefix != "" {
			full = prefix + "." + key
		}
		keys = append(keys, full)

		switch typed := value.(type) {
		case map[string]any:
			keys = append(keys, flattenKeys(typed, full)...)
		}
	}
	return keys
}
