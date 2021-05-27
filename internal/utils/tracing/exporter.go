package tracing

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

//InitTracer creates a new trace provider instance and registers it as global trace provider.
func InitTracer(serviceName string) func() {
	// Create and install Jaeger export pipeline
	exporter, err := jaeger.NewRawExporter(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(viper.GetString("tracing-endpoint"))))
	if err != nil {
		logrus.New().Fatal(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp))
	otel.SetTracerProvider(tp)

	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	return func() {
		tp.Shutdown(context.Background())
	}
}
