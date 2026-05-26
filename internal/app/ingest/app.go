package ingestapp

import (
	"context"
	"net/http"

	"github.com/mergerhq/merger/internal/checks"
	"github.com/mergerhq/merger/internal/config"
	"github.com/mergerhq/merger/internal/events"
	"github.com/mergerhq/merger/internal/github"
	"github.com/mergerhq/merger/internal/ingest"
	"github.com/mergerhq/merger/internal/lanes"
	"github.com/mergerhq/merger/internal/mutations"
	"github.com/mergerhq/merger/internal/policy"
	"github.com/mergerhq/merger/internal/risk"
	"github.com/mergerhq/merger/internal/runtimegraph"
	"github.com/mergerhq/merger/internal/store"
	"github.com/mergerhq/merger/internal/telemetry"
)

type App struct {
	server *http.Server
}

func New(
	cfg config.Config,
	logger *telemetry.Logger,
	tracer telemetry.Tracer,
	bus events.Bus,
	githubService github.Service,
	policyEngine policy.Engine,
	repository store.Repository,
) *App {
	processor := ingest.NewProcessor(
		logger,
		tracer,
		bus,
		githubService,
		mutations.DefaultEngine(),
		risk.DefaultEngine(),
		policyEngine,
		lanes.NewAssigner(lanes.Config{
			GreenMax:  cfg.Lanes.GreenMax,
			YellowMax: cfg.Lanes.YellowMax,
			RedMax:    cfg.Lanes.RedMax,
		}),
		checks.NewGitHubCheckPublisher(githubService),
		runtimegraph.NewResolver(runtimegraph.Options{
			EnableCodeOwners: cfg.RuntimeGraph.EnableCodeOwners,
		}),
		repository,
	)

	mux := http.NewServeMux()
	mux.Handle("/webhooks/github", ingest.NewWebhookHandler(processor, github.NewWebhookDecoder(cfg.GitHub.WebhookSecret)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &App{
		server: &http.Server{
			Addr:    cfg.Service.IngestAddress,
			Handler: mux,
		},
	}
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
