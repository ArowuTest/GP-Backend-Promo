package prize

import (
	"context"
	"github.com/google/uuid"
	
	prizeDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// DeletePrizeStructureInput defines the input for deleting a prize structure
type DeletePrizeStructureInput struct {
	ID        uuid.UUID
	DeletedBy uuid.UUID
}

// DeletePrizeStructureService defines the service for deleting prize structures
type DeletePrizeStructureService struct {
	prizeRepository prizeDomain.PrizeRepository
}

// NewDeletePrizeStructureService creates a new DeletePrizeStructureService
func NewDeletePrizeStructureService(prizeRepository prizeDomain.PrizeRepository) *DeletePrizeStructureService {
	return &DeletePrizeStructureService{
		prizeRepository: prizeRepository,
	}
}

// DeletePrizeStructure deletes a prize structure
func (s *DeletePrizeStructureService) DeletePrizeStructure(ctx context.Context, input DeletePrizeStructureInput) error {
	// Convert application input to domain input
	domainInput := prizeDomain.DeletePrizeStructureInput{
		ID:        input.ID,
		DeletedBy: input.DeletedBy,
	}
	
	// Verify the prize structure exists
	if _, err := s.prizeRepository.GetPrizeStructureByID(domainInput.ID); err != nil {
		return err
	}
	
	// Delete the prize structure
	return s.prizeRepository.DeletePrizeStructure(domainInput.ID, domainInput.DeletedBy)
}
