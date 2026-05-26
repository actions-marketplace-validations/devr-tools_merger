package mutations

import "github.com/mergerhq/merger/internal/domain"

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
			Title:        "runtime or deploy-time configuration mutation",
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
