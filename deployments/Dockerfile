# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make protobuf protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make gen-proto

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 50051

CMD ["./server"]
