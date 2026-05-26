package mutations

import (
	"path"
	"strings"

	"github.com/mergerhq/merger/internal/domain"
)

type Rule struct {
	Name            string
	Title           string
	Description     string
	Kind            domain.MutationKind
	Severity        domain.Severity
	Confidence      float64
	PathPrefixes    []string
	PathContains    []string
	PathSuffixes    []string
	FileGlobs       []string
	RequiredSignals []string
}

func DefaultRules() []Rule {
	return []Rule{
		{
			Name:         "auth-surface",
			Title:        "authentication or authorization behavior change",
			Description:  "Changes touch credential, token, policy, or auth control paths.",
			Kind:         domain.MutationAuthBehaviorChange,
			Severity:     domain.SeverityHigh,
			Confidence:   0.85,
			PathContains: []string{"auth/", "jwt", "oauth", "rbac", "acl"},
			RequiredSignals: []string{
				"go:auth_symbol",
			},
		},
		{
			Name:         "schema-ddl",
			Title:        "database schema mutation",
			Description:  "DDL or migration assets changed.",
			Kind:         domain.MutationDatabaseSchema,
			Severity:     domain.SeverityHigh,
			Confidence:   0.9,
			PathPrefixes: []string{"migrations/", "db/migrations/"},
			PathSuffixes: []string{".sql"},
			RequiredSignals: []string{
				"sql:ddl",
			},
		},
		{
			Name:         "runtime-config",
			Title:        "runtime configuration mutation",
			Description:  "Runtime or deploy-time configuration changed.",
			Kind:         domain.MutationRuntimeConfig,
			Severity:     domain.SeverityMedium,
			Confidence:   0.8,
			PathContains: []string{"helm/", "k8s/", "deploy/", "config/"},
			FileGlobs:    []string{"**/values.yaml", "**/values.yml"},
			RequiredSignals: []string{
				"yaml:runtime_key",
			},
		},
		{
			Name:            "api-contract",
			Title:           "API contract mutation",
			Description:     "Public or internal API definition changed.",
			Kind:            domain.MutationAPIContract,
			Severity:        domain.SeverityHigh,
			Confidence:      0.88,
			PathSuffixes:    []string{"openapi.yaml", "openapi.yml", ".proto"},
			FileGlobs:       []string{"**/openapi.yaml", "**/openapi.yml", "**/*.proto"},
			RequiredSignals: []string{"api:contract_surface"},
		},
		{
			Name:         "dependency",
			Title:        "dependency graph mutation",
			Description:  "Module or dependency manifest changed.",
			Kind:         domain.MutationDependency,
			Severity:     domain.SeverityMedium,
			Confidence:   0.92,
			PathSuffixes: []string{"go.mod", "go.sum", "package-lock.json", "package.json"},
		},
		{
			Name:         "deployment-workflow",
			Title:        "deployment workflow mutation",
			Description:  "CI/CD or rollout workflow changed.",
			Kind:         domain.MutationDeploymentWorkflow,
			Severity:     domain.SeverityHigh,
			Confidence:   0.82,
			PathContains: []string{".github/workflows/", "deployments/"},
		},
		{
			Name:         "observability-surface",
			Title:        "observability contract mutation",
			Description:  "Telemetry, alerting, or dashboard behavior changed.",
			Kind:         domain.MutationObservabilityContract,
			Severity:     domain.SeverityMedium,
			Confidence:   0.7,
			PathContains: []string{"telemetry/", "observability/", "alerts/", "dashboards/"},
		},
	}
}

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
