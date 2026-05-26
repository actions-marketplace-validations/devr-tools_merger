package runtimegraph

import (
	"context"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mergerhq/merger/internal/domain"
)

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

		system, ok := parseManifestSystem(content)
		if !ok {
			continue
		}

		fragment.Systems = append(fragment.Systems, system)
		fragment.Notes = append(fragment.Notes, "manifest source extracted runtime objects from deploy manifests")
	}

	return fragment, nil
}

func parseManifestSystem(content []byte) (domain.SystemRef, bool) {
	var parsed map[string]any
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return domain.SystemRef{}, false
	}

	kind, _ := parsed["kind"].(string)
	metadata, _ := parsed["metadata"].(map[string]any)
	name, _ := metadata["name"].(string)
	if kind == "" || name == "" {
		return domain.SystemRef{}, false
	}

	lowerKind := strings.ToLower(kind)
	if lowerKind != "service" && lowerKind != "deployment" && lowerKind != "statefulset" {
		return domain.SystemRef{}, false
	}

	return domain.SystemRef{
		Kind:        domain.SystemService,
		Name:        name,
		Owner:       name,
		Criticality: domain.CriticalityNormal,
	}, true
}
