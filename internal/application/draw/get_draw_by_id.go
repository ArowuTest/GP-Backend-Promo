package draw

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// GetDrawByIDService provides functionality for retrieving a draw by ID
type GetDrawByIDService struct {
	drawRepository drawDomain.DrawRepository
}

// NewGetDrawByIDService creates a new GetDrawByIDService
func NewGetDrawByIDService(drawRepository drawDomain.DrawRepository) *GetDrawByIDService {
	return &GetDrawByIDService{
		drawRepository: drawRepository,
	}
}

// GetDrawByIDInput defines the input for the GetDrawByID use case
type GetDrawByIDInput struct {
	ID uuid.UUID
}

// GetDrawByIDOutput defines the output for the GetDrawByID use case
type GetDrawByIDOutput struct {
	ID                  uuid.UUID
	DrawDate            time.Time
	PrizeStructureID    uuid.UUID
	Status              string
	TotalEligibleMSISDNs int
	TotalEntries        int
	ExecutedBy          uuid.UUID
	Winners             []drawDomain.Winner
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// GetDrawByID retrieves a draw by ID
func (s *GetDrawByIDService) GetDrawByID(ctx context.Context, input GetDrawByIDInput) (GetDrawByIDOutput, error) {
	// Get draw from repository
	draw, err := s.drawRepository.GetByID(input.ID)
	if err != nil {
		return GetDrawByIDOutput{}, fmt.Errorf("failed to get draw: %w", err)
	}

	return GetDrawByIDOutput{
		ID:                  draw.ID,
		DrawDate:            draw.DrawDate,
		PrizeStructureID:    draw.PrizeStructureID,
		Status:              draw.Status,
		TotalEligibleMSISDNs: draw.TotalEligibleMSISDNs,
		TotalEntries:        draw.TotalEntries,
		ExecutedBy:          draw.ExecutedByAdminID,
		Winners:             draw.Winners,
		CreatedAt:           draw.CreatedAt,
		UpdatedAt:           draw.UpdatedAt,
	}, nil
}
