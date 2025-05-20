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
	uploadAuditRepository participantDomain.UploadAuditRepository
}

// NewDeleteUploadService creates a new DeleteUploadService
func NewDeleteUploadService(
	participantRepository participantDomain.ParticipantRepository,
	uploadAuditRepository participantDomain.UploadAuditRepository,
) *DeleteUploadService {
	return &DeleteUploadService{
		participantRepository: participantRepository,
		uploadAuditRepository: uploadAuditRepository,
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
	// First delete all participants associated with this upload
	err := s.participantRepository.DeleteByUploadID(input.UploadID)
	if err != nil {
		return DeleteUploadOutput{}, fmt.Errorf("failed to delete participants: %w", err)
	}

	// Then delete the upload audit record
	err = s.uploadAuditRepository.Delete(input.UploadID)
	if err != nil {
		return DeleteUploadOutput{}, fmt.Errorf("failed to delete upload audit: %w", err)
	}

	return DeleteUploadOutput{
		UploadID: input.UploadID,
		Deleted:  true,
	}, nil
}
