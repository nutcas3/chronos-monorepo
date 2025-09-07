use anyhow::Result;
use sqlx::postgres::{PgPool, PgPoolOptions};
use std::env;
use tracing::info;

pub async fn init_db_pool() -> Result<PgPool> {
    let database_url = env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgres://postgres:postgres@localhost/chronos".to_string());
    
    info!("Connecting to database...");

    let pool = PgPoolOptions::new()
        .max_connections(20)
        .connect(&database_url)
        .await?;

    info!("Database connection established");

    // Run embedded migrations
    sqlx::migrate!("./migrations")
        .run(&pool)
        .await?;

    info!("Migrations applied successfully");

    Ok(pool)
}

// Example of type-safe queries using sqlx::query!() macro
// These queries are checked at compile time against your database schema

/// Get a task by ID with compile-time type checking
pub async fn get_task_by_id(pool: &PgPool, task_id: uuid::Uuid) -> Result<Option<Task>> {
    let row = sqlx::query!(
        "SELECT id, workflow_id, name, state, retry_count, max_retries, 
         created_at, updated_at, started_at, completed_at, timeout_seconds, 
         parameters, result, error 
         FROM tasks WHERE id = $1",
        task_id
    )
    .fetch_optional(pool)
    .await?;

    Ok(row.map(|r| Task {
        id: r.id,
        workflow_id: r.workflow_id,
        name: r.name,
        state: r.state,
        retry_count: r.retry_count,
        max_retries: r.max_retries,
        created_at: r.created_at,
        updated_at: r.updated_at,
        started_at: r.started_at,
        completed_at: r.completed_at,
        timeout_seconds: r.timeout_seconds,
        parameters: r.parameters,
        result: r.result,
        error: r.error,
    }))
}

/// Update task state with compile-time type checking
pub async fn update_task_state(
    pool: &PgPool, 
    task_id: uuid::Uuid, 
    new_state: &str
) -> Result<()> {
    sqlx::query!(
        "UPDATE tasks SET state = $1, updated_at = NOW() WHERE id = $2",
        new_state,
        task_id
    )
    .execute(pool)
    .await?;

    Ok(())
}

/// Get tasks by workflow ID with compile-time type checking
pub async fn get_tasks_by_workflow(
    pool: &PgPool, 
    workflow_id: uuid::Uuid
) -> Result<Vec<Task>> {
    let rows = sqlx::query!(
        "SELECT id, workflow_id, name, state, retry_count, max_retries, 
         created_at, updated_at, started_at, completed_at, timeout_seconds, 
         parameters, result, error 
         FROM tasks WHERE workflow_id = $1 ORDER BY created_at",
        workflow_id
    )
    .fetch_all(pool)
    .await?;

    Ok(rows.into_iter().map(|r| Task {
        id: r.id,
        workflow_id: r.workflow_id,
        name: r.name,
        state: r.state,
        retry_count: r.retry_count,
        max_retries: r.max_retries,
        created_at: r.created_at,
        updated_at: r.updated_at,
        started_at: r.started_at,
        completed_at: r.completed_at,
        timeout_seconds: r.timeout_seconds,
        parameters: r.parameters,
        result: r.result,
        error: r.error,
    }).collect())
}

// Task struct for the type-safe queries above
#[derive(Debug, Clone)]
pub struct Task {
    pub id: uuid::Uuid,
    pub workflow_id: uuid::Uuid,
    pub name: String,
    pub state: String,
    pub retry_count: i32,
    pub max_retries: i32,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub timeout_seconds: i32,
    pub parameters: serde_json::Value,
    pub result: Option<serde_json::Value>,
    pub error: Option<String>,
}

