package main

import (
	"log"
	"os"

	ingestapp "github.com/mergerhq/merger/internal/app/ingest"
	"github.com/mergerhq/merger/internal/config"
	"github.com/mergerhq/merger/internal/events"
	"github.com/mergerhq/merger/internal/github"
	"github.com/mergerhq/merger/internal/policy"
	"github.com/mergerhq/merger/internal/telemetry"
)

func main() {
	configPath := os.Getenv("MERGER_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/merger.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	policyConfig, err := policy.LoadConfig(cfg.Policy.Path)
	if err != nil {
		log.Fatal(err)
	}

	logger := telemetry.NewLogger(cfg.Logging.Level)
	bus := events.NewMemoryBus()
	app := ingestapp.New(
		cfg,
		logger,
		telemetry.NewTracer(),
		bus,
		github.NoopService{},
		policy.NewRuleEngine(policyConfig),
	)

	log.Fatal(app.Run())
}
