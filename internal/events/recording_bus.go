package events

import "context"

type EventRecorder interface {
	SaveEvent(context.Context, Envelope) error
}

type RecordingBus struct {
	base  Bus
	store EventRecorder
}

func NewRecordingBus(base Bus, eventStore EventRecorder) *RecordingBus {
	return &RecordingBus{
		base:  base,
		store: eventStore,
	}
}

func (b *RecordingBus) Publish(ctx context.Context, event Envelope) error {
	if err := b.base.Publish(ctx, event); err != nil {
		return err
	}

	if b.store == nil {
		return nil
	}

	return b.store.SaveEvent(ctx, event)
}

func (b *RecordingBus) Subscribe(kind EventType, handler Handler) error {
	return b.base.Subscribe(kind, handler)
}

func (b *RecordingBus) Close() error {
	return b.base.Close()
}
