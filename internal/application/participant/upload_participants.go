package application

import (
	"time"
	"github.com/google/uuid"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// UploadParticipantsUseCase represents the use case for uploading participant data
type UploadParticipantsUseCase struct {
	participantRepository participant.ParticipantRepository
	uploadAuditRepository participant.UploadAuditRepository
}

// NewUploadParticipantsUseCase creates a new UploadParticipantsUseCase
func NewUploadParticipantsUseCase(
	participantRepository participant.ParticipantRepository,
	uploadAuditRepository participant.UploadAuditRepository,
) *UploadParticipantsUseCase {
	return &UploadParticipantsUseCase{
		participantRepository:  participantRepository,
		uploadAuditRepository: uploadAuditRepository,
	}
}

// UploadParticipantsInput represents the input for the upload participants use case
type UploadParticipantsInput struct {
	FileName    string
	UploadedBy  uuid.UUID
	Participants []ParticipantData
}

// ParticipantData represents a single participant record from the CSV
type ParticipantData struct {
	MSISDN         string
	RechargeAmount float64
	RechargeDate   time.Time
}

// UploadParticipantsOutput represents the output of the upload participants use case
type UploadParticipantsOutput struct {
	AuditID            uuid.UUID
	Status             string
	TotalRowsProcessed int
	SuccessfulRows     int
	ErrorCount         int
	ErrorDetails       []string
	DuplicatesSkipped  int
}

// Execute processes the participant data upload
func (uc *UploadParticipantsUseCase) Execute(input UploadParticipantsInput) (*UploadParticipantsOutput, error) {
	// Create audit record
	audit := &participant.UploadAudit{
		ID:              uuid.New(),
		UploadedBy:      input.UploadedBy,
		UploadDate:      time.Now(),
		FileName:        input.FileName,
		Status:          "Processing",
		TotalRows:       len(input.Participants),
		SuccessfulRows:  0,
		ErrorCount:      0,
		ErrorDetails:    []string{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save initial audit record
	if err := uc.uploadAuditRepository.Create(audit); err != nil {
		return nil, participant.NewParticipantError("AUDIT_CREATION_FAILED", "Failed to create upload audit record", err)
	}

	// Process participants
	participantsToCreate := make([]*participant.Participant, 0)
	errorDetails := make([]string, 0)
	duplicatesSkipped := 0

	for _, data := range input.Participants {
		// Validate MSISDN
		if err := participant.ValidateMSISDN(data.MSISDN); err != nil {
			errorDetails = append(errorDetails, "Invalid MSISDN: "+data.MSISDN+": "+err.Error())
			continue
		}

		// Calculate points
		points := participant.CalculatePoints(data.RechargeAmount)

		// Create participant entity
		newParticipant := &participant.Participant{
			ID:             uuid.New(),
			MSISDN:         data.MSISDN,
			Points:         points,
			RechargeAmount: data.RechargeAmount,
			RechargeDate:   data.RechargeDate,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		participantsToCreate = append(participantsToCreate, newParticipant)
	}

	// Bulk create participants
	successfulRows, errors, err := uc.participantRepository.BulkCreate(participantsToCreate)
	if err != nil {
		audit.Status = "Failed"
		audit.ErrorCount = len(input.Participants)
		audit.ErrorDetails = append(audit.ErrorDetails, "Bulk creation failed: "+err.Error())
		uc.uploadAuditRepository.Update(audit)
		return nil, participant.NewParticipantError("BULK_CREATE_FAILED", "Failed to create participants", err)
	}

	// Update audit record
	audit.Status = "Completed"
	audit.SuccessfulRows = successfulRows
	audit.ErrorCount = len(errors)
	audit.ErrorDetails = append(audit.ErrorDetails, errors...)
	uc.uploadAuditRepository.Update(audit)

	// Prepare output
	output := &UploadParticipantsOutput{
		AuditID:            audit.ID,
		Status:             audit.Status,
		TotalRowsProcessed: len(input.Participants),
		SuccessfulRows:     successfulRows,
		ErrorCount:         len(errors),
		ErrorDetails:       errors,
		DuplicatesSkipped:  duplicatesSkipped,
	}

	return output, nil
}
