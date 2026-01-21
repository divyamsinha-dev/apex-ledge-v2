package account

import (
	"context"
	"fmt"
	"strings"

	"apex-ledger/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service defines the interface for ledger operations
type Service interface {
	PerformTransfer(ctx context.Context, from, to string, amount int64) (string, error)
	GetBalance(ctx context.Context, accountID string) (*Account, error)
}

// Handler implements the gRPC LedgerService
type Handler struct {
	api.UnimplementedLedgerServiceServer
	service Service
}

// NewHandler creates a new account handler
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// Transfer handles the Transfer gRPC call
func (h *Handler) Transfer(ctx context.Context, req *api.TransferRequest) (*api.TransferResponse, error) {
	// 1. Basic Validation
	if req.FromAccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "from_account_id is required")
	}
	if req.ToAccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "to_account_id is required")
	}
	if req.AmountCents <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	if req.Currency == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is required")
	}

	// 2. Call Service Layer
	txID, err := h.service.PerformTransfer(ctx, req.FromAccountId, req.ToAccountId, req.AmountCents)
	if err != nil {
		// Map internal errors to appropriate gRPC codes
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if strings.Contains(err.Error(), "insufficient funds") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if strings.Contains(err.Error(), "currency mismatch") || strings.Contains(err.Error(), "cannot be empty") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}

	return &api.TransferResponse{
		TransactionId: txID,
		Status:        "SUCCESS",
	}, nil
}

// GetBalance handles the GetBalance gRPC call
func (h *Handler) GetBalance(ctx context.Context, req *api.BalanceRequest) (*api.BalanceResponse, error) {
	// 1. Basic Validation
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	// 2. Call Service Layer
	acc, err := h.service.GetBalance(ctx, req.AccountId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("account %s not found", req.AccountId))
		}
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	return &api.BalanceResponse{
		BalanceCents: acc.BalanceCents,
		Currency:     acc.Currency,
	}, nil
}
