package store

import (
	"context"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/events"
)

type ChangePacketStore interface {
	SaveChangePacket(context.Context, domain.ChangePacket) error
}

type EventStore interface {
	SaveEvent(context.Context, events.Envelope) error
}

type Repository interface {
	ChangePacketStore
	EventStore
	Migrate(context.Context) error
	Close() error
}
