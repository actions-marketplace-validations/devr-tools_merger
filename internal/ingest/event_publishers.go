package ingest

import (
	"context"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/events"
)

func (p *Processor) publishPROpened(ctx context.Context, fullName string, prNumber int, action string) error {
	return p.bus.Publish(ctx, events.NewEnvelope(events.EventPROpened, "ingest", map[string]any{
		"repo":     fullName,
		"prNumber": prNumber,
		"action":   action,
	}))
}

func (p *Processor) publishMutations(ctx context.Context, mutations []domain.Mutation) error {
	return p.bus.Publish(ctx, events.NewEnvelope(events.EventMutationDetected, "mutations", mutations))
}

func (p *Processor) publishRisk(ctx context.Context, risks []domain.Risk) error {
	return p.bus.Publish(ctx, events.NewEnvelope(events.EventRiskAssigned, "risk", risks))
}

func (p *Processor) publishViolations(ctx context.Context, violations []domain.PolicyViolation) error {
	return p.bus.Publish(ctx, events.NewEnvelope(events.EventPolicyViolationDetected, "policy", violations))
}

func (p *Processor) publishLane(ctx context.Context, packetID string, lane domain.MergeLane) error {
	return p.bus.Publish(ctx, events.NewEnvelope(events.EventMergeLaneAssigned, "lanes", map[string]any{
		"changePacketId": packetID,
		"lane":           lane,
	}))
}
