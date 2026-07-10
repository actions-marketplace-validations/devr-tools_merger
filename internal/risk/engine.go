package risk

import (
	"context"

	"github.com/devr-tools/merger/internal/domain"
)

type Engine interface {
	Evaluate(context.Context, domain.ChangePacket) (domain.RiskSummary, []domain.Risk, error)
}

type WeightedEngine struct {
	weights map[domain.MutationKind]int
}

func DefaultEngine() *WeightedEngine {
	return &WeightedEngine{weights: defaultWeights()}
}
