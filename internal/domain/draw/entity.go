package draw

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Draw represents a draw entity in the domain
type Draw struct {
	ID                  uuid.UUID
	DrawDate            time.Time
	PrizeStructureID    uuid.UUID
	Status              string // "Pending", "Completed", "Failed"
	TotalEligibleMSISDNs int
	TotalEntries        int
	ExecutedByAdminID   uuid.UUID
	Winners             []Winner
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Winner represents a winner entity in the domain
type Winner struct {
	ID            uuid.UUID
	DrawID        uuid.UUID
	MSISDN        string
	PrizeTierID   uuid.UUID
	Status        string // "PendingNotification", "Notified", "Confirmed"
	PaymentStatus string // "Pending", "Paid", "Failed"
	PaymentNotes  string
	PaidAt        *time.Time
	IsRunnerUp    bool
	RunnerUpRank  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DrawRepository defines the interface for draw data access
type DrawRepository interface {
	Create(draw *Draw) error
	GetByID(id uuid.UUID) (*Draw, error)
	List(page, pageSize int) ([]Draw, int, error)
	GetByDate(date time.Time) (*Draw, error)
	Update(draw *Draw) error
	GetEligibilityStats(date time.Time) (int, int, error)
	CreateWinner(winner *Winner) error
	GetWinnerByID(id uuid.UUID) (*Winner, error)
	UpdateWinner(winner *Winner) error
	GetRunnerUps(drawID uuid.UUID, prizeTierID uuid.UUID, limit int) ([]Winner, error)
}

// DrawError represents domain-specific errors for the draw domain
type DrawError struct {
	Code    string
	Message string
	Err     error
}

// Error codes for the draw domain
const (
	ErrDrawNotFound          = "DRAW_NOT_FOUND"
	ErrDrawAlreadyExists     = "DRAW_ALREADY_EXISTS"
	ErrInvalidDrawDate       = "INVALID_DRAW_DATE"
	ErrNoEligibleParticipants = "NO_ELIGIBLE_PARTICIPANTS"
	ErrWinnerNotFound        = "WINNER_NOT_FOUND"
	ErrNoRunnerUpsAvailable  = "NO_RUNNER_UPS_AVAILABLE"
)

// Error implements the error interface
func (e *DrawError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *DrawError) Unwrap() error {
	return e.Err
}

// NewDrawError creates a new DrawError
func NewDrawError(code, message string, err error) *DrawError {
	return &DrawError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidateDrawDate validates that a draw date is valid
func ValidateDrawDate(date time.Time) error {
	if date.IsZero() {
		return errors.New("draw date cannot be empty")
	}
	
	// Additional validation logic can be added here
	// For example, ensuring the date is not in the past
	
	return nil
}
