FROM rust:1.88-slim as chef
WORKDIR /app

RUN apt-get update && \
    apt-get install -y pkg-config libssl-dev && \
    rm -rf /var/lib/apt/lists/*

RUN cargo install cargo-chef

COPY Cargo.toml Cargo.lock ./
COPY proto ./proto
RUN mkdir src && echo "fn main() {}" > src/main.rs

RUN cargo chef prepare --recipe-path recipe.json

FROM rust:1.88-slim as builder
WORKDIR /app

RUN apt-get update && \
    apt-get install -y pkg-config libssl-dev && \
    rm -rf /var/lib/apt/lists/*

RUN cargo install cargo-chef

COPY --from=chef /app/recipe.json recipe.json

RUN cargo chef cook --release --recipe-path recipe.json

COPY . .

RUN cargo build --release --bin durable-engine

FROM debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y ca-certificates libssl1.1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/target/release/durable-engine .

EXPOSE 50051

CMD ["./durable-engine"]
