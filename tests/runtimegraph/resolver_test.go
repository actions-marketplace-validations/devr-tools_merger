package runtimegraph_test

import (
	"context"
	"testing"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/runtimegraph"
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
