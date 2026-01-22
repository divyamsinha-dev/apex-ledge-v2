# Apex Ledger

A production-ready, high-performance **double-entry ledger system** built with Go, gRPC, and PostgreSQL. This system ensures financial transaction integrity and prevents race conditions in concurrent environments.

---

## 🎯 What is This Project?

**Apex Ledger** is a **financial ledger microservice** that implements double-entry bookkeeping principles. It provides a secure, scalable API for managing account balances and processing money transfers between accounts while maintaining strict accounting integrity.

### Real-World Use Cases:
- **Banking Systems**: Core transaction processing
- **Payment Gateways**: Fund transfers between accounts
- **E-commerce Platforms**: Wallet management
- **Financial Applications**: Account balance tracking

---

## 🔥 What Problem Does It Solve?

### **Problem 1: Race Conditions in Concurrent Transactions**
**Challenge**: When multiple transfers happen simultaneously on the same account, traditional systems can lose money or create inconsistent balances.

**Solution**: 
- **Pessimistic Locking** using `SELECT FOR UPDATE` to lock accounts during transfers
- **Ordered Locking** (alphabetical) to prevent deadlocks
- **Database Transactions** ensure atomicity

**Example Scenario**:
```
Account A has $100
- Transfer 1: A → B ($50) starts
- Transfer 2: A → C ($60) starts simultaneously
Without locking: Both might succeed, leaving negative balance!
With locking: One waits, ensuring only valid transfers succeed.
```

### **Problem 2: Data Integrity in Financial Systems**
**Challenge**: Financial systems must never lose or duplicate money. Every debit must have a corresponding credit.

**Solution**:
- **Double-Entry Bookkeeping**: Every transfer debits one account and credits another atomically
- **ACID Transactions**: All-or-nothing execution
- **Audit Trail**: All transactions recorded in `transactions` table

### **Problem 3: Scalability & Performance**
**Challenge**: Financial APIs need to handle high throughput with low latency.

**Solution**:
- **gRPC**: Binary protocol, faster than REST/JSON
- **Connection Pooling**: Efficient database connection management
- **Async Workers**: Background processing for notifications

### **Problem 4: Security**
**Challenge**: Financial APIs must be secure and authenticated.

**Solution**:
- **JWT Authentication**: All requests validated via interceptor
- **Algorithm Validation**: Prevents JWT algorithm confusion attacks
- **Secure by Default**: No unauthenticated endpoints

---

## 🏗️ Architecture Overview

### **Clean Architecture Layers**

```
┌─────────────────────────────────────────┐
│         gRPC Handler Layer               │  ← API Interface (HTTP/gRPC)
│    (Request Validation & Mapping)       │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Service Layer                    │  ← Business Logic
│    (Transfer Logic, Validation)          │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Repository Layer                 │  ← Data Access
│    (Database Queries, Transactions)      │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         PostgreSQL Database             │  ← Data Persistence
│    (ACID Transactions, Locking)         │
└─────────────────────────────────────────┘
```

### **Key Design Patterns**

1. **Repository Pattern**: Abstracts database operations
2. **Service Layer Pattern**: Encapsulates business logic
3. **Dependency Injection**: Loose coupling between layers
4. **Interceptor Pattern**: Cross-cutting concerns (auth)

---

## 🔄 Code Flow: How a Transfer Works

### **Step-by-Step Flow**

```
1. Client Request
   ↓
   [gRPC Client] → TransferRequest (from_account_id, to_account_id, amount)
   ↓
   
2. Authentication (Interceptor)
   ↓
   [AuthInterceptor] → Validates JWT token from metadata
   ↓
   ✓ Token valid → Continue
   ✗ Token invalid → Return Unauthenticated error
   ↓
   
3. Handler Layer
   ↓
   [Handler.Transfer()] → Validates request fields
   - Checks: account IDs not empty, amount > 0, currency present
   ↓
   
4. Service Layer (Business Logic)
   ↓
   [LedgerService.PerformTransfer()]
   ├─ Validates inputs (same account check, positive amount)
   ├─ Generates transaction ID (UUID)
   ├─ Starts database transaction
   ├─ Locks accounts in alphabetical order (prevents deadlock)
   │  └─ GetAccountWithLock() with SELECT FOR UPDATE
   ├─ Validates currency match
   ├─ Checks sufficient funds
   ├─ Performs double-entry:
   │  ├─ Debit: UpdateBalance(fromID, -amount)
   │  └─ Credit: UpdateBalance(toID, +amount)
   ├─ Records transaction in ledger table
   └─ Commits transaction (or rolls back on error)
   ↓
   
5. Repository Layer
   ↓
   [Repository Methods]
   ├─ GetAccountWithLock() → SELECT ... FOR UPDATE
   ├─ UpdateBalance() → UPDATE accounts SET balance_cents = ...
   └─ recordTransaction() → INSERT INTO transactions
   ↓
   
6. Database
   ↓
   [PostgreSQL]
   ├─ Locks rows during SELECT FOR UPDATE
   ├─ Executes updates atomically
   └─ Commits transaction
   ↓
   
7. Response
   ↓
   [Handler] → TransferResponse (transaction_id, status)
   ↓
   [gRPC Client] ← Success response
```

### **Critical Flow: Deadlock Prevention**

```go
// Always lock in alphabetical order
if fromID < toID {
    lock(fromID)  // Lock account A first
    lock(toID)    // Then lock account B
} else {
    lock(toID)    // Lock account B first
    lock(fromID)  // Then lock account A
}
```

**Why?** If Transfer A→B and Transfer B→A happen simultaneously:
- Without ordering: Deadlock! (A waits for B, B waits for A)
- With ordering: Both lock A first, then B → No deadlock!

---

## 📁 Project Structure

```
apex-ledge-v2/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point, server setup
│
├── internal/                     # Private application code
│   ├── account/
│   │   ├── handler.go           # gRPC handlers (API layer)
│   │   ├── repository.go        # Database operations (data layer)
│   │   ├── model.go             # Account data structures
│   │   └── worker.go            # Async notification workers
│   │
│   ├── auth/
│   │   └── interceptor.go       # JWT authentication middleware
│   │
│   ├── service/
│   │   └── ledger.go            # Business logic (service layer)
│   │
│   ├── config/
│   │   └── config.go             # Configuration management
│   │
│   └── platform/
│       └── database/
│           └── postgress.go      # Database connection & pooling
│
├── pkg/
│   └── api/                      # Generated gRPC code
│       ├── ledger.pb.go
│       └── ledger_grpc.pb.go
│
├── proto/
│   └── ledger.proto              # gRPC service definitions
│
├── migrations/                   # Database schema
│   ├── 001_create_schema.sql
│   └── 002_insert_sample_data.sql
│
├── deployments/                  # Deployment configs
│   ├── Dockerfile
│   ├── k8s-deployment.yaml
│   └── k8s-configmap.yaml
│
├── go.mod                        # Go dependencies
├── Makefile                      # Build commands
└── README.md                     # This file
```

---

## 🚀 Key Features

### **1. Double-Entry Bookkeeping**
Every transfer ensures:
- **Debit** from source account
- **Credit** to destination account
- **Atomic**: Both succeed or both fail
- **Audit Trail**: Recorded in transactions table

### **2. Race Condition Prevention**
- **Pessimistic Locking**: `SELECT FOR UPDATE` locks rows
- **Ordered Locking**: Prevents deadlocks
- **Transaction Isolation**: ACID guarantees

### **3. Complete CRUD Operations**
- ✅ **CreateAccount**: Create new accounts with initial balance
- ✅ **GetAccount**: Retrieve full account details
- ✅ **UpdateAccount**: Update account currency
- ✅ **DeleteAccount**: Remove accounts
- ✅ **ListAccounts**: Paginated listing
- ✅ **Transfer**: Double-entry transfers
- ✅ **GetBalance**: Quick balance check

### **4. Security**
- **JWT Authentication**: All endpoints protected
- **Algorithm Validation**: Prevents JWT attacks
- **Secure by Default**: No unauthenticated access

### **5. Scalability**
- **gRPC**: High-performance binary protocol
- **Connection Pooling**: Efficient DB connections
- **Async Workers**: Background task processing
- **Graceful Shutdown**: Clean server termination

---

## 💻 Technical Highlights (For Interviewers)

### **1. Concurrency Safety**
```go
// Prevents race conditions with pessimistic locking
SELECT id, balance_cents FROM accounts WHERE id = $1 FOR UPDATE
```
- **Why FOR UPDATE?** Locks row until transaction commits
- **Why in transaction?** Ensures atomicity
- **Why ordered locking?** Prevents deadlocks

### **2. Error Handling**
- **Layered Error Mapping**: Repository → Service → Handler
- **gRPC Status Codes**: Proper error codes (NotFound, InvalidArgument, etc.)
- **Error Wrapping**: Context preserved with `fmt.Errorf("...: %w", err)`

### **3. Database Design**
- **Cents Storage**: Avoids floating-point precision issues
- **Foreign Keys**: Referential integrity
- **Indexes**: Optimized queries
- **Triggers**: Auto-update timestamps

### **4. Clean Architecture**
- **Separation of Concerns**: Handler → Service → Repository
- **Dependency Inversion**: Service depends on Repository interface
- **Testability**: Each layer can be tested independently

### **5. Production Readiness**
- **Connection Pooling**: Prevents connection exhaustion
- **Graceful Shutdown**: Handles SIGTERM/SIGINT
- **Configuration**: Environment-based config
- **Logging**: Structured logging throughout

---

## 🔧 Setup & Installation

### Prerequisites
- **Go 1.21+**
- **PostgreSQL 12+**
- **protoc** (Protocol Buffers compiler)
- **protoc-gen-go** and **protoc-gen-go-grpc** plugins

### ⚠️ CRITICAL: Regenerate Proto Files First!

**The generated proto files are outdated and MUST be regenerated before running!**

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate gRPC code (REQUIRED!)
make gen-proto
# OR manually:
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/ledger.proto
```

**Without this step, you'll get compilation errors!** See `PROTO_REGENERATION_REQUIRED.md` for details.

### Quick Start

```bash
# 1. Install dependencies
go mod download

# 2. ⚠️ Generate gRPC code (REQUIRED - DO THIS FIRST!)
make gen-proto

# 3. Set up database
createdb ledger
psql -d ledger -f migrations/001_create_schema.sql
psql -d ledger -f migrations/002_insert_sample_data.sql

# 4. Configure (optional)
export DB_URL="postgres://user:pass@localhost:5432/ledger?sslmode=disable"
export GRPC_PORT="50051"
export JWT_SECRET="your-secret-key"
export WORKER_COUNT="5"

# 5. Run server
make run
# or
go run ./cmd/server
```

---

## 📡 API Endpoints

### **Transfer Funds**
```protobuf
rpc Transfer(TransferRequest) returns (TransferResponse)
```
- Debits source account, credits destination
- Validates currency match and sufficient funds
- Returns transaction ID

### **Get Balance**
```protobuf
rpc GetBalance(BalanceRequest) returns (BalanceResponse)
```
- Quick balance check
- Returns balance in cents and currency

### **CRUD Operations**
- `CreateAccount`: Create with initial balance
- `GetAccount`: Full account details with timestamps
- `UpdateAccount`: Update currency
- `DeleteAccount`: Remove account
- `ListAccounts`: Paginated listing (limit/offset)

---

## 🔐 Authentication

All requests require JWT token in gRPC metadata:

```go
md := metadata.New(map[string]string{
    "authorization": "Bearer <jwt-token>",
})
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

The `AuthInterceptor` validates:
1. Metadata presence
2. Authorization header
3. JWT signature (HMAC)
4. Token validity

---

## 🗄️ Database Schema

### **accounts** Table
```sql
CREATE TABLE accounts (
    id VARCHAR(255) PRIMARY KEY,
    balance_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### **transactions** Table
```sql
CREATE TABLE transactions (
    id VARCHAR(255) PRIMARY KEY,
    from_account_id VARCHAR(255) NOT NULL,
    to_account_id VARCHAR(255) NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (from_account_id) REFERENCES accounts(id),
    FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);
```

**Key Points**:
- **balance_cents**: Stored as integers (avoids float precision issues)
- **Foreign Keys**: Ensures referential integrity
- **Indexes**: On foreign keys and created_at for performance

---

## 🎓 Learning Points / Interview Talking Points

### **1. Why Double-Entry Bookkeeping?**
- **Accounting Standard**: Industry-standard for financial systems
- **Error Detection**: Imbalance indicates errors
- **Audit Trail**: Complete transaction history
- **Integrity**: Can't lose or duplicate money

### **2. Why Pessimistic Locking?**
- **Guarantees**: Strong consistency guarantees
- **Prevents**: Race conditions in concurrent systems
- **Trade-off**: Slightly slower but safer than optimistic locking

### **3. Why gRPC over REST?**
- **Performance**: Binary protocol, faster than JSON
- **Type Safety**: Strongly typed with Protocol Buffers
- **Streaming**: Built-in support for streaming
- **Code Generation**: Auto-generated client/server code

### **4. Why Clean Architecture?**
- **Maintainability**: Easy to modify and extend
- **Testability**: Each layer testable independently
- **Flexibility**: Can swap implementations (e.g., different DB)
- **Scalability**: Clear boundaries for microservices

### **5. Production Considerations**
- **Connection Pooling**: Prevents DB connection exhaustion
- **Graceful Shutdown**: Handles in-flight requests
- **Error Handling**: Proper error codes and messages
- **Monitoring**: Logging and metrics ready
- **Security**: JWT validation, input sanitization

---

## 🚢 Deployment

### Docker
```bash
docker build -t apex-ledger -f deployments/Dockerfile .
docker run -p 50051:50051 apex-ledger
```

### Kubernetes
See `deployments/k8s-deployment.yaml` for K8s configuration.

---

## 📊 Performance Considerations

- **Connection Pool**: Max 25 connections, prevents exhaustion
- **Query Optimization**: Indexed foreign keys
- **Binary Protocol**: gRPC faster than REST
- **Async Workers**: Background processing doesn't block API

---

## 🔍 Testing

```bash
# Run all tests
make test

# Or
go test ./...
```

---

## 📝 License

[Add your license here]

---

## 👨‍💻 Author Notes

This project demonstrates:
- **Production-ready** Go microservice architecture
- **Financial system** best practices
- **Concurrency** handling in distributed systems
- **Clean architecture** principles
- **Security** considerations for APIs

Perfect for demonstrating understanding of:
- System design
- Database transactions
- Concurrency control
- API design
- Security practices
