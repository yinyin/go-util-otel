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

func splitHostPort(hostPort string) (host string, port int) {
	portMultiplier := 1
	for idx := len(hostPort) - 1; idx >= 0; idx-- {
		ch := hostPort[idx]
		if ch == ':' {
			host = hostPort[:idx]
			if (len(host) > 2) && (host[0] == '[') && (host[len(host)-1] == ']') {
				host = host[1 : len(host)-1]
			}
			return
		}
		if ch >= '0' && ch <= '9' {
			port += int(ch-'0') * portMultiplier
			portMultiplier *= 10
		} else {
			port = 0
			host = hostPort
			return
		}
	}
	return
}

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
	if v := r.UserAgent(); v != "" {
		reqAttrs = append(reqAttrs, semconv.UserAgentOriginal(v))
	}
	remoteHost, remotePort := splitHostPort(r.RemoteAddr)
	if remoteHost != "" {
		reqAttrs = append(reqAttrs, semconv.NetworkPeerAddress(remoteHost))
	}
	if remotePort != 0 {
		reqAttrs = append(reqAttrs, semconv.NetworkPeerPort(remotePort))
	}
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
