package adapter

import (
	"context"

	"github.com/google/uuid"
)

// PrizeHandlerAdapter adapts the prize service adapter to match the handler's expected interface
type PrizeHandlerAdapter struct {
	prizeServiceAdapter *PrizeServiceAdapter
}

// NewPrizeHandlerAdapter creates a new PrizeHandlerAdapter
func NewPrizeHandlerAdapter(
	prizeServiceAdapter *PrizeServiceAdapter,
) *PrizeHandlerAdapter {
	return &PrizeHandlerAdapter{
		prizeServiceAdapter: prizeServiceAdapter,
	}
}

// DeletePrizeStructureService provides the missing service required by the handler
type DeletePrizeStructureService struct {
	prizeServiceAdapter *PrizeServiceAdapter
}

// NewDeletePrizeStructureService creates a new DeletePrizeStructureService
func NewDeletePrizeStructureService(
	prizeServiceAdapter *PrizeServiceAdapter,
) *DeletePrizeStructureService {
	return &DeletePrizeStructureService{
		prizeServiceAdapter: prizeServiceAdapter,
	}
}

// DeletePrizeStructure implements the service method required by the handler
func (s *DeletePrizeStructureService) DeletePrizeStructure(
	ctx context.Context,
	input struct {
		ID        uuid.UUID
		DeletedBy uuid.UUID
	},
) error {
	return s.prizeServiceAdapter.DeletePrizeStructure(ctx, input.ID, input.DeletedBy)
}
