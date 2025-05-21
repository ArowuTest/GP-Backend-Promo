package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// Repository defines the interface for draw repository
type Repository interface {
	GetDrawByID(ctx context.Context, id uuid.UUID) (Draw, []Winner, error)
	ListWinners(ctx context.Context, page, pageSize int, startDate, endDate string) ([]Winner, int, error)
	UpdateWinnerPaymentStatus(ctx context.Context, winnerID uuid.UUID, paymentStatus, notes string, updatedBy uuid.UUID) (Winner, error)
	ExecuteDraw(ctx context.Context, drawDate time.Time, prizeStructureID, executedBy uuid.UUID) (Draw, []Winner, error)
	InvokeRunnerUp(ctx context.Context, winnerID uuid.UUID, reason string, invokedBy uuid.UUID) (Winner, Winner, error)
}

// Winner represents a draw winner
type Winner struct {
	ID            uuid.UUID
	DrawID        uuid.UUID
	MSISDN        string
	PrizeTierID   uuid.UUID
	PrizeTierName string
	PrizeValue    float64
	Status        string
	PaymentStatus string
	PaymentNotes  string
	PaidAt        string
	IsRunnerUp    bool
	RunnerUpRank  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Draw represents a draw
type Draw struct {
	ID                   uuid.UUID
	DrawDate             time.Time
	PrizeStructureID     uuid.UUID
	Status               string
	TotalEligibleMSISDNs int
	TotalEntries         int
	ExecutedBy           uuid.UUID
	Winners              []Winner
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
