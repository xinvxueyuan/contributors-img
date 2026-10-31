// Package tracing carries request context across service boundaries.
//
// This deployment runs as a single container on Vercel: there are no GCP
// credentials and no Cloud Trace exporter, so spans come from OpenTelemetry's
// no-op provider and only W3C propagation is installed. The previous
// implementation built a Cloud Trace exporter at startup and called log.Fatal
// when it failed, which would have killed the process on every boot.
package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Init installs the propagators used to continue a trace across HTTP calls.
func Init() {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
}

func Tracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer("")
}

// TraceNameFromContext returns the Cloud Logging trace name. There is no Cloud
// Logging here, so it is always empty; the function stays because the logger
// reads it on every request.
func TraceNameFromContext(c context.Context) string {
	return ""
}
