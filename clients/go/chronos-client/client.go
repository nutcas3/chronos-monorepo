package chronosclient

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ChronosClient is the main client for interacting with the Chronos platform
type ChronosClient struct {
	schedulerConn   *grpc.ClientConn
	executorConn    *grpc.ClientConn
	durableEngConn  *grpc.ClientConn
	workerPoolConn  *grpc.ClientConn
	observatoryConn *grpc.ClientConn
	tracer          trace.Tracer
}

// ClientOptions contains options for creating a new ChronosClient
type ClientOptions struct {
	SchedulerURL   string
	ExecutorURL    string
	DurableEngURL  string
	WorkerPoolURL  string
	ObservatoryURL string
	TracerName     string
}

// DefaultClientOptions returns the default options for creating a new ChronosClient
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		SchedulerURL:   "localhost:8080",
		ExecutorURL:    "localhost:8081",
		DurableEngURL:  "localhost:50051",
		WorkerPoolURL:  "localhost:8082",
		ObservatoryURL: "localhost:8083",
		TracerName:     "chronos-client",
	}
}

// NewClient creates a new ChronosClient with the given options
func NewClient(opts *ClientOptions) (*ChronosClient, error) {
	if opts == nil {
		opts = DefaultClientOptions()
	}

	// Initialize tracer
	tracer := otel.Tracer(opts.TracerName)

	// Connect to scheduler service
	schedulerConn, err := grpc.NewClient(opts.SchedulerURL, 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to scheduler: %w", err)
	}

	// Connect to executor service
	executorConn, err := grpc.NewClient(opts.ExecutorURL, 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		schedulerConn.Close()
		return nil, fmt.Errorf("failed to connect to executor: %w", err)
	}

	// Connect to durable engine service
	durableEngConn, err := grpc.NewClient(opts.DurableEngURL, 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		schedulerConn.Close()
		executorConn.Close()
		return nil, fmt.Errorf("failed to connect to durable engine: %w", err)
	}

	// Connect to worker pool service
	workerPoolConn, err := grpc.NewClient(opts.WorkerPoolURL, 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		schedulerConn.Close()
		executorConn.Close()
		durableEngConn.Close()
		return nil, fmt.Errorf("failed to connect to worker pool: %w", err)
	}

	// Connect to observatory service
	observatoryConn, err := grpc.NewClient(opts.ObservatoryURL, 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		schedulerConn.Close()
		executorConn.Close()
		durableEngConn.Close()
		workerPoolConn.Close()
		return nil, fmt.Errorf("failed to connect to observatory: %w", err)
	}

	return &ChronosClient{
		schedulerConn:   schedulerConn,
		executorConn:    executorConn,
		durableEngConn:  durableEngConn,
		workerPoolConn:  workerPoolConn,
		observatoryConn: observatoryConn,
		tracer:          tracer,
	}, nil
}

// Close closes all connections
func (c *ChronosClient) Close() error {
	var errs []error

	if err := c.schedulerConn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close scheduler connection: %w", err))
	}

	if err := c.executorConn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close executor connection: %w", err))
	}

	if err := c.durableEngConn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close durable engine connection: %w", err))
	}

	if err := c.workerPoolConn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close worker pool connection: %w", err))
	}

	if err := c.observatoryConn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close observatory connection: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// Workflow represents a workflow in the Chronos system
type Workflow struct {
	ID          string
	Name        string
	Description string
	Tasks       []*Task
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Task represents a task in the Chronos system
type Task struct {
	ID          string
	WorkflowID  string
	Name        string
	Type        string
	Status      string
	Payload     []byte
	Result      []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// CreateWorkflow creates a new workflow
func (c *ChronosClient) CreateWorkflow(ctx context.Context, name, description string) (*Workflow, error) {
	ctx, span := c.tracer.Start(ctx, "ChronosClient.CreateWorkflow",
		trace.WithAttributes(
			attribute.String("workflow.name", name),
			attribute.String("workflow.description", description),
		))
	defer span.End()

	// In a real implementation, this would call the appropriate gRPC method
	// For now, we'll just create a mock workflow
	id := uuid.New().String()
	now := time.Now()

	return &Workflow{
		ID:          id,
		Name:        name,
		Description: description,
		Tasks:       []*Task{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AddTask adds a task to a workflow
func (c *ChronosClient) AddTask(ctx context.Context, workflowID, name, taskType string, payload []byte) (*Task, error) {
	ctx, span := c.tracer.Start(ctx, "ChronosClient.AddTask",
		trace.WithAttributes(
			attribute.String("workflow.id", workflowID),
			attribute.String("task.name", name),
			attribute.String("task.type", taskType),
		))
	defer span.End()

	// In a real implementation, this would call the appropriate gRPC method
	// For now, we'll just create a mock task
	id := uuid.New().String()
	now := time.Now()

	return &Task{
		ID:         id,
		WorkflowID: workflowID,
		Name:       name,
		Type:       taskType,
		Status:     "pending",
		Payload:    payload,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// StartWorkflow starts a workflow
func (c *ChronosClient) StartWorkflow(ctx context.Context, workflowID string) error {
	ctx, span := c.tracer.Start(ctx, "ChronosClient.StartWorkflow",
		trace.WithAttributes(
			attribute.String("workflow.id", workflowID),
		))
	defer span.End()

	// In a real implementation, this would call the appropriate gRPC method
	return nil
}

// GetWorkflow gets a workflow by ID
func (c *ChronosClient) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
	ctx, span := c.tracer.Start(ctx, "ChronosClient.GetWorkflow",
		trace.WithAttributes(
			attribute.String("workflow.id", workflowID),
		))
	defer span.End()

	// In a real implementation, this would call the appropriate gRPC method
	// For now, we'll just return a mock workflow
	now := time.Now()

	return &Workflow{
		ID:          workflowID,
		Name:        "Mock Workflow",
		Description: "This is a mock workflow",
		Tasks:       []*Task{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetTask gets a task by ID
func (c *ChronosClient) GetTask(ctx context.Context, taskID string) (*Task, error) {
	ctx, span := c.tracer.Start(ctx, "ChronosClient.GetTask",
		trace.WithAttributes(
			attribute.String("task.id", taskID),
		))
	defer span.End()

	// In a real implementation, this would call the appropriate gRPC method
	// For now, we'll just return a mock task
	now := time.Now()

	return &Task{
		ID:         taskID,
		WorkflowID: "mock-workflow-id",
		Name:       "Mock Task",
		Type:       "http",
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
