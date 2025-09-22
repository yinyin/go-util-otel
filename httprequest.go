package utilotel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type HTTPRequestSpanBuilder struct {
	Tracer     trace.Tracer
	Propagator propagation.TextMapPropagator
}

func (b *HTTPRequestSpanBuilder) Init(tracer trace.Tracer, propagator propagation.TextMapPropagator) {
	if tracer == nil {
		tracer = otel.Tracer("")
	}
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	b.Tracer = tracer
	b.Propagator = propagator
}

func (b *HTTPRequestSpanBuilder) Start(ctx context.Context, r *http.Request, spanName string) (context.Context, trace.Span) {
	ctx = b.Propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
	reqAttrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodOriginal(r.Method),
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		reqAttrs = append(reqAttrs, attribute.String("http.request.header.x_forwarded_for", v))
	}
	if v := r.Header.Get("Forwarded"); v != "" {
		reqAttrs = append(reqAttrs, attribute.String("http.request.header.forwarded", v))
	}
	if v := r.Header.Get("Origin"); v != "" {
		reqAttrs = append(reqAttrs, attribute.String("http.request.header.origin", v))
	}
	if vs := r.Header.Values("Access-Control-Request-Headers"); len(vs) > 0 {
		reqAttrs = append(reqAttrs, attribute.StringSlice("http.request.header.access_control_request_headers", vs))
	}
	if vs := r.Header.Values("Access-Control-Request-Method"); len(vs) > 0 {
		reqAttrs = append(reqAttrs, attribute.StringSlice("http.request.header.access_control_request_method", vs))
	}
	if v := r.Header.Get("Referer"); v != "" {
		reqAttrs = append(reqAttrs, attribute.String("http.request.header.referer", v))
	}
	if vs := r.Header.Values("Cookie"); len(vs) > 0 {
		reqAttrs = append(reqAttrs, attribute.StringSlice("http.request.header.cookie", vs))
	}
	if v := r.UserAgent(); v != "" {
		reqAttrs = append(reqAttrs, semconv.UserAgentOriginal(v))
	}
	reqAttrs = AppendNetPeerConnAttributes(reqAttrs, r.RemoteAddr)
	if r.URL != nil {
		reqAttrs = append(reqAttrs, semconv.URLPath(r.URL.Path))
		if r.URL.RawQuery != "" {
			reqAttrs = append(reqAttrs, semconv.URLQuery(r.URL.RawQuery))
		}
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(reqAttrs...),
	}
	if s := trace.SpanContextFromContext(ctx); s.IsValid() && s.IsRemote() {
		opts = append(opts, trace.WithNewRoot(), trace.WithLinks(trace.Link{SpanContext: s}))
	}
	return b.Tracer.Start(ctx, spanName, opts...)
}
