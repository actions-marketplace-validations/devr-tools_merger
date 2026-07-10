package mutations

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/pkg/identity"
)

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
