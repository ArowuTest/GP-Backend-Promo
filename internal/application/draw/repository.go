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

// ExecuteDrawInput represents input for ExecuteDraw
type ExecuteDrawInput struct {
	DrawDate         time.Time
	PrizeStructureID uuid.UUID
	ExecutedBy       uuid.UUID
}

// ExecuteDrawOutput represents output for ExecuteDraw
type ExecuteDrawOutput struct {
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

// ExecuteDrawService handles executing a draw
type ExecuteDrawService struct {
	repository Repository
}

// NewExecuteDrawService creates a new ExecuteDrawService
func NewExecuteDrawService(repository Repository) *ExecuteDrawService {
	return &ExecuteDrawService{
		repository: repository,
	}
}

// ExecuteDraw executes a draw
func (s *ExecuteDrawService) ExecuteDraw(ctx context.Context, input ExecuteDrawInput) (ExecuteDrawOutput, error) {
	// For now, return mock data
	mockWinners := []Winner{
		{
			ID:            uuid.New(),
			DrawID:        uuid.New(),
			MSISDN:        "234*****789",
			PrizeTierID:   uuid.New(),
			PrizeTierName: "First Prize",
			PrizeValue:    1000000,
			Status:        "Active",
			PaymentStatus: "Pending",
			IsRunnerUp:    false,
			RunnerUpRank:  0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	
	return ExecuteDrawOutput{
		ID:                   uuid.New(),
		DrawDate:             input.DrawDate,
		PrizeStructureID:     input.PrizeStructureID,
		Status:               "Completed",
		TotalEligibleMSISDNs: 1000,
		TotalEntries:         5000,
		ExecutedBy:           input.ExecutedBy,
		Winners:              mockWinners,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
}

// InvokeRunnerUpInput represents input for InvokeRunnerUp
type InvokeRunnerUpInput struct {
	WinnerID  uuid.UUID
	Reason    string
	InvokedBy uuid.UUID
}

// InvokeRunnerUpOutput represents output for InvokeRunnerUp
type InvokeRunnerUpOutput struct {
	OriginalWinner Winner
	NewWinner      Winner
}

// InvokeRunnerUpService handles invoking a runner-up
type InvokeRunnerUpService struct {
	repository Repository
}

// NewInvokeRunnerUpService creates a new InvokeRunnerUpService
func NewInvokeRunnerUpService(repository Repository) *InvokeRunnerUpService {
	return &InvokeRunnerUpService{
		repository: repository,
	}
}

// InvokeRunnerUp invokes a runner-up
func (s *InvokeRunnerUpService) InvokeRunnerUp(ctx context.Context, input InvokeRunnerUpInput) (InvokeRunnerUpOutput, error) {
	// For now, return mock data
	originalWinner := Winner{
		ID:            input.WinnerID,
		DrawID:        uuid.New(),
		MSISDN:        "234*****789",
		PrizeTierID:   uuid.New(),
		PrizeTierName: "First Prize",
		PrizeValue:    1000000,
		Status:        "Disqualified",
		PaymentStatus: "Cancelled",
		IsRunnerUp:    false,
		RunnerUpRank:  0,
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		UpdatedAt:     time.Now(),
	}
	
	newWinner := Winner{
		ID:            uuid.New(),
		DrawID:        originalWinner.DrawID,
		MSISDN:        "234*****456",
		PrizeTierID:   originalWinner.PrizeTierID,
		PrizeTierName: originalWinner.PrizeTierName,
		PrizeValue:    originalWinner.PrizeValue,
		Status:        "Active",
		PaymentStatus: "Pending",
		IsRunnerUp:    true,
		RunnerUpRank:  1,
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		UpdatedAt:     time.Now(),
	}
	
	return InvokeRunnerUpOutput{
		OriginalWinner: originalWinner,
		NewWinner:      newWinner,
	}, nil
}
