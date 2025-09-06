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

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
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
	workflowsStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_executor_workflows_started_total",
		Help: "Total number of workflows started",
	})
	
	tasksDispatched = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_executor_tasks_dispatched_total",
		Help: "Total number of tasks dispatched",
	})
	
	dispatchLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "chronos_executor_dispatch_latency_seconds",
		Help:    "Latency of task dispatch operations in seconds",
		Buckets: prometheus.DefBuckets,
	})
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(workflowsStarted)
	prometheus.MustRegister(tasksDispatched)
	prometheus.MustRegister(dispatchLatency)
	
	// Load configuration
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_TOPIC_IN", "chronos-workflows")
	viper.SetDefault("KAFKA_TOPIC_OUT", "chronos-tasks")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379/0")
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
		semconv.ServiceNameKey.String("chronos-executor"),
		semconv.ServiceVersionKey.String("0.1.0"),
	)
	
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	
	otel.SetTracerProvider(provider)
	
	return provider, nil
}

func initRedis() (*redis.Client, error) {
	opts, err := redis.ParseURL(viper.GetString("REDIS_URL"))
	if err != nil {
		return nil, fmt.Errorf("parsing Redis URL: %w", err)
	}
	
	client := redis.NewClient(opts)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to Redis: %w", err)
	}
	
	return client, nil
}

func initKafkaReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{viper.GetString("KAFKA_BROKERS")},
		Topic:       viper.GetString("KAFKA_TOPIC_IN"),
		GroupID:     "chronos-executor",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})
}

func initKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(viper.GetString("KAFKA_BROKERS")),
		Topic:    viper.GetString("KAFKA_TOPIC_OUT"),
		Balancer: &kafka.LeastBytes{},
	}
}

func main() {
	log.Println("Starting Chronos Executor service...")
	
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
	
	// Initialize Redis
	redisClient, err := initRedis()
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisClient.Close()
	
	// Initialize Kafka reader and writer
	kafkaReader := initKafkaReader()
	defer kafkaReader.Close()
	
	kafkaWriter := initKafkaWriter()
	defer kafkaWriter.Close()
	
	// Start Kafka consumer in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	go consumeWorkflows(ctx, kafkaReader, kafkaWriter, redisClient)
	
	// Set up gRPC server
	port := viper.GetString("PORT")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	// Register the executor service (implementation would be in a separate file)
	// executor.RegisterExecutorServiceServer(grpcServer, &executorServer{})
	
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
	httpServer := &http.Server{Addr: ":8091"}
	go func() {
		log.Println("Starting metrics server on :8091")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down servers...")
	
	// Cancel context to stop Kafka consumer
	cancel()
	
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

func consumeWorkflows(ctx context.Context, reader *kafka.Reader, writer *kafka.Writer, redisClient *redis.Client) {
	log.Println("Starting Kafka consumer for workflows")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer")
			return
		default:
			message, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}
			
			log.Printf("Received message: %s", string(message.Value))
			
			// Process the workflow message
			// In a real implementation, this would:
			// 1. Parse the workflow definition
			// 2. Check for duplicates using Redis
			// 3. Fan out tasks to the task queue
			// 4. Update metrics
			
			workflowsStarted.Inc()
		}
	}
}
