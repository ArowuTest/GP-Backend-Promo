package participant

import (
	"context"
	"errors"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// UploadParticipantsService provides functionality for uploading participants
type UploadParticipantsService struct {
	participantRepository participant.ParticipantRepository
	auditService          audit.AuditService
}

// NewUploadParticipantsService creates a new UploadParticipantsService
func NewUploadParticipantsService(
	participantRepository participant.ParticipantRepository,
	auditService audit.AuditService,
) *UploadParticipantsService {
	return &UploadParticipantsService{
		participantRepository: participantRepository,
		auditService:          auditService,
	}
}

// UploadParticipantsInput defines the input for the UploadParticipants use case
type UploadParticipantsInput struct {
	Participants []ParticipantInput `json:"participants"`
	UploadedBy   uuid.UUID          `json:"uploadedBy"`
}

// ParticipantInput defines the input for a participant
type ParticipantInput struct {
	MSISDN      string    `json:"msisdn"`
	RechargeAmount float64   `json:"rechargeAmount"`
	RechargeDate string    `json:"rechargeDate"`
}

// UploadParticipantsOutput defines the output for the UploadParticipants use case
type UploadParticipantsOutput struct {
	TotalUploaded int       `json:"totalUploaded"`
	UploadID      uuid.UUID `json:"uploadId"`
	UploadedAt    time.Time `json:"uploadedAt"`
}

// UploadParticipants uploads a batch of participants
func (s *UploadParticipantsService) UploadParticipants(ctx context.Context, input UploadParticipantsInput) (*UploadParticipantsOutput, error) {
	if len(input.Participants) == 0 {
		return nil, errors.New("at least one participant is required")
	}
	
	if input.UploadedBy == uuid.Nil {
		return nil, errors.New("uploaded by is required")
	}
	
	// Create upload record
	uploadID := uuid.New()
	now := time.Now()
	
	// Process participants
	participants := make([]*participant.Participant, 0, len(input.Participants))
	for _, p := range input.Participants {
		// Parse recharge date
		rechargeDate, err := time.Parse("2006-01-02", p.RechargeDate)
		if err != nil {
			return nil, fmt.Errorf("invalid recharge date for MSISDN %s: %w", p.MSISDN, err)
		}
		
		// Calculate points (1 point per N100)
		points := int(p.RechargeAmount / 100)
		
		participant := &participant.Participant{
			ID:             uuid.New(),
			MSISDN:         p.MSISDN,
			RechargeAmount: p.RechargeAmount,
			RechargeDate:   rechargeDate,
			Points:         points,
			UploadID:       uploadID,
			CreatedAt:      now,
		}
		
		participants = append(participants, participant)
	}
	
	// Save participants
	successCount, _, err := s.participantRepository.CreateBatch(participants)
	if err != nil {
		return nil, fmt.Errorf("failed to create participants: %w", err)
	}
	
	// Log audit
	if err := s.auditService.LogAudit(
		"UPLOAD_PARTICIPANTS",
		"Participant",
		uploadID,
		input.UploadedBy,
		fmt.Sprintf("Participants uploaded: %d", len(participants)),
		"",
	); err != nil {
		// Log error but continue
		fmt.Printf("Failed to log audit: %v\n", err)
	}
	
	return &UploadParticipantsOutput{
		TotalUploaded: successCount,
		UploadID:      uploadID,
		UploadedAt:    now,
	}, nil
}
