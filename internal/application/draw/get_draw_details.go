package draw

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	
	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// GetDrawDetailsService provides functionality for retrieving draw details
type GetDrawDetailsService struct {
	drawRepository drawDomain.DrawRepository
}

// NewGetDrawDetailsService creates a new GetDrawDetailsService
func NewGetDrawDetailsService(drawRepository drawDomain.DrawRepository) *GetDrawDetailsService {
	return &GetDrawDetailsService{
		drawRepository: drawRepository,
	}
}

// GetDrawDetailsInput defines the input for the GetDrawDetails use case
type GetDrawDetailsInput struct {
	DrawID uuid.UUID
}

// GetDrawDetailsOutput defines the output for the GetDrawDetails use case
type GetDrawDetailsOutput struct {
	Draw    drawDomain.Draw
	Winners []drawDomain.Winner
}

// GetDrawDetails retrieves details for a specific draw
func (s *GetDrawDetailsService) GetDrawDetails(ctx context.Context, input GetDrawDetailsInput) (*GetDrawDetailsOutput, error) {
	if input.DrawID == uuid.Nil {
		return nil, fmt.Errorf("draw ID is required")
	}

	draw, err := s.drawRepository.GetByID(input.DrawID)
	if err != nil {
		return nil, fmt.Errorf("failed to get draw: %w", err)
	}

	return &GetDrawDetailsOutput{
		Draw:    *draw,
		Winners: draw.Winners,
	}, nil
}
