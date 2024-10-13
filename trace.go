package utilotel

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type SwitchableTraceExporter struct {
	currentModeIndex int32

	stdoutExporter *stdouttrace.Exporter
	otlpExporter   *otlptrace.Exporter

	fnExportSpans [switchModeCount]callableExportSpans
}

func NewSwitchableTraceExporter(
	ctx context.Context,
	stdoutOpts []stdouttrace.Option,
	otlpGRPCOpts []otlptracegrpc.Option) (traceExporter *SwitchableTraceExporter, err error) {
	x := &SwitchableTraceExporter{}
	if stdoutOpts != nil {
		if x.stdoutExporter, err = stdouttrace.New(stdoutOpts...); nil != err {
			err = fmt.Errorf("failed to create stdout trace exporter: %w", err)
			return
		}
		x.fnExportSpans[ExportModeSTDOUT] = x.stdoutExporter.ExportSpans
	}
	if otlpGRPCOpts != nil {
		if x.otlpExporter, err = otlptracegrpc.New(ctx, otlpGRPCOpts...); nil != err {
			err = fmt.Errorf("failed to create OTLP gRPC trace exporter: %w", err)
			x.Shutdown(ctx)
			return
		}
		x.fnExportSpans[ExportModeOTLPgRPC] = x.otlpExporter.ExportSpans
	}
	traceExporter = x
	return
}

func (x *SwitchableTraceExporter) SetMode(ctx context.Context, mode ExportMode) (err error) {
	if err = checkExportMode(mode); nil != err {
		return
	}
	atomic.StoreInt32(&x.currentModeIndex, int32(mode))
	return
}

func (x *SwitchableTraceExporter) modeIndex() int {
	return int(atomic.LoadInt32(&x.currentModeIndex))
}

func (x *SwitchableTraceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) (err error) {
	fn := x.fnExportSpans[x.modeIndex()]
	if nil == fn {
		return
	}
	return fn(ctx, spans)
}

func (x *SwitchableTraceExporter) Shutdown(ctx context.Context) (err error) {
	var allErrs []error
	if x.stdoutExporter != nil {
		if err = x.stdoutExporter.Shutdown(ctx); nil != err {
			allErrs = append(allErrs, err)
		}
	}
	if x.otlpExporter != nil {
		if err = x.otlpExporter.Shutdown(ctx); nil != err {
			allErrs = append(allErrs, err)
		}
	}
	if errCount := len(allErrs); errCount == 1 {
		err = allErrs[0]
	} else if errCount > 0 {
		err = errors.Join(allErrs...)
	}
	x.fnExportSpans = [switchModeCount]callableExportSpans{}
	return
}
