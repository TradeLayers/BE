package service

import (
	"context"
	"errors"
	"testing"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/google/uuid"
)

func TestCreatePortfolio_Success(t *testing.T) {
	mock := &repository.MockPortfolioRepository{}
	svc := NewPortfolioService(mock)

	name := "My Portfolio"
	balance := 10000.0

	p, err := svc.CreatePortfolio(context.Background(), name, balance)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if p.Name != name {
		t.Errorf("expected name %q, got %q", name, p.Name)
	}
	if p.Balance != balance {
		t.Errorf("expected balance %v, got %v", balance, p.Balance)
	}
	if p.InitialBalance != balance {
		t.Errorf("expected initial balance %v, got %v", balance, p.InitialBalance)
	}
	if p.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if p.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
	if len(mock.CreateCalls) != 1 {
		t.Fatalf("expected 1 repo call, got %d", len(mock.CreateCalls))
	}
	if mock.CreateCalls[0] != p {
		t.Error("repo was called with a different portfolio pointer")
	}
}

func TestCreatePortfolio_RepoError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	mock := &repository.MockPortfolioRepository{
		CreateFn: func(_ context.Context, _ *model.Portfolio) error {
			return repoErr
		},
	}
	svc := NewPortfolioService(mock)

	p, err := svc.CreatePortfolio(context.Background(), "Test", 500.0)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if p != nil {
		t.Error("expected nil portfolio on error")
	}
}
