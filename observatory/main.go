package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
)

// Prometheus metrics
var (
	tracesReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_observatory_traces_received_total",
		Help: "Total number of traces received",
	})
	
	metricsReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_observatory_metrics_received_total",
		Help: "Total number of metrics received",
	})
	
	logsReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_observatory_logs_received_total",
		Help: "Total number of logs received",
	})
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(tracesReceived)
	prometheus.MustRegister(metricsReceived)
	prometheus.MustRegister(logsReceived)
	
	// Load configuration
	viper.SetDefault("PORT", "8083")
	viper.SetDefault("PROMETHEUS_PORT", "9090")
	viper.SetDefault("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces")
	viper.SetDefault("OTLP_ENDPOINT", "localhost:4317")
	
	viper.AutomaticEnv()
}

func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(viper.GetString("OTLP_ENDPOINT")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}
	
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("chronos-observatory"),
		semconv.ServiceVersionKey.String("0.1.0"),
	)
	
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	
	otel.SetTracerProvider(provider)
	
	return provider, nil
}

func setupOTLPCollector() error {
	// In a real implementation, this would set up the OpenTelemetry Collector
	// with appropriate receivers, processors, and exporters
	
	// For this sample, we'll just simulate the setup
	log.Println("Setting up OpenTelemetry Collector")
	
	// This is a placeholder for the actual OpenTelemetry Collector setup
	// In a production environment, you would:
	// 1. Create component factories for receivers, processors, and exporters
	// 2. Load configuration from files or environment variables
	// 3. Build and start the collector pipeline
	
	return nil
}

func main() {
	log.Println("Starting Chronos Observatory service...")
	
	// Initialize OpenTelemetry
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	
	// Set up OpenTelemetry Collector
	if err := setupOTLPCollector(); err != nil {
		log.Fatalf("Failed to set up OpenTelemetry Collector: %v", err)
	}
	
	// Set up gRPC server
	port := viper.GetString("PORT")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	// In a real implementation, this would register the observatory service
	
	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	
	// Set up HTTP server for metrics
	http.Handle("/metrics", promhttp.Handler())
	
	// Add a simple status endpoint
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Observatory service is running"))
	})
	
	// Start HTTP server in a goroutine
	prometheusPort := viper.GetString("PROMETHEUS_PORT")
	httpServer := &http.Server{Addr: fmt.Sprintf(":%s", prometheusPort)}
	go func() {
		log.Printf("Starting metrics server on port %s", prometheusPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down servers...")
	
	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	// Stop gRPC server
	grpcServer.GracefulStop()
	
	log.Println("Servers exited properly")
}
