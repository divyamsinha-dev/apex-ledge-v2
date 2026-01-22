.PHONY: gen-proto build run test clean migrate

# Generate proto files
gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/ledger.proto
	@if [ -f proto/ledger.pb.go ]; then mv proto/ledger.pb.go pkg/api/; fi
	@if [ -f proto/ledger_grpc.pb.go ]; then mv proto/ledger_grpc.pb.go pkg/api/; fi

# Build the server
build:
	go build -o bin/server ./cmd/server

# Run the server
run:
	go run ./cmd/server

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run database migrations (requires psql or similar)
migrate:
	@echo "Run migrations manually using psql or your migration tool"
	@echo "Example: psql -d ledger -f migrations/001_create_schema.sql"

# Install dependencies
deps:
	go mod download
	go mod tidy
