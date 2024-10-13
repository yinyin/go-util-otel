package utilotel

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// define callable interface for methods of exporters.
//
// Ref:
// https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#SpanExporter

// callableExportSpans define callable for ExportSpans() of SpanExporter.
type callableExportSpans func(ctx context.Context, spans []sdktrace.ReadOnlySpan) error

// callableShutdown define callable for Shutdown() of SpanExporter.
// type callableShutdown func(ctx context.Context) error
