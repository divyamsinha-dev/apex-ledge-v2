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
	CreateAccount(ctx context.Context, id string, balanceCents int64, currency string) (*Account, error)
	GetAccount(ctx context.Context, accountID string) (*Account, error)
	UpdateAccount(ctx context.Context, accountID string, currency string) (*Account, error)
	DeleteAccount(ctx context.Context, accountID string) error
	ListAccounts(ctx context.Context, limit, offset int) ([]Account, int, error)
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

// CreateAccount handles the CreateAccount gRPC call
func (h *Handler) CreateAccount(ctx context.Context, req *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {
	// Validation
	if req.Currency == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is required")
	}

	// Set defaults
	id := req.Id // If empty, service will generate UUID
	balanceCents := req.InitialBalanceCents
	if balanceCents < 0 {
		return nil, status.Error(codes.InvalidArgument, "initial balance cannot be negative")
	}

	// Call service
	acc, err := h.service.CreateAccount(ctx, id, balanceCents, req.Currency)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to create account: %v", err)
	}

	return &api.CreateAccountResponse{
		AccountId:     acc.ID,
		BalanceCents: acc.BalanceCents,
		Currency:     acc.Currency,
		Status:       "CREATED",
	}, nil
}

// GetAccount handles the GetAccount gRPC call
func (h *Handler) GetAccount(ctx context.Context, req *api.GetAccountRequest) (*api.GetAccountResponse, error) {
	// Validation
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	// Call service
	acc, err := h.service.GetAccount(ctx, req.AccountId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("account %s not found", req.AccountId))
		}
		return nil, status.Errorf(codes.Internal, "failed to get account: %v", err)
	}

	return &api.GetAccountResponse{
		AccountId:     acc.ID,
		BalanceCents:  acc.BalanceCents,
		Currency:      acc.Currency,
		CreatedAt:     acc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     acc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// UpdateAccount handles the UpdateAccount gRPC call
func (h *Handler) UpdateAccount(ctx context.Context, req *api.UpdateAccountRequest) (*api.UpdateAccountResponse, error) {
	// Validation
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}
	if req.Currency == "" {
		return nil, status.Error(codes.InvalidArgument, "currency is required")
	}

	// Call service
	acc, err := h.service.UpdateAccount(ctx, req.AccountId, req.Currency)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to update account: %v", err)
	}

	return &api.UpdateAccountResponse{
		AccountId: acc.ID,
		Currency:  acc.Currency,
		Status:    "UPDATED",
	}, nil
}

// DeleteAccount handles the DeleteAccount gRPC call
func (h *Handler) DeleteAccount(ctx context.Context, req *api.DeleteAccountRequest) (*api.DeleteAccountResponse, error) {
	// Validation
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	// Call service
	err := h.service.DeleteAccount(ctx, req.AccountId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete account: %v", err)
	}

	return &api.DeleteAccountResponse{
		AccountId: req.AccountId,
		Status:    "DELETED",
	}, nil
}

// ListAccounts handles the ListAccounts gRPC call
func (h *Handler) ListAccounts(ctx context.Context, req *api.ListAccountsRequest) (*api.ListAccountsResponse, error) {
	// Set defaults
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Call service
	accounts, total, err := h.service.ListAccounts(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %v", err)
	}

	// Convert to response
	accountResponses := make([]*api.GetAccountResponse, len(accounts))
	for i, acc := range accounts {
		accountResponses[i] = &api.GetAccountResponse{
			AccountId:    acc.ID,
			BalanceCents: acc.BalanceCents,
			Currency:     acc.Currency,
			CreatedAt:    acc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    acc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &api.ListAccountsResponse{
		Accounts: accountResponses,
		Total:    int32(total),
	}, nil
}
