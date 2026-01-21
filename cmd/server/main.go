package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apex-ledger/internal/account"
	"apex-ledger/internal/auth"
	"apex-ledger/internal/config"
	"apex-ledger/internal/platform/database"
	"apex-ledger/internal/service"
	"apex-ledger/pkg/api"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting server with config: GRPC_PORT=%s, DB_URL=%s", cfg.GRPCPort, maskDBURL(cfg.DBURL))

	// Initialize database connection
	db, err := database.NewPostgres(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// Initialize repositories
	accountRepo := account.NewRepository(db)

	// Initialize services
	ledgerService := service.NewLedgerService(accountRepo, db)

	// Initialize handlers
	accountHandler := account.NewHandler(ledgerService)

	// Initialize worker pool for async notifications
	workerPool := account.NewNotificationWorkerPool(100)
	workerPool.Start(cfg.WorkerCount)
	log.Printf("Started %d notification workers", cfg.WorkerCount)

	// Initialize gRPC server with auth interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.AuthInterceptor(cfg.JWTSecret)),
	)

	// Register gRPC services
	api.RegisterLedgerServiceServer(grpcServer, accountHandler)

	// Start listening
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh // Block until a signal is received

		log.Println("Shutting down gRPC server gracefully...")

		// Create a context with timeout to force-kill if shutdown takes too long
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Gracefully stop the server
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Println("Server stopped gracefully")
		case <-ctx.Done():
			log.Println("Shutdown timeout exceeded, forcing stop")
			grpcServer.Stop()
		}
	}()

	log.Printf("Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// maskDBURL masks sensitive information in database URL for logging
func maskDBURL(url string) string {
	// Simple masking - in production, use a proper URL parser
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-10:]
	}
	return "***"
}
