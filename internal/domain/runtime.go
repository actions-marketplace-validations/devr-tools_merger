package domain

type BlastRadius string

const (
	BlastRadiusUnknown   BlastRadius = "unknown"
	BlastRadiusIsolated  BlastRadius = "isolated"
	BlastRadiusLocalized BlastRadius = "localized"
	BlastRadiusSystemic  BlastRadius = "systemic"
)

type Criticality string

const (
	CriticalityLow    Criticality = "low"
	CriticalityNormal Criticality = "normal"
	CriticalityHigh   Criticality = "high"
	CriticalityTier0  Criticality = "tier0"
)

type SystemKind string

const (
	SystemService  SystemKind = "service"
	SystemAPI      SystemKind = "api"
	SystemDatabase SystemKind = "database"
	SystemQueue    SystemKind = "queue"
	SystemInfra    SystemKind = "infra"
)

type SystemRef struct {
	Kind        SystemKind  `json:"kind"`
	Name        string      `json:"name"`
	Namespace   string      `json:"namespace,omitempty"`
	Owner       string      `json:"owner,omitempty"`
	Criticality Criticality `json:"criticality,omitempty"`
}

type RuntimeImpact struct {
	BlastRadius BlastRadius `json:"blastRadius"`
	Criticality Criticality `json:"criticality"`
	Services    []SystemRef `json:"services,omitempty"`
	APIs        []SystemRef `json:"apis,omitempty"`
	Datastores  []SystemRef `json:"datastores,omitempty"`
	Queues      []SystemRef `json:"queues,omitempty"`
	Notes       []string    `json:"notes,omitempty"`
}

type OwnershipBoundary struct {
	Domain     string   `json:"domain"`
	Team       string   `json:"team"`
	Systems    []string `json:"systems,omitempty"`
	Escalation string   `json:"escalation,omitempty"`
	Critical   bool     `json:"critical"`
}
