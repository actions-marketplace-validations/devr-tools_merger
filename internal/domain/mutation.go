package domain

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type MutationKind string

const (
	MutationUnknown               MutationKind = "unknown_mutation"
	MutationAuthBehaviorChange    MutationKind = "auth_behavior_change"
	MutationDatabaseSchema        MutationKind = "database_schema_mutation"
	MutationRuntimeConfig         MutationKind = "runtime_config_mutation"
	MutationAPIContract           MutationKind = "api_contract_mutation"
	MutationDependency            MutationKind = "dependency_mutation"
	MutationInfrastructure        MutationKind = "infrastructure_mutation"
	MutationDataAccess            MutationKind = "data_access_mutation"
	MutationDeploymentWorkflow    MutationKind = "deployment_workflow_mutation"
	MutationObservabilityContract MutationKind = "observability_contract_mutation"
)

type MutationSignal struct {
	Source string `json:"source"`
	Value  string `json:"value"`
	Weight int    `json:"weight"`
}

type Mutation struct {
	ID          string            `json:"id"`
	Kind        MutationKind      `json:"kind"`
	Severity    Severity          `json:"severity"`
	Confidence  float64           `json:"confidence"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Files       []string          `json:"files,omitempty"`
	Signals     []MutationSignal  `json:"signals,omitempty"`
	Detector    string            `json:"detector,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
