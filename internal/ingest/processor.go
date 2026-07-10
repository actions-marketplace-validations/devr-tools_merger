package ingest

import (
	"context"

	"github.com/devr-tools/merger/internal/checks"
	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/events"
	"github.com/devr-tools/merger/internal/github"
	"github.com/devr-tools/merger/internal/lanes"
	"github.com/devr-tools/merger/internal/mutations"
	"github.com/devr-tools/merger/internal/policy"
	"github.com/devr-tools/merger/internal/risk"
	"github.com/devr-tools/merger/internal/runtimegraph"
	"github.com/devr-tools/merger/internal/store"
	"github.com/devr-tools/merger/internal/telemetry"
)

type Processor struct {
	logger    *telemetry.Logger
	tracer    telemetry.Tracer
	bus       events.Bus
	github    github.Service
	mutations mutations.Engine
	risk      risk.Engine
	policy    policy.Engine
	assigner  lanes.Assigner
	checks    checks.Publisher
	runtime   runtimegraph.Resolver
	store     store.ChangePacketStore
}

func NewProcessor(
	logger *telemetry.Logger,
	tracer telemetry.Tracer,
	bus events.Bus,
	githubService github.Service,
	mutationEngine mutations.Engine,
	riskEngine risk.Engine,
	policyEngine policy.Engine,
	assigner lanes.Assigner,
	checkPublisher checks.Publisher,
	runtimeResolver runtimegraph.Resolver,
	packetStore store.ChangePacketStore,
) *Processor {
	return &Processor{
		logger:    logger,
		tracer:    tracer,
		bus:       bus,
		github:    githubService,
		mutations: mutationEngine,
		risk:      riskEngine,
		policy:    policyEngine,
		assigner:  assigner,
		checks:    checkPublisher,
		runtime:   runtimeResolver,
		store:     packetStore,
	}
}

func (p *Processor) ProcessPROpened(ctx context.Context, payload github.PullRequestWebhookPayload) (*domain.ChangePacket, error) {
	return p.processPROpened(ctx, payload)
}
