package risk

import "github.com/devr-tools/merger/internal/domain"

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

func clampScore(score int) int {
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
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
