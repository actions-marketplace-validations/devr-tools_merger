package store

import (
	"context"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/events"
)

type NoopRepository struct{}

func NewNoopRepository() NoopRepository {
	return NoopRepository{}
}

func (NoopRepository) SaveChangePacket(context.Context, domain.ChangePacket) error { return nil }
func (NoopRepository) SaveEvent(context.Context, events.Envelope) error            { return nil }
func (NoopRepository) Migrate(context.Context) error                               { return nil }
func (NoopRepository) Close() error                                                { return nil }
