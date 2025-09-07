use crate::models::{Task, TaskEvent, TaskState};
use anyhow::{Context, Result};
use rdkafka::consumer::{Consumer, StreamConsumer};
use sqlx::PgPool;
use std::sync::Arc;
use tokio::sync::Mutex;
use tracing::{error, info, warn};
use uuid::Uuid;

pub struct TaskEngine {
    db_pool: PgPool,
    active_tasks: Arc<Mutex<Vec<Uuid>>>,
}

impl TaskEngine {
    pub fn new(db_pool: PgPool) -> Self {
        Self {
            db_pool,
            active_tasks: Arc::new(Mutex::new(Vec::new())),
        }
    }

    /// Start processing tasks from the Kafka queue
    pub async fn start_processing(&self, consumer: StreamConsumer) -> Result<()> {
        info!("Starting task processing loop");
        
        // Start the reconciliation loop in a separate task
        let db_pool_clone = self.db_pool.clone();
        let active_tasks_clone = self.active_tasks.clone();
        tokio::spawn(async move {
            if let Err(e) = Self::run_reconciliation_loop(db_pool_clone, active_tasks_clone).await {
                error!("Reconciliation loop failed: {:?}", e);
            }
        });
        
        // Main processing loop
        // In a real implementation, this would consume messages from Kafka
        // and process them
        
        Ok(())
    }
    
    /// Process a single task
    async fn process_task(&self, task_id: Uuid) -> Result<()> {
        // 1. Lock the task in the database
        // 2. Update its state to RUNNING
        // 3. Execute the task logic
        // 4. Update the state based on the result
        // 5. Store any task events
        
        // This is a simplified implementation
        let mut tx = self.db_pool.begin().await?;
        
        // Update task state to RUNNING
        let task = sqlx::query_as!(
            Task,
            "UPDATE tasks SET state = $1, updated_at = NOW(), started_at = NOW() 
             WHERE id = $2 AND state = $3
             RETURNING *",
            TaskState::Running.to_string(),
            task_id,
            TaskState::Queued.to_string()
        )
        .fetch_one(&mut *tx)
        .await
        .context("Failed to update task state to RUNNING")?;
        
        // Record the state change event
        let event_id = Uuid::new_v4();
        sqlx::query!(
            "INSERT INTO task_events (id, task_id, workflow_id, event_type, previous_state, new_state, timestamp)
             VALUES ($1, $2, $3, $4, $5, $6, NOW())",
            event_id,
            task.id,
            task.workflow_id,
            "STATE_CHANGE",
            Some(TaskState::Queued.to_string()),
            TaskState::Running.to_string()
        )
        .execute(&mut *tx)
        .await
        .context("Failed to record task event")?;
        
        tx.commit().await?;
        
        // Add to active tasks
        {
            let mut active_tasks = self.active_tasks.lock().await;
            active_tasks.push(task_id);
        }
        
        // In a real implementation, this would communicate with the worker
        // and handle timeouts, retries, etc.
        
        Ok(())
    }
    
    /// Reconciliation loop to find and fix "stuck" tasks
    async fn run_reconciliation_loop(
        db_pool: PgPool,
        active_tasks: Arc<Mutex<Vec<Uuid>>>
    ) -> Result<()> {
        let interval = tokio::time::Duration::from_secs(60); // Run every minute
        
        loop {
            tokio::time::sleep(interval).await;
            
            // Find tasks that have been in RUNNING state for too long
            let stuck_tasks = sqlx::query!(
                "SELECT id FROM tasks 
                 WHERE state = $1 
                 AND started_at < NOW() - INTERVAL '1 hour'",
                TaskState::Running.to_string()
            )
            .fetch_all(&db_pool)
            .await?;
            
            for task in stuck_tasks {
                warn!("Found stuck task: {}", task.id);
                
                // In a real implementation, this would reset the task
                // and potentially retry it
            }
        }
    }
}
