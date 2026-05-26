package policy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mergerhq/merger/internal/domain"
)

type Evaluation struct {
	Decision   domain.PolicyDecision
	Evidence   []domain.EvidenceRequirement
	Reviewers  []domain.ReviewerRequirement
	Deployment domain.DeploymentRequirement
}

type Engine interface {
	Evaluate(context.Context, domain.ChangePacket) (Evaluation, error)
}

type RuleEngine struct {
	config Config
}

func NewRuleEngine(config Config) *RuleEngine {
	return &RuleEngine{config: config}
}

func (e *RuleEngine) Evaluate(_ context.Context, packet domain.ChangePacket) (Evaluation, error) {
	evaluation := Evaluation{
		Decision: domain.PolicyDecision{
			Status:    domain.DecisionApproved,
			DecidedAt: time.Now().UTC(),
		},
		Deployment: domain.DeploymentRequirement{
			Strategy: domain.DeployDirect,
		},
	}

	for _, rule := range e.config.Policies {
		if !matchesRule(rule, packet) {
			continue
		}

		evaluation.Decision.AppliedPolicies = append(evaluation.Decision.AppliedPolicies, rule.Name)
		if rule.Description != "" {
			evaluation.Decision.Reasons = append(evaluation.Decision.Reasons, rule.Description)
		}

		for _, reviewer := range rule.Require.Reviewers {
			evaluation.Reviewers = appendUniqueReviewer(evaluation.Reviewers, domain.ReviewerRequirement{
				Team:      reviewer,
				Reason:    fmt.Sprintf("required by policy %s", rule.Name),
				Mandatory: true,
			})
		}

		for _, evidence := range rule.Require.Evidence {
			evaluation.Evidence = appendUniqueEvidence(evaluation.Evidence, domain.EvidenceRequirement{
				Type:     domain.EvidenceType(evidence),
				Name:     evidence,
				Required: true,
				Reason:   fmt.Sprintf("required by policy %s", rule.Name),
				Producer: "policy-engine",
			})
		}

		if rule.Require.Deployment.Strategy != "" {
			evaluation.Deployment.Strategy = domain.DeploymentStrategy(rule.Require.Deployment.Strategy)
		}
		if len(rule.Require.Deployment.Environments) > 0 {
			evaluation.Deployment.Environments = append(evaluation.Deployment.Environments, rule.Require.Deployment.Environments...)
		}
		evaluation.Deployment.RequiresCanary = evaluation.Deployment.RequiresCanary || rule.Require.Deployment.RequiresCanary
		evaluation.Deployment.RequiresRollbackPlan = evaluation.Deployment.RequiresRollbackPlan || rule.Require.Deployment.RequiresRollbackPlan

		if rule.Action.MinimumLane != "" {
			evaluation.Decision.MinimumLane = maxLane(evaluation.Decision.MinimumLane, rule.Action.MinimumLane)
		}

		if rule.Action.Block {
			evaluation.Decision.Status = domain.DecisionBlocked
			evaluation.Decision.Violations = append(evaluation.Decision.Violations, domain.PolicyViolation{
				Policy:   rule.Name,
				Reason:   "policy blocked change propagation",
				Severity: domain.SeverityCritical,
			})
		} else if len(evaluation.Reviewers) > 0 || len(evaluation.Evidence) > 0 {
			evaluation.Decision.Status = domain.DecisionPending
		}
	}

	evaluation.Decision.Summary = strings.Join(evaluation.Decision.Reasons, "; ")
	if evaluation.Decision.Summary == "" {
		evaluation.Decision.Summary = "no blocking policy constraints"
	}

	return evaluation, nil
}

func matchesRule(rule RuleConfig, packet domain.ChangePacket) bool {
	if len(rule.When.Mutations) > 0 && !packetHasAnyMutation(packet, rule.When.Mutations) {
		return false
	}
	if len(rule.When.Paths) > 0 && !packetTouchesAnyPath(packet, rule.When.Paths) {
		return false
	}
	if len(rule.When.Criticalities) > 0 && !containsCriticality(rule.When.Criticalities, packet.Runtime.Criticality) {
		return false
	}
	if rule.When.RiskScoreGTE > 0 && packet.RiskSummary.Score < rule.When.RiskScoreGTE {
		return false
	}
	if len(rule.When.OwnershipTeams) > 0 && !packetTouchesOwners(packet, rule.When.OwnershipTeams) {
		return false
	}

	return true
}

func packetHasAnyMutation(packet domain.ChangePacket, kinds []domain.MutationKind) bool {
	for _, mutation := range packet.Mutations {
		for _, kind := range kinds {
			if mutation.Kind == kind {
				return true
			}
		}
	}
	return false
}

func packetTouchesAnyPath(packet domain.ChangePacket, patterns []string) bool {
	for _, file := range packet.Files {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(file.Path), strings.ToLower(pattern)) {
				return true
			}
		}
	}
	return false
}

func containsCriticality(values []domain.Criticality, candidate domain.Criticality) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func packetTouchesOwners(packet domain.ChangePacket, teams []string) bool {
	for _, owner := range packet.Ownership {
		for _, team := range teams {
			if strings.EqualFold(owner.Team, team) {
				return true
			}
		}
	}
	return false
}

func appendUniqueReviewer(values []domain.ReviewerRequirement, candidate domain.ReviewerRequirement) []domain.ReviewerRequirement {
	for _, value := range values {
		if strings.EqualFold(value.Team, candidate.Team) {
			return values
		}
	}
	return append(values, candidate)
}

func appendUniqueEvidence(values []domain.EvidenceRequirement, candidate domain.EvidenceRequirement) []domain.EvidenceRequirement {
	for _, value := range values {
		if value.Name == candidate.Name {
			return values
		}
	}
	return append(values, candidate)
}

func maxLane(a, b domain.MergeLane) domain.MergeLane {
	if a == "" {
		return b
	}
	order := map[domain.MergeLane]int{
		domain.MergeLaneGreen:  1,
		domain.MergeLaneYellow: 2,
		domain.MergeLaneRed:    3,
		domain.MergeLaneBlack:  4,
	}
	if order[b] > order[a] {
		return b
	}
	return a
}
