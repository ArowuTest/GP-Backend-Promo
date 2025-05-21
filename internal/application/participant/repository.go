package participant

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// Repository defines the interface for participant repository
type Repository interface {
	ListParticipants(ctx context.Context, page, pageSize int) ([]Participant, int, error)
	ListUploadAudits(ctx context.Context, page, pageSize int) ([]UploadAudit, int, error)
	DeleteUpload(ctx context.Context, uploadID, deletedBy uuid.UUID) (bool, error)
	UploadParticipants(ctx context.Context, participants []ParticipantInput, uploadedBy uuid.UUID, fileName string) (uuid.UUID, int, int, int, error)
	GetParticipantStats(ctx context.Context, startDate, endDate string) (int, int, error)
}

// Participant represents a participant record
type Participant struct {
	ID             uuid.UUID
	MSISDN         string
	Points         int
	RechargeAmount float64
	RechargeDate   time.Time
	CreatedAt      time.Time
	UploadID       uuid.UUID
}

// UploadAudit represents an upload audit record
type UploadAudit struct {
	ID             uuid.UUID
	UploadedBy     uuid.UUID
	UploadDate     time.Time
	FileName       string
	Status         string
	TotalRows      int
	SuccessfulRows int
	ErrorCount     int
	ErrorDetails   string
}

// ParticipantInput represents input for uploading a participant
type ParticipantInput struct {
	MSISDN         string
	RechargeAmount float64
	RechargeDate   string
}
