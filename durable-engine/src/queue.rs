use anyhow::Result;
use rdkafka::client::ClientContext;
use rdkafka::config::{ClientConfig, RDKafkaLogLevel};
use rdkafka::consumer::{ConsumerContext, Rebalance, StreamConsumer};
use rdkafka::error::KafkaResult;
use rdkafka::topic_partition_list::TopicPartitionList;
use std::env;
use tracing::{info, warn};

// A context can be used to change the behavior of producers and consumers by adding callbacks
// that will be executed by librdkafka.
struct CustomContext;

impl ClientContext for CustomContext {}

impl ConsumerContext for CustomContext {
    fn pre_rebalance(&self, rebalance: &Rebalance) {
        info!("Pre rebalance: {:?}", rebalance);
    }

    fn post_rebalance(&self, rebalance: &Rebalance) {
        info!("Post rebalance: {:?}", rebalance);
    }

    fn commit_callback(&self, result: KafkaResult<()>, _offsets: &TopicPartitionList) {
        match result {
            Ok(_) => info!("Offsets committed successfully"),
            Err(e) => warn!("Error while committing offsets: {}", e),
        }
    }
}

type LoggingConsumer = StreamConsumer<CustomContext>;

/// Initialize the Kafka consumer
pub fn init_kafka_consumer() -> Result<LoggingConsumer> {
    let brokers = env::var("KAFKA_BROKERS").unwrap_or_else(|_| "localhost:9092".to_string());
    let group_id = env::var("KAFKA_GROUP_ID").unwrap_or_else(|_| "chronos-durable-engine".to_string());
    let topic = env::var("KAFKA_TOPIC").unwrap_or_else(|_| "chronos-tasks".to_string());

    info!("Connecting to Kafka at {}", brokers);

    // Create the consumer
    let consumer: LoggingConsumer = ClientConfig::new()
        .set("group.id", &group_id)
        .set("bootstrap.servers", &brokers)
        .set("enable.auto.commit", "true")
        .set("auto.offset.reset", "earliest")
        .set_log_level(RDKafkaLogLevel::Debug)
        .create_with_context(CustomContext)
        .expect("Consumer creation failed");

    // Subscribe to the topic
    consumer
        .subscribe(&[&topic])
        .expect("Can't subscribe to specified topic");

    info!("Kafka consumer initialized and subscribed to {}", topic);

    Ok(consumer)
}
