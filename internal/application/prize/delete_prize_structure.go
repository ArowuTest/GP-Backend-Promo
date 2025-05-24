package prize

import (
	"context"
	
	prizeDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

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
func (s *DeletePrizeStructureService) DeletePrizeStructure(ctx context.Context, input prizeDomain.DeletePrizeStructureInput) error {
	// Verify the prize structure exists
	if _, err := s.prizeRepository.GetPrizeStructureByID(input.ID); err != nil {
		return err
	}
	
	// Delete the prize structure
	return s.prizeRepository.DeletePrizeStructure(input.ID, input.DeletedBy)
}
