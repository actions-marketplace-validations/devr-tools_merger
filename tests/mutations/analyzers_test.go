package mutations_test

import (
	"context"
	"testing"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/mutations"
)

type staticContentLoader map[string][]byte

func (l staticContentLoader) Load(_ context.Context, path string) ([]byte, error) {
	return l[path], nil
}

func TestDefaultEngineClassifiesAuthMutationFromAST(t *testing.T) {
	engine := mutations.DefaultEngine()

	results, err := engine.Classify(context.Background(), mutations.AnalysisRequest{
		Repo: domain.RepoRef{
			Owner:    "acme",
			Name:     "authz",
			FullName: "acme/authz",
		},
		Ref: "deadbeef",
		Files: []domain.ChangedFile{{
			Path:     "internal/auth/jwt.go",
			Status:   domain.FileModified,
			Language: "go",
		}},
		Content: staticContentLoader{
			"internal/auth/jwt.go": []byte("package auth\n\nimport \"github.com/golang-jwt/jwt/v5\"\n\nfunc AuthorizeToken() {}\n"),
		},
	})
	if err != nil {
		t.Fatalf("classify: %v", err)
	}

	if !hasMutationKind(results, domain.MutationAuthBehaviorChange) {
		t.Fatalf("expected auth behavior mutation, got %#v", results)
	}
}

func TestDefaultEngineClassifiesRuntimeConfigMutationFromStructuredYAML(t *testing.T) {
	engine := mutations.DefaultEngine()

	results, err := engine.Classify(context.Background(), mutations.AnalysisRequest{
		Repo: domain.RepoRef{
			Owner:    "acme",
			Name:     "deploy",
			FullName: "acme/deploy",
		},
		Ref: "deadbeef",
		Files: []domain.ChangedFile{{
			Path:     "helm/payments/values.yaml",
			Status:   domain.FileModified,
			Language: "yaml",
		}},
		Content: staticContentLoader{
			"helm/payments/values.yaml": []byte("deployment:\n  replicas: 3\n  image:\n    repository: payments\n"),
		},
	})
	if err != nil {
		t.Fatalf("classify: %v", err)
	}

	if !hasMutationKind(results, domain.MutationRuntimeConfig) {
		t.Fatalf("expected runtime config mutation, got %#v", results)
	}
}

func hasMutationKind(mutationsList []domain.Mutation, kind domain.MutationKind) bool {
	for _, mutation := range mutationsList {
		if mutation.Kind == kind {
			return true
		}
	}
	return false
}
