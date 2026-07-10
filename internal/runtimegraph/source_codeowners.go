package runtimegraph

import (
	"context"
	"strings"

	"github.com/devr-tools/merger/internal/domain"
)

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

	owners := collectCodeOwners(string(content), input.Packet.Files)
	fragment := Fragment{Ownership: owners}
	if len(fragment.Ownership) > 0 {
		fragment.Notes = append(fragment.Notes, "codeowners source mapped changed files to owner teams")
	}

	return fragment, nil
}

func collectCodeOwners(codeOwners string, files []domain.ChangedFile) []domain.OwnershipBoundary {
	lines := strings.Split(codeOwners, "\n")
	matches := make(map[string]domain.OwnershipBoundary)

	for _, file := range files {
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
			if !strings.Contains(file.Path, strings.Trim(pattern, "*")) {
				continue
			}

			team := strings.TrimPrefix(fields[1], "@")
			matches[team] = domain.OwnershipBoundary{
				Domain:   pattern,
				Team:     team,
				Systems:  []string{file.Path},
				Critical: false,
			}
		}
	}

	owners := make([]domain.OwnershipBoundary, 0, len(matches))
	for _, owner := range matches {
		owners = append(owners, owner)
	}
	return owners
}
