package main

import (
	"log"
	"os"

	controlplaneapp "github.com/mergerhq/merger/internal/app/controlplane"
	"github.com/mergerhq/merger/internal/config"
	"github.com/mergerhq/merger/internal/events"
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
	bus := events.NewMemoryBus()
	app := controlplaneapp.New(cfg, logger, bus)

	log.Fatal(app.Run())
}
