package participant

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	participantDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// Repository defines the interface for participant repository
type Repository interface {
	GetParticipantByMSISDN(ctx context.Context, msisdn string) (*participantDomain.Participant, error)
	ListParticipants(ctx context.Context, page, pageSize int) ([]*participantDomain.Participant, int, error)
	GetParticipantStats(ctx context.Context, date time.Time) (int, int, float64, error)
	UploadParticipants(ctx context.Context, participants []*participantDomain.ParticipantInput, uploadedBy uuid.UUID, fileName string) (*participantDomain.UploadAudit, error)
	ListUploadAudits(ctx context.Context, page, pageSize int) ([]*participantDomain.UploadAudit, int, error)
	DeleteUpload(ctx context.Context, uploadID, deletedBy uuid.UUID) error
}
