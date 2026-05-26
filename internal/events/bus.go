package events

import (
	"context"
	"errors"
	"sync"

	"github.com/mergerhq/merger/pkg/identity"
)

type EventType string

const (
	EventPROpened                EventType = "PROpened"
	EventChangePacketCreated     EventType = "ChangePacketCreated"
	EventMutationDetected        EventType = "MutationDetected"
	EventRiskAssigned            EventType = "RiskAssigned"
	EventMergeLaneAssigned       EventType = "MergeLaneAssigned"
	EventPolicyViolationDetected EventType = "PolicyViolationDetected"
)

type Envelope struct {
	ID             string            `json:"id"`
	Type           EventType         `json:"type"`
	Source         string            `json:"source"`
	CorrelationID  string            `json:"correlationId,omitempty"`
	CausationID    string            `json:"causationId,omitempty"`
	ChangePacketID string            `json:"changePacketId,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Payload        any               `json:"payload,omitempty"`
}

type Handler func(context.Context, Envelope) error

type Bus interface {
	Publish(context.Context, Envelope) error
	Subscribe(EventType, Handler) error
	Close() error
}

type MemoryBus struct {
	mu       sync.RWMutex
	closed   bool
	handlers map[EventType][]Handler
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{handlers: make(map[EventType][]Handler)}
}

func NewEnvelope(kind EventType, source string, payload any) Envelope {
	return Envelope{
		ID:      identity.New("evt"),
		Type:    kind,
		Source:  source,
		Payload: payload,
	}
}

func (b *MemoryBus) Publish(ctx context.Context, event Envelope) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return errors.New("event bus closed")
	}

	for _, handler := range b.handlers[event.Type] {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (b *MemoryBus) Subscribe(kind EventType, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return errors.New("event bus closed")
	}

	b.handlers[kind] = append(b.handlers[kind], handler)
	return nil
}

func (b *MemoryBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.handlers = map[EventType][]Handler{}
	return nil
}
