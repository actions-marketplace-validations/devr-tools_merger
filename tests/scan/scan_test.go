package scan_test

import (
	"context"
	"strings"
	"testing"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/lanes"
	"github.com/devr-tools/merger/internal/policy"
	"github.com/devr-tools/merger/internal/scan"
)

const authDiff = `diff --git a/internal/auth/session.go b/internal/auth/session.go
index 3333333..4444444 100644
--- a/internal/auth/session.go
+++ b/internal/auth/session.go
@@ -10,6 +10,9 @@ func Authenticate(token string) bool {
-	return legacyCheck(token)
+	if token == "" {
+		return false
+	}
+	return verify(token)
 }
`

func defaultLanes() lanes.Config {
	return lanes.Config{GreenMax: 20, YellowMax: 55, RedMax: 85}
}

func authPolicy() policy.Config {
	return policy.Config{Policies: []policy.RuleConfig{{
		Name: "auth_requires_security_review",
		When: policy.WhenClause{Mutations: []domain.MutationKind{domain.MutationAuthBehaviorChange}},
		Require: policy.RequirementClause{
			Reviewers:  []string{"security"},
			Evidence:   []string{"auth_integration_tests"},
			Deployment: policy.DeploymentClause{Strategy: "canary", RequiresCanary: true},
		},
		Action: policy.ActionClause{MinimumLane: domain.MergeLaneRed},
	}}}
}

func TestRunClassifiesAuthMutationAndEscalatesLane(t *testing.T) {
	packet, err := scan.Run(context.Background(), scan.Options{
		Diff:   authDiff,
		Repo:   domain.RepoRef{Owner: "acme", Name: "payments", FullName: "acme/payments"},
		Policy: authPolicy(),
		Lanes:  defaultLanes(),
	})
	if err != nil {
		t.Fatalf("scan.Run: %v", err)
	}

	if len(packet.Files) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(packet.Files))
	}

	if !hasMutationKind(packet.Mutations, domain.MutationAuthBehaviorChange) {
		t.Fatalf("expected an auth_behavior_change mutation, got %+v", packet.Mutations)
	}

	if packet.MergeLane != domain.MergeLaneRed {
		t.Fatalf("expected RED lane after auth policy escalation, got %s", packet.MergeLane)
	}

	if !hasReviewer(packet.Reviewers, "security") {
		t.Fatalf("expected a security reviewer requirement, got %+v", packet.Reviewers)
	}
}

func TestRunWithEmptyDiffProducesNoMutations(t *testing.T) {
	packet, err := scan.Run(context.Background(), scan.Options{
		Diff:  "",
		Lanes: defaultLanes(),
	})
	if err != nil {
		t.Fatalf("scan.Run: %v", err)
	}
	if len(packet.Files) != 0 {
		t.Fatalf("expected no files, got %d", len(packet.Files))
	}
	if len(packet.Mutations) != 0 {
		t.Fatalf("expected no mutations, got %d", len(packet.Mutations))
	}
	if packet.MergeLane != domain.MergeLaneGreen {
		t.Fatalf("expected GREEN lane for an empty change, got %s", packet.MergeLane)
	}
}

func TestRunAssignsIDAndSource(t *testing.T) {
	packet, err := scan.Run(context.Background(), scan.Options{Diff: authDiff, Lanes: defaultLanes()})
	if err != nil {
		t.Fatalf("scan.Run: %v", err)
	}
	if !strings.HasPrefix(packet.ID, "cp_") {
		t.Fatalf("expected change packet id prefixed with cp_, got %q", packet.ID)
	}
	if packet.Source != "cli.scan" {
		t.Fatalf("expected source cli.scan, got %q", packet.Source)
	}
}

func hasMutationKind(mutations []domain.Mutation, kind domain.MutationKind) bool {
	for _, mutation := range mutations {
		if mutation.Kind == kind {
			return true
		}
	}
	return false
}

func hasReviewer(reviewers []domain.ReviewerRequirement, team string) bool {
	for _, reviewer := range reviewers {
		if reviewer.Team == team {
			return true
		}
	}
	return false
}
