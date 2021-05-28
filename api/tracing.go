package api

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func registerTelemetryMiddleware(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "ataas.gateway/ServeHTTP")
}
