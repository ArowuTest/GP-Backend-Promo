package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// GetDrawByIDInput represents input for GetDrawByID
type GetDrawByIDInput struct {
	ID string
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
	Winners              []draw.Winner
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Draw                 *draw.Draw // Keep the original field for backward compatibility
}

// GetDrawByIDService handles retrieving draw details by ID
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
func (s *GetDrawByIDService) GetDrawByID(ctx context.Context, input GetDrawByIDInput) (*GetDrawByIDOutput, error) {
	// Parse ID to UUID
	id, err := parseUUID(input.ID)
	if err != nil {
		return nil, err
	}
	
	// Get draw from repository
	drawEntity, err := s.repository.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	return &GetDrawByIDOutput{
		ID:                   drawEntity.ID,
		DrawDate:             drawEntity.DrawDate,
		PrizeStructureID:     drawEntity.PrizeStructureID,
		Status:               drawEntity.Status,
		TotalEligibleMSISDNs: drawEntity.TotalEligibleMSISDNs,
		TotalEntries:         drawEntity.TotalEntries,
		ExecutedBy:           drawEntity.ExecutedBy,
		Winners:              drawEntity.Winners,
		CreatedAt:            drawEntity.CreatedAt,
		UpdatedAt:            drawEntity.UpdatedAt,
		Draw:                 drawEntity,
	}, nil
}
