package mutations

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/pkg/identity"
)

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
