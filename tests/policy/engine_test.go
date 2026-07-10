package policy_test

import (
	"context"
	"testing"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/policy"
)

func TestPolicyRequiresSecurityReviewForAuthMutation(t *testing.T) {
	engine := policy.NewRuleEngine(policy.Config{
		Policies: []policy.RuleConfig{
			{
				Name: "auth_requires_security_review",
				When: policy.WhenClause{
					Mutations: []domain.MutationKind{domain.MutationAuthBehaviorChange},
				},
				Require: policy.RequirementClause{
					Reviewers: []string{"security"},
					Evidence:  []string{string(domain.EvidenceAuthTests)},
				},
				Action: policy.ActionClause{
					MinimumLane: domain.MergeLaneRed,
				},
			},
		},
	})

	result, err := engine.Evaluate(context.Background(), domain.ChangePacket{
		Mutations: []domain.Mutation{
			{Kind: domain.MutationAuthBehaviorChange},
		},
	})
	if err != nil {
		t.Fatalf("evaluate policy: %v", err)
	}
	if len(result.Reviewers) != 1 || result.Reviewers[0].Team != "security" {
		t.Fatalf("unexpected reviewers: %#v", result.Reviewers)
	}
	if result.Decision.MinimumLane != domain.MergeLaneRed {
		t.Fatalf("expected RED minimum lane, got %s", result.Decision.MinimumLane)
	}
}
