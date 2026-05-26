package lanes_test

import (
	"context"
	"testing"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/lanes"
)

func TestAssignerReturnsGreenForLowRiskAutomatableChanges(t *testing.T) {
	assigner := lanes.NewAssigner(lanes.Config{GreenMax: 20, YellowMax: 55, RedMax: 85})

	lane, err := assigner.Assign(context.Background(), domain.ChangePacket{
		RiskSummary: domain.RiskSummary{Score: 10},
		Decision: domain.PolicyDecision{
			Status: domain.DecisionApproved,
		},
		Deployment: domain.DeploymentRequirement{
			Strategy: domain.DeployDirect,
		},
	})
	if err != nil {
		t.Fatalf("assign lane: %v", err)
	}
	if lane != domain.MergeLaneGreen {
		t.Fatalf("expected GREEN, got %s", lane)
	}
}
