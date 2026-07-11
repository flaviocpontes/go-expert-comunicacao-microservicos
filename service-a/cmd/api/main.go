package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-a/internal/config"
	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-a/pkg/clients/serviceb"
	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-a/pkg/servicea"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initTracer(otelCollectorAddr string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("service-a"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, otelCollectorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}

func main() {
	cfg := config.LoadConfig()

	tp, err := initTracer(cfg.OtelExporterEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	serviceBClient := serviceb.NewClient(cfg.ServiceBURL)

	http.HandleFunc("/", servicea.NewHandler(serviceBClient))

	fmt.Printf("Service A starting on port %s...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
