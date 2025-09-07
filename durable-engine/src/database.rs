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
    
    // Run migrations if they exist
    // In a real implementation, this would use sqlx::migrate!() macro
    
    Ok(pool)
}

/// Create database schema for the durable engine
pub async fn create_schema(pool: &PgPool) -> Result<()> {
    info!("Creating database schema...");
    
    // Create tasks table
    sqlx::query(
        r#"
        CREATE TABLE IF NOT EXISTS tasks (
            id UUID PRIMARY KEY,
            workflow_id UUID NOT NULL,
            name VARCHAR(255) NOT NULL,
            state VARCHAR(50) NOT NULL,
            retry_count INT NOT NULL DEFAULT 0,
            max_retries INT NOT NULL DEFAULT 3,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ,
            timeout_seconds INT NOT NULL DEFAULT 3600,
            parameters JSONB NOT NULL DEFAULT '{}'::JSONB,
            result JSONB,
            error TEXT
        )
        "#
    )
    .execute(pool)
    .await?;
    
    // Create task_events table for event sourcing
    sqlx::query(
        r#"
        CREATE TABLE IF NOT EXISTS task_events (
            id UUID PRIMARY KEY,
            task_id UUID NOT NULL,
            workflow_id UUID NOT NULL,
            event_type VARCHAR(50) NOT NULL,
            previous_state VARCHAR(50),
            new_state VARCHAR(50) NOT NULL,
            timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            metadata JSONB,
            FOREIGN KEY (task_id) REFERENCES tasks(id)
        )
        "#
    )
    .execute(pool)
    .await?;
    
    // Create workflows table
    sqlx::query(
        r#"
        CREATE TABLE IF NOT EXISTS workflows (
            id UUID PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            state VARCHAR(50) NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            started_at TIMESTAMPTZ,
            completed_at TIMESTAMPTZ
        )
        "#
    )
    .execute(pool)
    .await?;
    
    // Create indexes
    sqlx::query("CREATE INDEX IF NOT EXISTS idx_tasks_workflow_id ON tasks(workflow_id)")
        .execute(pool)
        .await?;
    
    sqlx::query("CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state)")
        .execute(pool)
        .await?;
    
    sqlx::query("CREATE INDEX IF NOT EXISTS idx_task_events_task_id ON task_events(task_id)")
        .execute(pool)
        .await?;
    
    info!("Database schema created successfully");
    
    Ok(())
}
