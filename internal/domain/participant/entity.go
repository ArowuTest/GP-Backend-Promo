package participant

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Participant represents a participant entity in the domain
type Participant struct {
	ID             uuid.UUID
	MSISDN         string
	Points         int
	RechargeAmount float64
	RechargeDate   time.Time
	UploadID       uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ParticipantInput represents input for creating a participant
type ParticipantInput struct {
	MSISDN         string
	Points         int
	RechargeAmount float64
	RechargeDate   time.Time
}

// ParticipantRepository defines the interface for participant data access
type ParticipantRepository interface {
	Create(participant *Participant) error
	GetByMSISDN(msisdn string) (*Participant, error)
	GetByMSISDNAndDate(msisdn string, date time.Time) (*Participant, error)
	List(page, pageSize int) ([]Participant, int, error)
	ListByDate(date time.Time, page, pageSize int) ([]Participant, int, error)
	GetStatsByDate(date time.Time) (int, int, error)
	GetStats(date time.Time) (int, int, float64, error)
	BulkCreate(participants []*Participant) (int, []string, error)
	CreateBatch(participants []*Participant) (int, []string, error)
	DeleteByUploadID(uploadID uuid.UUID) error
}

// UploadAudit represents an audit record for participant data uploads
type UploadAudit struct {
	ID              uuid.UUID
	UploadedBy      uuid.UUID
	UploadDate      time.Time
	FileName        string
	Status          string // "Completed", "Failed", "Processing"
	TotalRows       int
	SuccessfulRows  int
	ErrorCount      int
	ErrorDetails    []string
	ErrorMessage    string       // Added for adapter layer compatibility
	ProcessingTime  string       // Added for adapter layer compatibility
	RecordCount     int          // Added for adapter layer compatibility
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UploadAuditRepository defines the interface for upload audit data access
type UploadAuditRepository interface {
	Create(audit *UploadAudit) error
	GetByID(id uuid.UUID) (*UploadAudit, error)
	List(page, pageSize int) ([]UploadAudit, int, error)
	Update(audit *UploadAudit) error
	Delete(id uuid.UUID) error
}

// ParticipantError represents domain-specific errors for the participant domain
type ParticipantError struct {
	Code    string
	Message string
	Err     error
}

// Error codes for the participant domain
const (
	ErrParticipantNotFound     = "PARTICIPANT_NOT_FOUND"
	ErrInvalidMSISDN           = "INVALID_MSISDN"
	ErrInvalidRechargeAmount   = "INVALID_RECHARGE_AMOUNT"
	ErrInvalidRechargeDate     = "INVALID_RECHARGE_DATE"
	ErrDuplicateParticipant    = "DUPLICATE_PARTICIPANT"
	ErrUploadAuditNotFound     = "UPLOAD_AUDIT_NOT_FOUND"
	ErrInvalidCSVFormat        = "INVALID_CSV_FORMAT"
)

// Error implements the error interface
func (e *ParticipantError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *ParticipantError) Unwrap() error {
	return e.Err
}

// NewParticipantError creates a new ParticipantError
func NewParticipantError(code, message string, err error) *ParticipantError {
	return &ParticipantError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidateMSISDN validates that an MSISDN is in the correct format
func ValidateMSISDN(msisdn string) error {
	if msisdn == "" {
		return errors.New("MSISDN cannot be empty")
	}
	
	// Additional validation logic can be added here
	// For example, ensuring the MSISDN follows the correct format for Nigerian phone numbers
	
	return nil
}

// CalculatePoints calculates the number of points based on recharge amount
func CalculatePoints(rechargeAmount float64) int {
	// Every full N100 recharge is 1 point
	return int(rechargeAmount / 100)
}
