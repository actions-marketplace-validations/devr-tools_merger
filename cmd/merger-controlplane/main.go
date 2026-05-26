package main

import (
	"context"
	"log"
	"os"

	controlplaneapp "github.com/mergerhq/merger/internal/app/controlplane"
	"github.com/mergerhq/merger/internal/bootstrap"
	"github.com/mergerhq/merger/internal/config"
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

	logger := telemetry.NewLogger(cfg.Logging.Level)
	repo, err := bootstrap.BuildRepository(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	bus, err := bootstrap.BuildEventBus(cfg, repo)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	app := controlplaneapp.New(cfg, logger, bus, repo)

	log.Fatal(app.Run())
}
