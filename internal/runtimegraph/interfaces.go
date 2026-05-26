package runtimegraph

import (
	"context"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mergerhq/merger/internal/domain"
)

type Snapshot interface {
	Nodes(context.Context) ([]Node, error)
	Edges(context.Context) ([]Edge, error)
}

type ContentLoader interface {
	Load(context.Context, string) ([]byte, error)
}

type ResolutionInput struct {
	Packet  domain.ChangePacket
	Ref     string
	Loader  ContentLoader
	Options Options
}

type Options struct {
	EnableCodeOwners bool
}

type Resolver interface {
	ResolveImpact(context.Context, ResolutionInput) (domain.RuntimeImpact, []domain.OwnershipBoundary, error)
}

type Source interface {
	Name() string
	Collect(context.Context, ResolutionInput) (Fragment, error)
}

type Fragment struct {
	Nodes       []Node
	Edges       []Edge
	Systems     []domain.SystemRef
	Ownership   []domain.OwnershipBoundary
	Notes       []string
	Criticality domain.Criticality
}

type Builder struct {
	sources []Source
}

func NewResolver(options Options) *Builder {
	sources := []Source{
		repositoryTopologySource{},
		manifestSource{},
	}
	if options.EnableCodeOwners {
		sources = append(sources, codeOwnersSource{})
	}
	return &Builder{sources: sources}
}

func (r *Builder) ResolveImpact(ctx context.Context, input ResolutionInput) (domain.RuntimeImpact, []domain.OwnershipBoundary, error) {
	serviceIndex := make(map[string]domain.SystemRef)
	ownerIndex := make(map[string]domain.OwnershipBoundary)
	impact := domain.RuntimeImpact{
		BlastRadius: domain.BlastRadiusUnknown,
		Criticality: domain.CriticalityNormal,
	}

	for _, source := range r.sources {
		fragment, err := source.Collect(ctx, input)
		if err != nil {
			return domain.RuntimeImpact{}, nil, err
		}

		for _, system := range fragment.Systems {
			serviceIndex[system.Name] = system
		}
		for _, owner := range fragment.Ownership {
			ownerIndex[owner.Team] = owner
		}
		if criticalityRank(fragment.Criticality) > criticalityRank(impact.Criticality) {
			impact.Criticality = fragment.Criticality
		}
		impact.Notes = append(impact.Notes, fragment.Notes...)
	}

	switch len(serviceIndex) {
	case 0:
		impact.BlastRadius = domain.BlastRadiusUnknown
	case 1:
		impact.BlastRadius = domain.BlastRadiusIsolated
	default:
		impact.BlastRadius = domain.BlastRadiusLocalized
	}

	for _, system := range serviceIndex {
		impact.Services = append(impact.Services, system)
	}

	var boundaries []domain.OwnershipBoundary
	for _, owner := range ownerIndex {
		boundaries = append(boundaries, owner)
	}

	if len(impact.Notes) == 0 {
		impact.Notes = []string{"runtime graph sources did not find additional topology metadata"}
	}

	return impact, boundaries, nil
}

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

type manifestSource struct{}

func (manifestSource) Name() string { return "manifest-source" }

func (manifestSource) Collect(ctx context.Context, input ResolutionInput) (Fragment, error) {
	fragment := Fragment{}

	for _, file := range input.Packet.Files {
		ext := strings.ToLower(filepath.Ext(file.Path))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		if input.Loader == nil {
			continue
		}

		content, err := input.Loader.Load(ctx, file.Path)
		if err != nil || len(content) == 0 {
			continue
		}

		var parsed map[string]any
		if err := yaml.Unmarshal(content, &parsed); err != nil {
			continue
		}

		kind, _ := parsed["kind"].(string)
		metadata, _ := parsed["metadata"].(map[string]any)
		name, _ := metadata["name"].(string)
		if kind == "" || name == "" {
			continue
		}

		lowerKind := strings.ToLower(kind)
		if lowerKind == "service" || lowerKind == "deployment" || lowerKind == "statefulset" {
			fragment.Systems = append(fragment.Systems, domain.SystemRef{
				Kind:        domain.SystemService,
				Name:        name,
				Owner:       name,
				Criticality: domain.CriticalityNormal,
			})
			fragment.Notes = append(fragment.Notes, "manifest source extracted runtime objects from deploy manifests")
		}
	}

	return fragment, nil
}

type codeOwnersSource struct{}

func (codeOwnersSource) Name() string { return "codeowners-source" }

func (codeOwnersSource) Collect(ctx context.Context, input ResolutionInput) (Fragment, error) {
	if input.Loader == nil {
		return Fragment{}, nil
	}

	content, err := input.Loader.Load(ctx, ".github/CODEOWNERS")
	if err != nil || len(content) == 0 {
		content, _ = input.Loader.Load(ctx, "CODEOWNERS")
	}
	if len(content) == 0 {
		return Fragment{}, nil
	}

	lines := strings.Split(string(content), "\n")
	matches := make(map[string]domain.OwnershipBoundary)
	for _, file := range input.Packet.Files {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			pattern := strings.Trim(fields[0], "/")
			if strings.Contains(file.Path, strings.Trim(pattern, "*")) {
				team := strings.TrimPrefix(fields[1], "@")
				matches[team] = domain.OwnershipBoundary{
					Domain:   pattern,
					Team:     team,
					Systems:  []string{file.Path},
					Critical: false,
				}
			}
		}
	}

	fragment := Fragment{}
	for _, owner := range matches {
		fragment.Ownership = append(fragment.Ownership, owner)
	}
	if len(fragment.Ownership) > 0 {
		fragment.Notes = append(fragment.Notes, "codeowners source mapped changed files to owner teams")
	}

	return fragment, nil
}

func criticalityRank(value domain.Criticality) int {
	switch value {
	case domain.CriticalityTier0:
		return 4
	case domain.CriticalityHigh:
		return 3
	case domain.CriticalityNormal:
		return 2
	default:
		return 1
	}
}
