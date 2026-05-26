package domain

type ReviewerRequirement struct {
	Team         string `json:"team"`
	Reason       string `json:"reason,omitempty"`
	Mandatory    bool   `json:"mandatory"`
	MaxReviewers int    `json:"maxReviewers,omitempty"`
}

type DeploymentStrategy string

const (
	DeployDirect         DeploymentStrategy = "direct"
	DeployCanary         DeploymentStrategy = "canary"
	DeployPhased         DeploymentStrategy = "phased"
	DeployManualApproval DeploymentStrategy = "manual_approval"
)

type DeploymentRequirement struct {
	Strategy             DeploymentStrategy `json:"strategy"`
	Environments         []string           `json:"environments,omitempty"`
	RequiresCanary       bool               `json:"requiresCanary"`
	RequiresRollbackPlan bool               `json:"requiresRollbackPlan"`
	FreezeWindow         string             `json:"freezeWindow,omitempty"`
	Owners               []string           `json:"owners,omitempty"`
}
