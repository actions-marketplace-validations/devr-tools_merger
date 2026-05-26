package events

import (
	"fmt"

	"github.com/mergerhq/merger/internal/config"
)

func NewBusFromConfig(cfg config.EventsConfig) (Bus, error) {
	switch cfg.Backend {
	case "", "memory":
		return NewMemoryBus(), nil
	case "nats":
		return NewNATSBus(NATSConfig{
			URL:           cfg.NATSURL,
			StreamName:    cfg.StreamName,
			SubjectPrefix: cfg.SubjectPrefix,
			DurablePrefix: cfg.DurablePrefix,
		})
	default:
		return nil, fmt.Errorf("unsupported events backend: %s", cfg.Backend)
	}
}
