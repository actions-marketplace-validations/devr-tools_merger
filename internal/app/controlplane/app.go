package controlplaneapp

import (
	"context"
	"net/http"

	"github.com/mergerhq/merger/internal/config"
	"github.com/mergerhq/merger/internal/events"
	"github.com/mergerhq/merger/internal/telemetry"
)

type App struct {
	logger *telemetry.Logger
	server *http.Server
	bus    events.Bus
}

func New(cfg config.Config, logger *telemetry.Logger, bus events.Bus) *App {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &App{
		logger: logger,
		bus:    bus,
		server: &http.Server{
			Addr:    cfg.Service.ControlPlaneAddress,
			Handler: mux,
		},
	}
}

func (a *App) Run() error {
	if err := a.bus.Subscribe(events.EventMergeLaneAssigned, func(_ context.Context, event events.Envelope) error {
		a.logger.Info("controlplane observed lane decision", "event", event)
		return nil
	}); err != nil {
		return err
	}

	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
