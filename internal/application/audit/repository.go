package audit

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	auditDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// Repository defines the interface for audit repository
type Repository interface {
	CreateAuditLog(ctx context.Context, log *auditDomain.AuditLog) error
	GetAuditLogByID(ctx context.Context, id uuid.UUID) (*auditDomain.AuditLog, error)
	ListAuditLogs(ctx context.Context, filters auditDomain.AuditLogFilters, page, pageSize int) ([]*auditDomain.AuditLog, int, error)
	GetDataUploadAudits(ctx context.Context, page, pageSize int, startDate, endDate time.Time) ([]*auditDomain.DataUploadAudit, int, error)
}
