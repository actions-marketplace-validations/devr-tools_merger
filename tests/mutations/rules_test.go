package mutations_test

import (
	"testing"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/mutations"
)

func TestDefaultRulesMatchOpenAPIBySignal(t *testing.T) {
	rules := mutations.DefaultRules()

	var matched bool
	for _, rule := range rules {
		if rule.Kind != domain.MutationAPIContract {
			continue
		}
		matched = rule.Matches("docs/spec.yaml", []domain.MutationSignal{{
			Source: "openapi",
			Value:  "api:contract_surface",
			Weight: 4,
		}})
		break
	}

	if !matched {
		t.Fatal("expected api contract rule to match by signal")
	}
}
