# Database Migrations Setup

This project uses SQLx embedded migrations for safe schema evolution.

## Quick Start

1. **Install sqlx-cli** (if not already installed):
   ```bash
   cargo install sqlx-cli
   ```

2. **Set your database URL**:
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost/chronos"
   ```

3. **Run the application** - migrations run automatically:
   ```bash
   cargo run
   ```

## Migration Management

### Creating New Migrations

```bash
sqlx migrate add your_migration_name

```

### Manual Migration Commands

```bash
sqlx migrate run

sqlx migrate revert

sqlx migrate info
```

## Compile-Time Query Checking

This project uses `sqlx::query!()` macros for type-safe SQL queries that are checked at compile time.

### Offline Compilation Support

For CI/CD environments without database access:

```bash
cargo sqlx prepare

cargo build
```

### Environment Variables

```bash
export SQLX_OFFLINE=true

export SQLX_OFFLINE=false
```

## Example Usage

```rust
use crate::database::{get_task_by_id, update_task_state};

let task = get_task_by_id(&pool, task_id).await?;

update_task_state(&pool, task_id, "completed").await?;
```

## Benefits

- ✅ **Safe schema evolution** - versioned migrations
- ✅ **Compile-time SQL validation** - catch errors before runtime  
- ✅ **Type safety** - automatic Rust type inference from SQL
- ✅ **Zero-cost abstractions** - no runtime SQL parsing
- ✅ **Offline compilation** - works in CI/CD without database
