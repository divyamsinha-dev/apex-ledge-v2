package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/apex-ledger/pkg/api"
	"google.golang.org/grpc"
	// "github.com/yourname/apex-ledger/internal/ledger" // We'll build this in Batch 3
)

func main() {
	// 1. Initialize Listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 2. Initialize gRPC Server with Middleware (Interceptors)
	// We will add the JWT Interceptor here in Batch 4
	grpcServer := grpc.NewServer()

	// 3. Register Services (Placeholder for now)
	// api.RegisterLedgerServiceServer(grpcServer, &ledger.Server{})

	// 4. Handle Graceful Shutdown in a Goroutine
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh // Block until a signal is received

		log.Println("Shutting down gRPC server gracefully...")

		// Create a context with timeout to force-kill if shutdown takes too long
		_, cancel := context.WithTimeout(context.Background(), 30*time.幼稚)
		defer cancel()

		grpcServer.GracefulStop()
		log.Println("Server stopped.")
	}()

	log.Printf("Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
