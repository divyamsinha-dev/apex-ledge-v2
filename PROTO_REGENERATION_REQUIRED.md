# ⚠️ IMPORTANT: Proto Files Must Be Regenerated

## Problem
The code has CRUD operations implemented, but the generated proto files (`pkg/api/*.pb.go`) are outdated and only contain `Transfer` and `GetBalance` methods.

## Error You're Seeing
Compilation errors like:
- `api.CreateAccountRequest undefined`
- `api.GetAccountRequest undefined`
- `api.UpdateAccountRequest undefined`
- `api.DeleteAccountRequest undefined`
- `api.ListAccountsRequest undefined`
- Handler methods don't match gRPC interface

## Solution: Regenerate Proto Files

### Option 1: Using Make (Recommended)
```bash
cd apex-ledge-v2
make gen-proto
```

### Option 2: Manual Command
```bash
cd apex-ledge-v2

# Install protoc plugins if not installed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/ledger.proto

# Move generated files to correct location
if [ -f proto/ledger.pb.go ]; then mv proto/ledger.pb.go pkg/api/; fi
if [ -f proto/ledger_grpc.pb.go ]; then mv proto/ledger_grpc.pb.go pkg/api/; fi
```

### Option 3: Windows PowerShell
```powershell
cd apex-ledge-v2

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate proto files
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/ledger.proto

# Move files (if needed)
if (Test-Path proto/ledger.pb.go) { Move-Item proto/ledger.pb.go pkg/api/ }
if (Test-Path proto/ledger_grpc.pb.go) { Move-Item proto/ledger_grpc.pb.go pkg/api/ }
```

## After Regeneration
After running the command, the following should be generated:
- `pkg/api/ledger.pb.go` - Should contain all message types (CreateAccountRequest, GetAccountRequest, etc.)
- `pkg/api/ledger_grpc.pb.go` - Should contain all gRPC service methods (CreateAccount, GetAccount, etc.)

## Verify It Worked
Check that `pkg/api/ledger_grpc.pb.go` contains:
```go
type LedgerServiceServer interface {
    Transfer(...)
    GetBalance(...)
    CreateAccount(...)  // ← Should be here
    GetAccount(...)     // ← Should be here
    UpdateAccount(...)  // ← Should be here
    DeleteAccount(...)  // ← Should be here
    ListAccounts(...)   // ← Should be here
}
```

## Prerequisites
- `protoc` (Protocol Buffers compiler) must be installed
- Go protoc plugins installed (`protoc-gen-go`, `protoc-gen-go-grpc`)

## Then Run
```bash
go mod download
go run ./cmd/server
```

