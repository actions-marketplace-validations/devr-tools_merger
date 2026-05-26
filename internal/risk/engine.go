package risk

import (
	"context"
	"strings"

	"github.com/mergerhq/merger/internal/domain"
)

type Engine interface {
	Evaluate(context.Context, domain.ChangePacket) (domain.RiskSummary, []domain.Risk, error)
}

type WeightedEngine struct {
	weights map[domain.MutationKind]int
}

func DefaultEngine() *WeightedEngine {
	return &WeightedEngine{
		weights: map[domain.MutationKind]int{
			domain.MutationAuthBehaviorChange:    35,
			domain.MutationDatabaseSchema:        32,
			domain.MutationRuntimeConfig:         20,
			domain.MutationAPIContract:           26,
			domain.MutationDependency:            15,
			domain.MutationInfrastructure:        28,
			domain.MutationDeploymentWorkflow:    24,
			domain.MutationObservabilityContract: 12,
			domain.MutationUnknown:               8,
		},
	}
}

func (e *WeightedEngine) Evaluate(_ context.Context, packet domain.ChangePacket) (domain.RiskSummary, []domain.Risk, error) {
	riskByType := make(map[domain.RiskType]*domain.Risk)
	score := 0

	for _, mutation := range packet.Mutations {
		base := e.weights[mutation.Kind]
		riskType, summary, mitigations := classifyRisk(mutation.Kind)
		score += base

		current := riskByType[riskType]
		if current == nil {
			current = &domain.Risk{
				Type:        riskType,
				Summary:     summary,
				Mitigations: mitigations,
				Severity:    mutation.Severity,
			}
			riskByType[riskType] = current
		}

		current.Score += base
		current.Reason = strings.TrimSpace(current.Reason + "; mutation=" + string(mutation.Kind))
		current.AffectedSystems = appendUnique(current.AffectedSystems, mutation.Files...)
	}

	switch packet.Runtime.Criticality {
	case domain.CriticalityHigh:
		score += 10
	case domain.CriticalityTier0:
		score += 20
	}

	if len(packet.Files) > 20 {
		score += 8
	}

	if score > 100 {
		score = 100
	}

	risks := make([]domain.Risk, 0, len(riskByType))
	contributors := make([]domain.RiskType, 0, len(riskByType))
	for _, risk := range riskByType {
		risk.Severity = severityFromScore(risk.Score)
		risks = append(risks, *risk)
		contributors = append(contributors, risk.Type)
	}

	return domain.RiskSummary{
		Score:        score,
		Severity:     severityFromScore(score),
		Contributors: contributors,
	}, risks, nil
}

func classifyRisk(kind domain.MutationKind) (domain.RiskType, string, []string) {
	switch kind {
	case domain.MutationAuthBehaviorChange:
		return domain.RiskSecurity, "security control surface changed", []string{"security review", "auth integration tests"}
	case domain.MutationDatabaseSchema:
		return domain.RiskSchema, "schema compatibility may affect runtime behavior", []string{"migration plan", "rollback validation"}
	case domain.MutationRuntimeConfig, domain.MutationDeploymentWorkflow:
		return domain.RiskRollout, "rollout behavior changed", []string{"canary", "runtime smoke tests"}
	case domain.MutationDependency:
		return domain.RiskDependency, "dependency graph changed", []string{"dependency diff review", "build provenance"}
	default:
		return domain.RiskRuntime, "runtime behavior may have shifted", []string{"integration tests"}
	}
}

func severityFromScore(score int) domain.Severity {
	switch {
	case score >= 80:
		return domain.SeverityCritical
	case score >= 55:
		return domain.SeverityHigh
	case score >= 25:
		return domain.SeverityMedium
	default:
		return domain.SeverityLow
	}
}

func appendUnique(values []string, candidates ...string) []string {
	for _, candidate := range candidates {
		found := false
		for _, value := range values {
			if value == candidate {
				found = true
				break
			}
		}
		if !found {
			values = append(values, candidate)
		}
	}
	return values
}
