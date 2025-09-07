use anyhow::Result;
use chronos_client::{ChronosClient, ClientOptions};

#[tokio::main]
async fn main() -> Result<()> {
    // Create a client with default options
    let client = ChronosClient::new(ClientOptions::default()).await?;
    
    // Create a new workflow
    let workflow = client.create_workflow("Example Workflow", "An example workflow").await?;
    println!("Created workflow: {}", workflow.id);
    
    // Add a task to the workflow
    let task = client.add_task(
        &workflow.id,
        "Example Task",
        "http",
        serde_json::to_vec(&serde_json::json!({
            "url": "https://example.com",
            "method": "GET"
        }))?,
    ).await?;
    println!("Added task: {}", task.id);
    
    // Start the workflow
    client.start_workflow(&workflow.id).await?;
    println!("Started workflow");
    
    // Get the workflow
    let workflow = client.get_workflow(&workflow.id).await?;
    println!("Retrieved workflow: {}", workflow.name);
    
    // Get the task
    let task = client.get_task(&task.id).await?;
    println!("Retrieved task: {} (status: {})", task.name, task.status);
    
    Ok(())
}
