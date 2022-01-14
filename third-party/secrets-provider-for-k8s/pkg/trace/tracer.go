package trace

import (
	"context"
	"fmt"

	traceotel "go.opentelemetry.io/otel/trace"
)

// Tracer is responsible for creating trace Spans.
type Tracer interface {
	// Start creates a span and a context.Context containing the newly-created
	// span.
	//
	// If the context.Context provided in `ctx` contains a Span then the
	// newly-created Span will be a child of that span, otherwise it will be a
	// root span.
	Start(ctx context.Context, spanName string) (context.Context, Span)
}

// otelTracer implements the Tracer interface based on an OpenTelemetry
// Tracer.
type otelTracer struct {
	tracerOtel traceotel.Tracer
}

func NewOtelTracer(tracerOtel traceotel.Tracer) otelTracer {
	return otelTracer{tracerOtel: tracerOtel}
}

func (t otelTracer) Start(ctx context.Context, spanName string) (context.Context, Span) {
	fmt.Printf("***TEMP*** Method Start() called in secrets provider trace package!\n")
	newCtx, spanOtel := t.tracerOtel.Start(ctx, spanName)
	spanCtx := traceotel.SpanContextFromContext(newCtx)
	fmt.Printf("***TEMP*** Span Context:\n")
	fmt.Printf("***TEMP***    SpanID: %x\n", spanCtx.SpanID())
	fmt.Printf("***TEMP***    TraceID: %x\n", spanCtx.TraceID())
	return newCtx, newOtelSpan(spanOtel)
}
