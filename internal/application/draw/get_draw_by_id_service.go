package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// GetDrawByIDInput represents input for GetDrawByID
type GetDrawByIDInput struct {
	ID uuid.UUID
}

// GetDrawByIDOutput represents output for GetDrawByID
type GetDrawByIDOutput struct {
	ID                   uuid.UUID
	DrawDate             time.Time
	PrizeStructureID     uuid.UUID
	Status               string
	TotalEligibleMSISDNs int
	TotalEntries         int
	ExecutedBy           uuid.UUID
	Winners              []drawDomain.Winner
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// GetDrawByIDService handles retrieving a draw by ID
type GetDrawByIDService struct {
	repository Repository
}

// NewGetDrawByIDService creates a new GetDrawByIDService
func NewGetDrawByIDService(repository Repository) *GetDrawByIDService {
	return &GetDrawByIDService{
		repository: repository,
	}
}

// GetDrawByID retrieves a draw by ID
func (s *GetDrawByIDService) GetDrawByID(ctx context.Context, input GetDrawByIDInput) (GetDrawByIDOutput, error) {
	// Implementation using domain types
	draw, winners, err := s.repository.GetDrawByID(ctx, input.ID)
	if err != nil {
		return GetDrawByIDOutput{}, err
	}
	
	// Convert to output format
	winnerOutputs := make([]drawDomain.Winner, len(winners))
	for i, winner := range winners {
		winnerOutputs[i] = *winner
	}
	
	return GetDrawByIDOutput{
		ID:                   draw.ID,
		DrawDate:             draw.DrawDate,
		PrizeStructureID:     draw.PrizeStructureID,
		Status:               draw.Status,
		TotalEligibleMSISDNs: draw.TotalEligibleMSISDNs,
		TotalEntries:         draw.TotalEntries,
		ExecutedBy:           draw.ExecutedByAdminID, // Map from domain field
		Winners:              winnerOutputs,
		CreatedAt:            draw.CreatedAt,
		UpdatedAt:            draw.UpdatedAt,
	}, nil
}
