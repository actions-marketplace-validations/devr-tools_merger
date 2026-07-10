package runtimegraph

import "github.com/devr-tools/merger/internal/domain"

func criticalityRank(value domain.Criticality) int {
	switch value {
	case domain.CriticalityTier0:
		return 4
	case domain.CriticalityHigh:
		return 3
	case domain.CriticalityNormal:
		return 2
	default:
		return 1
	}
}
