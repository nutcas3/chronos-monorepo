use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TaskState {
    Queued,
    Running,
    Completed,
    Failed,
    Retrying,
    Cancelled,
    TimedOut,
}

impl std::fmt::Display for TaskState {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            TaskState::Queued => write!(f, "QUEUED"),
            TaskState::Running => write!(f, "RUNNING"),
            TaskState::Completed => write!(f, "COMPLETED"),
            TaskState::Failed => write!(f, "FAILED"),
            TaskState::Retrying => write!(f, "RETRYING"),
            TaskState::Cancelled => write!(f, "CANCELLED"),
            TaskState::TimedOut => write!(f, "TIMED_OUT"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Task {
    pub id: Uuid,
    pub workflow_id: Uuid,
    pub name: String,
    pub state: TaskState,
    pub retry_count: i32,
    pub max_retries: i32,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub timeout_seconds: i32,
    pub parameters: serde_json::Value,
    pub result: Option<serde_json::Value>,
    pub error: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Workflow {
    pub id: Uuid,
    pub name: String,
    pub state: TaskState,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub tasks: Vec<Task>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskEvent {
    pub id: Uuid,
    pub task_id: Uuid,
    pub workflow_id: Uuid,
    pub event_type: String,
    pub previous_state: Option<TaskState>,
    pub new_state: TaskState,
    pub timestamp: DateTime<Utc>,
    pub metadata: Option<serde_json::Value>,
}
