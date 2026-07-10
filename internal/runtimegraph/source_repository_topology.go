package runtimegraph

import (
	"context"
	"strings"

	"github.com/devr-tools/merger/internal/domain"
)

type repositoryTopologySource struct{}

func (repositoryTopologySource) Name() string { return "repository-topology" }

func (repositoryTopologySource) Collect(_ context.Context, input ResolutionInput) (Fragment, error) {
	services := make(map[string]domain.SystemRef)
	owners := make(map[string]domain.OwnershipBoundary)

	for _, file := range input.Packet.Files {
		parts := strings.Split(strings.Trim(file.Path, "/"), "/")
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "cmd", "internal", "pkg", "services":
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

	fragment := Fragment{
		Notes: []string{"repository topology source derived impacted services from changed paths"},
	}
	for _, service := range services {
		fragment.Systems = append(fragment.Systems, service)
	}
	for _, owner := range owners {
		fragment.Ownership = append(fragment.Ownership, owner)
	}
	return fragment, nil
}
