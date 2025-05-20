package participant

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// DeleteUploadService provides functionality for deleting participant uploads
type DeleteUploadService struct {
	participantRepository participantDomain.ParticipantRepository
}

// NewDeleteUploadService creates a new DeleteUploadService
func NewDeleteUploadService(participantRepository participantDomain.ParticipantRepository) *DeleteUploadService {
	return &DeleteUploadService{
		participantRepository: participantRepository,
	}
}

// DeleteUploadInput defines the input for the DeleteUpload use case
type DeleteUploadInput struct {
	UploadID  uuid.UUID
	DeletedBy uuid.UUID
}

// DeleteUploadOutput defines the output for the DeleteUpload use case
type DeleteUploadOutput struct {
	UploadID uuid.UUID
	Deleted  bool
}

// DeleteUpload deletes a participant upload and its associated participants
func (s *DeleteUploadService) DeleteUpload(ctx context.Context, input DeleteUploadInput) (DeleteUploadOutput, error) {
	// Delete upload and its participants from repository
	err := s.participantRepository.DeleteUpload(ctx, input.UploadID)
	if err != nil {
		return DeleteUploadOutput{}, fmt.Errorf("failed to delete upload: %w", err)
	}

	return DeleteUploadOutput{
		UploadID: input.UploadID,
		Deleted:  true,
	}, nil
}
