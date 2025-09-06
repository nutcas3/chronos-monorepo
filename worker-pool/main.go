package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	tasksExecuted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_worker_tasks_executed_total",
		Help: "Total number of tasks executed",
	})
	
	taskSuccesses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_worker_tasks_succeeded_total",
		Help: "Total number of tasks executed successfully",
	})
	
	taskFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chronos_worker_tasks_failed_total",
		Help: "Total number of tasks that failed execution",
	})
	
	executionLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "chronos_worker_execution_latency_seconds",
		Help:    "Latency of task execution operations in seconds",
		Buckets: prometheus.DefBuckets,
	})
)

// Worker represents a single worker in the pool
type Worker struct {
	ID          string
	TaskTypes   []string
	Capacity    int
	CurrentLoad int
	ActiveTasks map[string]struct{}
	mu          sync.Mutex
}

// WorkerPool manages a collection of workers
type WorkerPool struct {
	Workers map[string]*Worker
	mu      sync.RWMutex
}

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(tasksExecuted)
	prometheus.MustRegister(taskSuccesses)
	prometheus.MustRegister(taskFailures)
	prometheus.MustRegister(executionLatency)
	
	// Load configuration
	viper.SetDefault("PORT", "8082")
	viper.SetDefault("DURABLE_ENGINE_URL", "localhost:50051")
	viper.SetDefault("WORKER_COUNT", 5)
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
		semconv.ServiceNameKey.String("chronos-worker-pool"),
		semconv.ServiceVersionKey.String("0.1.0"),
	)
	
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	
	otel.SetTracerProvider(provider)
	
	return provider, nil
}

func createWorkerPool() *WorkerPool {
	workerCount := viper.GetInt("WORKER_COUNT")
	pool := &WorkerPool{
		Workers: make(map[string]*Worker),
	}
	
	for i := 0; i < workerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		worker := &Worker{
			ID:          workerID,
			TaskTypes:   []string{"http", "process", "database", "file"},
			Capacity:    10,
			CurrentLoad: 0,
			ActiveTasks: make(map[string]struct{}),
		}
		
		pool.Workers[workerID] = worker
	}
	
	return pool
}

// WorkerServer implements the gRPC worker service
type WorkerServer struct {
	Pool *WorkerPool
	// In a real implementation, this would include the generated gRPC server interface
}

func main() {
	log.Println("Starting Chronos Worker Pool service...")
	
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
	
	// Create worker pool
	pool := createWorkerPool()
	
	// Set up gRPC server
	port := viper.GetString("PORT")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	// Register the worker service
	// worker.RegisterWorkerServiceServer(grpcServer, &WorkerServer{Pool: pool})
	
	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	
	// Start task polling for each worker
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	
	for _, worker := range pool.Workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			pollForTasks(ctx, w)
		}(worker)
	}
	
	// Set up HTTP server for metrics
	http.Handle("/metrics", promhttp.Handler())
	
	// Start HTTP server in a goroutine
	httpServer := &http.Server{Addr: ":8092"}
	go func() {
		log.Println("Starting metrics server on :8092")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down servers...")
	
	// Cancel context to stop task polling
	cancel()
	
	// Wait for all workers to finish
	wg.Wait()
	
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

func pollForTasks(ctx context.Context, worker *Worker) {
	log.Printf("Worker %s started polling for tasks", worker.ID)
	
	// In a real implementation, this would:
	// 1. Connect to the Durable Engine via gRPC
	// 2. Poll for available tasks
	// 3. Execute tasks and report results
	// 4. Update metrics
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %s stopping", worker.ID)
			return
		case <-ticker.C:
			// Simulate task polling and execution
			log.Printf("Worker %s polling for tasks", worker.ID)
		}
	}
}
