use anyhow::Result;
use async_trait::async_trait;
use chrono::{DateTime, Utc};
use opentelemetry::trace::{Span, Tracer};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use thiserror::Error;
use tokio::sync::Mutex;
use tonic::transport::{Channel, Endpoint};
use uuid::Uuid;

pub mod proto;

#[derive(Debug, Error)]
pub enum ChronosError {
    #[error("Connection error: {0}")]
    ConnectionError(String),
    
    #[error("Workflow error: {0}")]
    WorkflowError(String),
    
    #[error("Task error: {0}")]
    TaskError(String),
    
    #[error("Internal error: {0}")]
    InternalError(String),
}

#[derive(Debug, Clone)]
pub struct ClientOptions {
    pub scheduler_url: String,
    pub executor_url: String,
    pub durable_engine_url: String,
    pub worker_pool_url: String,
    pub observatory_url: String,
}

impl Default for ClientOptions {
    fn default() -> Self {
        Self {
            scheduler_url: "http://localhost:8080".to_string(),
            executor_url: "http://localhost:8081".to_string(),
            durable_engine_url: "http://localhost:50051".to_string(),
            worker_pool_url: "http://localhost:8082".to_string(),
            observatory_url: "http://localhost:8083".to_string(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Workflow {
    pub id: String,
    pub name: String,
    pub description: String,
    pub tasks: Vec<Task>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Task {
    pub id: String,
    pub workflow_id: String,
    pub name: String,
    pub task_type: String,
    pub status: TaskStatus,
    pub payload: Vec<u8>,
    pub result: Option<Vec<u8>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TaskStatus {
    Pending,
    Running,
    Completed,
    Failed,
    Cancelled,
}

impl std::fmt::Display for TaskStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            TaskStatus::Pending => write!(f, "pending"),
            TaskStatus::Running => write!(f, "running"),
            TaskStatus::Completed => write!(f, "completed"),
            TaskStatus::Failed => write!(f, "failed"),
            TaskStatus::Cancelled => write!(f, "cancelled"),
        }
    }
}

#[derive(Clone)]
pub struct ChronosClient {
    scheduler_channel: Channel,
    executor_channel: Channel,
    durable_engine_channel: Channel,
    worker_pool_channel: Channel,
    observatory_channel: Channel,
    tracer: Arc<opentelemetry::trace::Tracer>,
}

impl ChronosClient {
    pub async fn new(options: ClientOptions) -> Result<Self> {
        let scheduler_channel = Endpoint::from_shared(options.scheduler_url)?
            .connect()
            .await
            .map_err(|e| ChronosError::ConnectionError(format!("Failed to connect to scheduler: {}", e)))?;

        let executor_channel = Endpoint::from_shared(options.executor_url)?
            .connect()
            .await
            .map_err(|e| ChronosError::ConnectionError(format!("Failed to connect to executor: {}", e)))?;

        let durable_engine_channel = Endpoint::from_shared(options.durable_engine_url)?
            .connect()
            .await
            .map_err(|e| ChronosError::ConnectionError(format!("Failed to connect to durable engine: {}", e)))?;

        let worker_pool_channel = Endpoint::from_shared(options.worker_pool_url)?
            .connect()
            .await
            .map_err(|e| ChronosError::ConnectionError(format!("Failed to connect to worker pool: {}", e)))?;

        let observatory_channel = Endpoint::from_shared(options.observatory_url)?
            .connect()
            .await
            .map_err(|e| ChronosError::ConnectionError(format!("Failed to connect to observatory: {}", e)))?;

        let tracer = opentelemetry::global::tracer("chronos-client");

        Ok(Self {
            scheduler_channel,
            executor_channel,
            durable_engine_channel,
            worker_pool_channel,
            observatory_channel,
            tracer: Arc::new(tracer),
        })
    }

    pub async fn create_workflow(&self, name: &str, description: &str) -> Result<Workflow> {
        let mut span = self.tracer.start("ChronosClient.create_workflow");
        span.set_attribute(opentelemetry::KeyValue::new("workflow.name", name.to_string()));
        span.set_attribute(opentelemetry::KeyValue::new("workflow.description", description.to_string()));

        // In a real implementation, this would call the appropriate gRPC method
        // For now, we'll just create a mock workflow
        let id = Uuid::new_v4().to_string();
        let now = Utc::now();

        let workflow = Workflow {
            id,
            name: name.to_string(),
            description: description.to_string(),
            tasks: Vec::new(),
            created_at: now,
            updated_at: now,
        };

        Ok(workflow)
    }

    pub async fn add_task(&self, workflow_id: &str, name: &str, task_type: &str, payload: Vec<u8>) -> Result<Task> {
        let mut span = self.tracer.start("ChronosClient.add_task");
        span.set_attribute(opentelemetry::KeyValue::new("workflow.id", workflow_id.to_string()));
        span.set_attribute(opentelemetry::KeyValue::new("task.name", name.to_string()));
        span.set_attribute(opentelemetry::KeyValue::new("task.type", task_type.to_string()));

        // In a real implementation, this would call the appropriate gRPC method
        // For now, we'll just create a mock task
        let id = Uuid::new_v4().to_string();
        let now = Utc::now();

        let task = Task {
            id,
            workflow_id: workflow_id.to_string(),
            name: name.to_string(),
            task_type: task_type.to_string(),
            status: TaskStatus::Pending,
            payload,
            result: None,
            created_at: now,
            updated_at: now,
            started_at: None,
            completed_at: None,
        };

        Ok(task)
    }

    /// Start a workflow
    pub async fn start_workflow(&self, workflow_id: &str) -> Result<()> {
        let mut span = self.tracer.start("ChronosClient.start_workflow");
        span.set_attribute(opentelemetry::KeyValue::new("workflow.id", workflow_id.to_string()));

        // In a real implementation, this would call the appropriate gRPC method
        Ok(())
    }

    /// Get a workflow by ID
    pub async fn get_workflow(&self, workflow_id: &str) -> Result<Workflow> {
        let mut span = self.tracer.start("ChronosClient.get_workflow");
        span.set_attribute(opentelemetry::KeyValue::new("workflow.id", workflow_id.to_string()));

        // In a real implementation, this would call the appropriate gRPC method
        // For now, we'll just return a mock workflow
        let now = Utc::now();

        let workflow = Workflow {
            id: workflow_id.to_string(),
            name: "Mock Workflow".to_string(),
            description: "This is a mock workflow".to_string(),
            tasks: Vec::new(),
            created_at: now,
            updated_at: now,
        };

        Ok(workflow)
    }

    /// Get a task by ID
    pub async fn get_task(&self, task_id: &str) -> Result<Task> {
        let mut span = self.tracer.start("ChronosClient.get_task");
        span.set_attribute(opentelemetry::KeyValue::new("task.id", task_id.to_string()));

        // In a real implementation, this would call the appropriate gRPC method
        // For now, we'll just return a mock task
        let now = Utc::now();

        let task = Task {
            id: task_id.to_string(),
            workflow_id: "mock-workflow-id".to_string(),
            name: "Mock Task".to_string(),
            task_type: "http".to_string(),
            status: TaskStatus::Pending,
            payload: Vec::new(),
            result: None,
            created_at: now,
            updated_at: now,
            started_at: None,
            completed_at: None,
        };

        Ok(task)
    }
}

#[async_trait]
pub trait WorkflowExecutor {
    async fn execute(&self, workflow: &Workflow) -> Result<()>;
}

#[async_trait]
pub trait TaskExecutor {
    async fn execute(&self, task: &Task) -> Result<Vec<u8>>;
}
