# Project Chronos

A highly scalable and fault-tolerant distributed workflow engine for defining, scheduling, and executing complex data and task pipelines.

## Architecture Overview

Project Chronos is a modern re-imagination of a workflow orchestrator, built from scratch to handle extreme scale and guarantee durability. It consists of five interconnected microservices:

1. **Scheduler (Go)**: The brain of the operation, responsible for parsing workflow definitions, managing schedules, and triggering new runs.
2. **Executor (Go)**: A stateless, highly-scalable service that acts as the entry point for executing individual task runs.
3. **Durable Engine (Rust)**: The most critical component - a state machine that guarantees the persistence and correct execution of individual tasks.
4. **Worker Pool (Go)**: A pool of external workers that do the actual work defined in the tasks.
5. **Observatory (Go)**: The observability hub that collects metrics, logs, and traces from all other microservices.

## Getting Started

### Prerequisites

- Go 1.24+
- Rust 1.80+
- Docker and Docker Compose
- Kafka
- PostgreSQL
- Redis

### Local Development

```bash
# Clone the repository
git clone https://github.com/nutcas3/chronos-monorepo
cd chronos-monorepo

# Start all services and dependencies
docker-compose up
```

## Project Structure

```
/chronos-monorepo
├── scheduler/         # Go-based scheduler service
├── executor/          # Go-based executor service
├── durable-engine/    # Rust-based durable engine
├── worker-pool/       # Go-based worker pool
├── observatory/       # Go-based observability service
├── proto/             # gRPC Protocol Buffer definitions
├── clients/           # Client libraries for Go and Rust
└── scripts/           # Build and deployment scripts
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
