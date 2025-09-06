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
	"github.com/robfig/cron/v3"
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
	scheduledWorkflows = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_scheduler_workflows_scheduled_total",
		Help: "Total number of workflows scheduled",
	})
	
	schedulingLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "chronos_scheduler_scheduling_latency_seconds",
		Help:    "Latency of scheduling operations in seconds",
		Buckets: prometheus.DefBuckets,
	})
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(scheduledWorkflows)
	prometheus.MustRegister(schedulingLatency)
	
	// Load configuration
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_TOPIC", "chronos-workflows")
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
		semconv.ServiceNameKey.String("chronos-scheduler"),
		semconv.ServiceVersionKey.String("0.1.0"),
	)
	
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	
	otel.SetTracerProvider(provider)
	
	return provider, nil
}

func main() {
	log.Println("Starting Chronos Scheduler service...")
	
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
	
	// Create a new cron scheduler
	c := cron.New(cron.WithSeconds())
	
	// Start the cron scheduler
	c.Start()
	defer c.Stop()
	
	// Set up gRPC server
	port := viper.GetString("PORT")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	// Register the scheduler service (implementation would be in a separate file)
	// scheduler.RegisterSchedulerServiceServer(grpcServer, &schedulerServer{})
	
	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	
	// Set up HTTP server for metrics
	http.Handle("/metrics", promhttp.Handler())
	
	// Start HTTP server in a goroutine
	httpServer := &http.Server{Addr: ":8090"}
	go func() {
		log.Println("Starting metrics server on :8090")
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	// Stop gRPC server
	grpcServer.GracefulStop()
	
	log.Println("Servers exited properly")
}
