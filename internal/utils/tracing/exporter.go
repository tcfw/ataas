package tracing

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	prop "go.opentelemetry.io/contrib/propagators/jaeger"
	// prop "go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

//InitTracer creates a new trace provider instance and registers it as global trace provider.
func InitTracer(serviceName string) func() {
	// Create and install Jaeger export pipeline
	exporter, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(viper.GetString("tracing.endpoint"))),
	)
	if err != nil {
		logrus.New().Fatal(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(&prop.Jaeger{})

	return func() {
		tp.Shutdown(context.Background())
	}
}
