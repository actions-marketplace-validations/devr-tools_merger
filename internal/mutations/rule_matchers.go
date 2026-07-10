package mutations

import (
	"path"
	"strings"

	"github.com/devr-tools/merger/internal/domain"
)

func (r Rule) Matches(filePath string, signals []domain.MutationSignal) bool {
	lowerPath := strings.ToLower(filePath)

	if containsAnyPrefix(lowerPath, r.PathPrefixes) || containsAnySubstring(lowerPath, r.PathContains) ||
		hasAnySuffix(lowerPath, r.PathSuffixes) || matchesAnyGlob(lowerPath, r.FileGlobs) {
		return true
	}

	if len(r.RequiredSignals) == 0 {
		return false
	}

	for _, signal := range signals {
		for _, required := range r.RequiredSignals {
			if strings.EqualFold(signal.Value, required) {
				return true
			}
		}
	}

	return false
}

func containsAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

func containsAnySubstring(value string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(value, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func hasAnySuffix(value string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, strings.ToLower(suffix)) {
			return true
		}
	}
	return false
}

func matchesAnyGlob(value string, globs []string) bool {
	for _, glob := range globs {
		normalized := strings.TrimPrefix(strings.ToLower(glob), "**/")
		ok, _ := path.Match(normalized, value)
		if ok {
			return true
		}
		if strings.HasSuffix(value, strings.TrimPrefix(normalized, "*")) {
			return true
		}
	}
	return false
}
