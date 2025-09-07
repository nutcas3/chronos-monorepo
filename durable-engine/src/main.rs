mod api;
mod engine;
mod models;
mod database;
mod queue;
mod client;

use std::error::Error;
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    // Initialize tracing
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .finish();
    tracing::subscriber::set_global_default(subscriber)?;
    
    info!("Starting Durable Engine service...");
    
    // Initialize database connection
    let db_pool = database::init_db_pool().await?;
    
    // Initialize Kafka consumer
    let kafka_consumer = queue::init_kafka_consumer()?;
    
    // Start the gRPC server
    let grpc_server = api::start_grpc_server(db_pool.clone()).await?;
    
    // Start the task processor
    let engine = engine::TaskEngine::new(db_pool);
    engine.start_processing(kafka_consumer).await?;
    
    info!("Durable Engine service started successfully");
    
    // Keep the application running
    tokio::signal::ctrl_c().await?;
    info!("Shutting down Durable Engine service...");
    
    Ok(())
}
