package policy

import "github.com/devr-tools/merger/internal/domain"

type Config struct {
	Policies []RuleConfig `yaml:"policies"`
}

type RuleConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	When        WhenClause        `yaml:"when"`
	Require     RequirementClause `yaml:"require"`
	Action      ActionClause      `yaml:"action"`
}

type WhenClause struct {
	Mutations      []domain.MutationKind `yaml:"mutations"`
	Paths          []string              `yaml:"paths"`
	Criticalities  []domain.Criticality  `yaml:"criticalities"`
	RiskScoreGTE   int                   `yaml:"risk_score_gte"`
	OwnershipTeams []string              `yaml:"ownership_teams"`
}

type RequirementClause struct {
	Reviewers  []string         `yaml:"reviewers"`
	Evidence   []string         `yaml:"evidence"`
	Deployment DeploymentClause `yaml:"deployment"`
}

type DeploymentClause struct {
	Strategy             string   `yaml:"strategy"`
	RequiresCanary       bool     `yaml:"requires_canary"`
	RequiresRollbackPlan bool     `yaml:"requires_rollback_plan"`
	Environments         []string `yaml:"environments"`
}

type ActionClause struct {
	Block       bool             `yaml:"block"`
	MinimumLane domain.MergeLane `yaml:"minimum_lane"`
}
