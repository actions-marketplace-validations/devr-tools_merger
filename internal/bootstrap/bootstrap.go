package bootstrap

import (
	"context"
	"fmt"

	"github.com/mergerhq/merger/internal/config"
	"github.com/mergerhq/merger/internal/events"
	"github.com/mergerhq/merger/internal/github"
	"github.com/mergerhq/merger/internal/store"
)

func BuildRepository(ctx context.Context, cfg config.Config) (store.Repository, error) {
	switch cfg.Persistence.Backend {
	case "", "memory":
		return store.NewNoopRepository(), nil
	case "postgres":
		repo, err := store.NewPostgresRepository(cfg.Persistence.DatabaseURL)
		if err != nil {
			return nil, err
		}

		if cfg.Persistence.AutoMigrate {
			if err := repo.Migrate(ctx); err != nil {
				_ = repo.Close()
				return nil, err
			}
		}

		return repo, nil
	default:
		return nil, fmt.Errorf("unsupported persistence backend: %s", cfg.Persistence.Backend)
	}
}

func BuildEventBus(cfg config.Config, repo store.Repository) (events.Bus, error) {
	baseBus, err := events.NewBusFromConfig(cfg.Events)
	if err != nil {
		return nil, err
	}

	return events.NewRecordingBus(baseBus, repo), nil
}

func BuildGitHubService(cfg config.Config) (github.Service, error) {
	if !cfg.GitHub.Enabled {
		return github.NoopService{}, nil
	}

	return github.NewClient(github.ClientConfig{
		AppID:          cfg.GitHub.AppID,
		InstallationID: cfg.GitHub.InstallationID,
		PrivateKeyPath: cfg.GitHub.PrivateKeyPath,
		WebhookSecret:  cfg.GitHub.WebhookSecret,
		APIBaseURL:     cfg.GitHub.APIBaseURL,
		Timeout:        cfg.GitHub.Timeout,
	})
}
