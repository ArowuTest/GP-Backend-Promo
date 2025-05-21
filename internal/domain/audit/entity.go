package audit

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entity in the domain
type AuditLog struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Username    string
	Action      string
	EntityType  string
	EntityID    string
	Description string
	IPAddress   string
	UserAgent   string
	Metadata    map[string]interface{}
	CreatedAt   time.Time
}

// SystemAuditLog represents a system-level audit log
type SystemAuditLog struct {
	ID          uuid.UUID
	Action      string
	Description string
	Severity    string // "Info", "Warning", "Error", "Critical"
	Source      string
	Metadata    map[string]interface{}
	CreatedAt   time.Time
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
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// AuditLogFilters defines filters for retrieving audit logs
type AuditLogFilters struct {
	StartDate  time.Time
	EndDate    time.Time
	UserID     uuid.UUID
	Action     string
	EntityType string
	Page       int
	PageSize   int
}

// AuditService defines the interface for audit logging
type AuditService interface {
	LogAudit(action, entityType string, entityID uuid.UUID, userID uuid.UUID, summary, details string) error
}

// AuditRepository defines the interface for audit log data access
type AuditRepository interface {
	Create(log *AuditLog) error
	GetByID(id uuid.UUID) (*AuditLog, error)
	List(filters AuditLogFilters, page, pageSize int) ([]AuditLog, int, error)
	
	CreateSystemAuditLog(log *SystemAuditLog) error
	GetSystemAuditLogByID(id uuid.UUID) (*SystemAuditLog, error)
	ListSystemAuditLogs(filters map[string]interface{}, page, pageSize int) ([]SystemAuditLog, int, error)
}

// AuditError represents domain-specific errors for the audit domain
type AuditError struct {
	Code    string
	Message string
	Err     error
}

// Error codes for the audit domain
const (
	ErrAuditLogNotFound      = "AUDIT_LOG_NOT_FOUND"
	ErrInvalidAuditLog       = "INVALID_AUDIT_LOG"
	ErrSystemAuditLogNotFound = "SYSTEM_AUDIT_LOG_NOT_FOUND"
	ErrInvalidSystemAuditLog  = "INVALID_SYSTEM_AUDIT_LOG"
)

// Error implements the error interface
func (e *AuditError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AuditError) Unwrap() error {
	return e.Err
}

// NewAuditError creates a new AuditError
func NewAuditError(code, message string, err error) *AuditError {
	return &AuditError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidateAuditLog validates that an audit log is valid
func ValidateAuditLog(log *AuditLog) error {
	if log.Action == "" {
		return errors.New("audit log action cannot be empty")
	}
	
	if log.EntityType == "" {
		return errors.New("audit log entity type cannot be empty")
	}
	
	return nil
}

// ValidateSystemAuditLog validates that a system audit log is valid
func ValidateSystemAuditLog(log *SystemAuditLog) error {
	if log.Action == "" {
		return errors.New("system audit log action cannot be empty")
	}
	
	if log.Description == "" {
		return errors.New("system audit log description cannot be empty")
	}
	
	if log.Severity == "" {
		return errors.New("system audit log severity cannot be empty")
	}
	
	return nil
}
