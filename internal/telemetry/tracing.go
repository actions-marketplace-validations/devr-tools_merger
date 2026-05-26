package telemetry

import "context"

// OTelReadyTracer keeps the call-sites instrumented while Phase 1 remains SDK-light.
// A future OpenTelemetry provider can replace this implementation without touching
// the services or engine packages.
type OTelReadyTracer struct{}

type Span interface {
	SetAttributes(map[string]any)
	RecordError(error)
	End()
}

type Tracer interface {
	Start(context.Context, string) (context.Context, Span)
}

type noopSpan struct{}

func NewTracer() Tracer {
	return OTelReadyTracer{}
}

func (OTelReadyTracer) Start(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, noopSpan{}
}

func (noopSpan) SetAttributes(map[string]any) {}
func (noopSpan) RecordError(error)            {}
func (noopSpan) End()                         {}
