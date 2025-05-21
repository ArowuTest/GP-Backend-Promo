package audit

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// Repository defines the interface for audit repository
type Repository interface {
	GetAuditLogs(ctx context.Context, page, pageSize int, filters AuditLogFilters) ([]AuditLog, int, error)
	GetDataUploadAudits(ctx context.Context, page, pageSize int, startDate, endDate time.Time) ([]DataUploadAudit, int, error)
}

// DataUploadAudit represents a data upload audit record
type DataUploadAudit struct {
	ID                  uuid.UUID
	UploadedBy          uuid.UUID
	UploadedAt          time.Time
	FileName            string
	TotalUploaded       int
	SuccessfullyImported int
	DuplicatesSkipped   int
	ErrorsEncountered   int
	Status              string
	Details             string
	OperationType       string
}
