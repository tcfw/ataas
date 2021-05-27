package api

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func registerTelemetryMiddleware(ctx context.Context, handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "ataas.gateway/ServeHTTP")
}
