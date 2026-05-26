package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

type NATSConfig struct {
	URL           string
	StreamName    string
	SubjectPrefix string
	DurablePrefix string
}

type NATSBus struct {
	conn *nats.Conn
	js   nats.JetStreamContext
	cfg  NATSConfig
	subs []*nats.Subscription
}

func NewNATSBus(cfg NATSConfig) (*NATSBus, error) {
	conn, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, err
	}

	bus := &NATSBus{
		conn: conn,
		js:   js,
		cfg:  cfg,
	}

	if err := bus.ensureStream(); err != nil {
		bus.conn.Close()
		return nil, err
	}

	return bus, nil
}

func (b *NATSBus) Publish(_ context.Context, event Envelope) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &nats.Msg{
		Subject: b.subjectFor(event.Type),
		Data:    payload,
		Header:  nats.Header{},
	}
	msg.Header.Set("X-MergeR-Event-ID", event.ID)
	msg.Header.Set("X-MergeR-Event-Type", string(event.Type))
	if event.CorrelationID != "" {
		msg.Header.Set("X-MergeR-Correlation-ID", event.CorrelationID)
	}

	_, err = b.js.PublishMsg(msg)
	return err
}

func (b *NATSBus) Subscribe(kind EventType, handler Handler) error {
	subject := b.subjectFor(kind)
	durable := fmt.Sprintf("%s-%s", b.cfg.DurablePrefix, sanitizeToken(string(kind)))

	sub, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		var envelope Envelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			_ = msg.Nak()
			return
		}

		if err := handler(context.Background(), envelope); err != nil {
			_ = msg.Nak()
			return
		}

		_ = msg.Ack()
	}, nats.ManualAck(), nats.Durable(durable), nats.DeliverNew(), nats.AckExplicit())
	if err != nil {
		return err
	}

	b.subs = append(b.subs, sub)
	return nil
}

func (b *NATSBus) Close() error {
	for _, sub := range b.subs {
		_ = sub.Drain()
	}
	if b.conn != nil {
		b.conn.Drain()
		b.conn.Close()
	}
	return nil
}

func (b *NATSBus) ensureStream() error {
	subjects := []string{fmt.Sprintf("%s.events.*", b.cfg.SubjectPrefix)}

	if _, err := b.js.StreamInfo(b.cfg.StreamName); err == nil {
		return nil
	}

	_, err := b.js.AddStream(&nats.StreamConfig{
		Name:     b.cfg.StreamName,
		Subjects: subjects,
	})

	return err
}

func (b *NATSBus) subjectFor(kind EventType) string {
	return fmt.Sprintf("%s.events.%s", b.cfg.SubjectPrefix, sanitizeToken(string(kind)))
}

func sanitizeToken(value string) string {
	sanitized := strings.ToLower(value)
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return sanitized
}
