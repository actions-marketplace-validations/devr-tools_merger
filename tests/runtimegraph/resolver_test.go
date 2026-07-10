package runtimegraph_test

import (
	"context"
	"testing"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/runtimegraph"
)

func TestResolverDerivesOwnershipFromTopology(t *testing.T) {
	resolver := runtimegraph.NewResolver(runtimegraph.Options{})

	impact, owners, err := resolver.ResolveImpact(context.Background(), runtimegraph.ResolutionInput{
		Packet: domain.ChangePacket{
			Files: []domain.ChangedFile{
				{Path: "internal/auth/jwt.go"},
			},
		},
	})
	if err != nil {
		t.Fatalf("resolve impact: %v", err)
	}
	if impact.BlastRadius != domain.BlastRadiusIsolated {
		t.Fatalf("expected isolated blast radius, got %s", impact.BlastRadius)
	}
	if len(owners) == 0 || owners[0].Team != "auth" {
		t.Fatalf("expected auth owner, got %#v", owners)
	}
}

func TestResolverIncludesManifestDerivedService(t *testing.T) {
	resolver := runtimegraph.NewResolver(runtimegraph.Options{})
	loader := stubLoader{
		files: map[string][]byte{
			"deploy/api.yaml": []byte("kind: Deployment\nmetadata:\n  name: payments-api\n"),
		},
	}

	impact, owners, err := resolver.ResolveImpact(context.Background(), runtimegraph.ResolutionInput{
		Packet: domain.ChangePacket{
			Files: []domain.ChangedFile{
				{Path: "deploy/api.yaml"},
			},
		},
		Loader: loader,
	})
	if err != nil {
		t.Fatalf("resolve impact: %v", err)
	}
	if len(impact.Services) != 1 || impact.Services[0].Name != "payments-api" {
		t.Fatalf("expected manifest-derived service, got %#v", impact.Services)
	}
	if len(owners) != 0 {
		t.Fatalf("expected no owners from manifest-only change, got %#v", owners)
	}
}

func TestResolverMapsCodeOwnersWhenEnabled(t *testing.T) {
	resolver := runtimegraph.NewResolver(runtimegraph.Options{EnableCodeOwners: true})
	loader := stubLoader{
		files: map[string][]byte{
			".github/CODEOWNERS": []byte("/internal/auth/ @security\n"),
		},
	}

	_, owners, err := resolver.ResolveImpact(context.Background(), runtimegraph.ResolutionInput{
		Packet: domain.ChangePacket{
			Files: []domain.ChangedFile{
				{Path: "internal/auth/jwt.go"},
			},
		},
		Loader: loader,
	})
	if err != nil {
		t.Fatalf("resolve impact: %v", err)
	}
	if len(owners) != 2 {
		t.Fatalf("expected topology and codeowners boundaries, got %#v", owners)
	}
}
