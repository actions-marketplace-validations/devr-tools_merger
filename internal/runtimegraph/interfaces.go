package runtimegraph

import (
	"context"
	"strings"

	"github.com/mergerhq/merger/internal/domain"
)

type Snapshot interface {
	Nodes(context.Context) ([]Node, error)
	Edges(context.Context) ([]Edge, error)
}

type Resolver interface {
	ResolveImpact(context.Context, domain.ChangePacket) (domain.RuntimeImpact, []domain.OwnershipBoundary, error)
}

type StaticResolver struct {
	nodes []Node
	edges []Edge
}

func NewStaticResolver(nodes []Node, edges []Edge) *StaticResolver {
	return &StaticResolver{nodes: nodes, edges: edges}
}

func (r *StaticResolver) ResolveImpact(_ context.Context, packet domain.ChangePacket) (domain.RuntimeImpact, []domain.OwnershipBoundary, error) {
	services := make(map[string]domain.SystemRef)
	owners := make(map[string]domain.OwnershipBoundary)

	for _, file := range packet.Files {
		parts := strings.Split(strings.Trim(file.Path, "/"), "/")
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "cmd", "internal", "pkg":
			if len(parts) > 1 {
				serviceName := parts[1]
				services[serviceName] = domain.SystemRef{
					Kind:        domain.SystemService,
					Name:        serviceName,
					Owner:       serviceName,
					Criticality: domain.CriticalityNormal,
				}
				owners[serviceName] = domain.OwnershipBoundary{
					Domain:   serviceName,
					Team:     serviceName,
					Systems:  []string{serviceName},
					Critical: false,
				}
			}
		}
	}

	impact := domain.RuntimeImpact{
		BlastRadius: domain.BlastRadiusUnknown,
		Criticality: domain.CriticalityNormal,
		Notes: []string{
			"Runtime graph resolver is scaffolded for Phase 1 and currently derives impact from repository topology.",
		},
	}

	switch len(services) {
	case 0:
		impact.BlastRadius = domain.BlastRadiusUnknown
	case 1:
		impact.BlastRadius = domain.BlastRadiusIsolated
	default:
		impact.BlastRadius = domain.BlastRadiusLocalized
	}

	for _, service := range services {
		impact.Services = append(impact.Services, service)
	}

	var boundaries []domain.OwnershipBoundary
	for _, boundary := range owners {
		boundaries = append(boundaries, boundary)
	}

	return impact, boundaries, nil
}
