package mutations_test

import (
	"context"
	"testing"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/mutations"
)

func TestDefaultEngineClassifiesSchemaMutation(t *testing.T) {
	engine := mutations.DefaultEngine()

	results, err := engine.Classify(context.Background(), mutations.AnalysisRequest{
		Repo: domain.RepoRef{
			Owner:    "acme",
			Name:     "payments",
			FullName: "acme/payments",
		},
		Ref: "deadbeef",
		Files: []domain.ChangedFile{{
			Path:     "migrations/20260525_add_users.sql",
			Status:   domain.FileAdded,
			Patch:    "+ALTER TABLE users ADD COLUMN timezone TEXT;",
			Language: "sql",
		}},
	})
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one mutation")
	}
	if results[0].Kind != domain.MutationDatabaseSchema {
		t.Fatalf("expected database schema mutation, got %s", results[0].Kind)
	}
}
