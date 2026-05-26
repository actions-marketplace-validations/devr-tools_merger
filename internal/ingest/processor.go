package ingest

import (
	"context"

	"github.com/mergerhq/merger/internal/checks"
	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/events"
	"github.com/mergerhq/merger/internal/github"
	"github.com/mergerhq/merger/internal/lanes"
	"github.com/mergerhq/merger/internal/mutations"
	"github.com/mergerhq/merger/internal/policy"
	"github.com/mergerhq/merger/internal/risk"
	"github.com/mergerhq/merger/internal/runtimegraph"
	"github.com/mergerhq/merger/internal/store"
	"github.com/mergerhq/merger/internal/telemetry"
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
