package account

import (
	"context"
	"github.com/yourname/apex-ledger/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	PerformTransfer(ctx context.Context, from, to string, amount int64) (string, error)
}

type Handler struct {
	api.UnimplementedLedgerServiceServer
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Transfer(ctx context.Context, req *api.TransferRequest) (*api.TransferResponse, error) {
	// 1. Basic Validation
	if req.AmountCents <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	// 2. Call Service Layer
	txID, err := h.service.PerformTransfer(ctx, req.FromAccountId, req.ToAccountId, req.AmountCents)
	if err != nil {
		// In production, map internal errors to gRPC codes
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}

	return &api.TransferResponse{
		TransactionId: txID,
		Status:        "SUCCESS",
	}, nil
}
