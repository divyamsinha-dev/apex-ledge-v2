# Apex Ledger

A production-ready gRPC-based double-entry ledger system built with Go.

## Features

- **Double-Entry Bookkeeping**: Ensures all transactions maintain accounting integrity
- **gRPC API**: High-performance RPC interface for ledger operations
- **PostgreSQL Backend**: Robust database with ACID guarantees
- **Transaction Safety**: Pessimistic locking prevents race conditions
- **JWT Authentication**: Secure API access with JWT tokens
- **Async Notifications**: Worker pool for handling async tasks
- **Graceful Shutdown**: Clean server shutdown handling

## Architecture

```
cmd/server/          - Application entry point
internal/
  account/           - Account domain (handler, repository, model)
  auth/              - JWT authentication interceptor
  config/            - Configuration management
  service/           - Business logic layer
  platform/
    database/        - Database connection and utilities
pkg/api/             - Generated gRPC code from proto files
proto/               - Protocol buffer definitions
migrations/          - Database schema migrations
deployments/         - Docker and Kubernetes configurations
```

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Protocol Buffers compiler (`protoc`)
- Go plugins for protoc:
  - `protoc-gen-go`
  - `protoc-gen-go-grpc`

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd apex-ledge
```

2. Install dependencies:
```bash
go mod download
```

3. Install protoc plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

4. Generate gRPC code:
```bash
make gen-proto
```

5. Set up the database:
```bash
# Create database
createdb ledger

# Run migrations
psql -d ledger -f migrations/001_create_schema.sql
psql -d ledger -f migrations/002_insert_sample_data.sql
```

## Configuration

The application uses environment variables for configuration:

- `DB_URL`: PostgreSQL connection string (default: `postgres://user:pass@localhost:5432/ledger?sslmode=disable`)
- `GRPC_PORT`: gRPC server port (default: `50051`)
- `JWT_SECRET`: Secret key for JWT validation (default: `production-secret-key`)
- `WORKER_COUNT`: Number of async worker goroutines (default: `5`)

Example:
```bash
export DB_URL="postgres://user:password@localhost:5432/ledger?sslmode=disable"
export GRPC_PORT="50051"
export JWT_SECRET="your-secret-key-here"
export WORKER_COUNT="10"
```

## Running the Server

```bash
# Using make
make run

# Or directly
go run ./cmd/server
```

## Building

```bash
make build
# Binary will be in bin/server
```

## API Usage

### Transfer Funds

```go
// Example gRPC client call
req := &api.TransferRequest{
    FromAccountId: "account-001",
    ToAccountId:   "account-002",
    AmountCents:   10000,  // $100.00
    Currency:      "USD",
}

resp, err := client.Transfer(ctx, req)
```

### Get Balance

```go
req := &api.BalanceRequest{
    AccountId: "account-001",
}

resp, err := client.GetBalance(ctx, req)
// resp.BalanceCents contains the balance
```

## Authentication

All gRPC requests require a JWT token in the metadata:

```go
md := metadata.New(map[string]string{
    "authorization": "Bearer <your-jwt-token>",
})
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

## Database Schema

### Accounts Table
- `id`: Account identifier (VARCHAR, PRIMARY KEY)
- `balance_cents`: Account balance in cents (BIGINT)
- `currency`: Currency code (VARCHAR)
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp (auto-updated)

### Transactions Table
- `id`: Transaction identifier (VARCHAR, PRIMARY KEY)
- `from_account_id`: Source account (VARCHAR, FOREIGN KEY)
- `to_account_id`: Destination account (VARCHAR, FOREIGN KEY)
- `amount_cents`: Transfer amount in cents (BIGINT)
- `currency`: Currency code (VARCHAR)
- `created_at`: Transaction timestamp

## Development

### Running Tests
```bash
make test
```

### Code Generation
```bash
# Regenerate proto files
make gen-proto
```

### Database Migrations
```bash
# Apply migrations manually
psql -d ledger -f migrations/001_create_schema.sql
```

## Deployment

### Docker
```bash
docker build -t apex-ledger -f deployments/Dockerfile .
docker run -p 50051:50051 apex-ledger
```

### Kubernetes
See `deployments/k8s-deployment.yaml` and `deployments/k8s-configmap.yaml` for Kubernetes configuration.

## Project Structure

- **Clean Architecture**: Separation of concerns with clear boundaries
- **Domain-Driven Design**: Account domain with proper encapsulation
- **Repository Pattern**: Data access abstraction
- **Service Layer**: Business logic separation
- **gRPC Handlers**: API layer implementation

## License

[Add your license here]
