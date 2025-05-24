package entity

import (
	"time"

	"github.com/google/uuid"
)

// Base entity fields
type BaseEntity struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// User represents a user entity in the system
// Note: This is defined as a pointer type in the adapter layer
type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
	Role      string
	IsActive  bool
	LastLoginAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// AuthResult represents the result of an authentication attempt
type AuthResult struct {
	User      *User // Changed to pointer to match adapter usage
	Token     string
	ExpiresAt time.Time
}

// PaginatedUsers represents a paginated list of users
type PaginatedUsers struct {
	Users      []User
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// Participant represents a participant in the promotion
type Participant struct {
	ID           uuid.UUID
	MSISDN       string
	Points       int
	DateAdded    time.Time
	UploadAuditID uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    uuid.UUID
	UpdatedBy    uuid.UUID
}

// UploadAudit represents an audit record for participant uploads
type UploadAudit struct {
	ID            uuid.UUID
	FileName      string
	RecordsCount  int
	RecordCount   int
	UploadedBy    uuid.UUID
	UploadedByName string
	Status        string
	UploadDate    time.Time
	TotalRows     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     uuid.UUID
	UpdatedBy     uuid.UUID
}

// ParticipantStats represents statistics about participants
type ParticipantStats struct {
	TotalParticipants int
	TotalPoints       int
	LastUploadDate    time.Time
	LastUpdated       time.Time
}

// PaginatedParticipants represents a paginated list of participants
type PaginatedParticipants struct {
	Participants []Participant
	TotalCount   int
	Page         int
	PageSize     int
	TotalPages   int
}

// PaginatedUploadAudits represents a paginated list of upload audits
type PaginatedUploadAudits struct {
	UploadAudits []UploadAudit
	TotalCount   int
	Page         int
	PageSize     int
	TotalPages   int
}

// Prize represents a prize tier within a prize structure
type Prize struct {
	ID                uuid.UUID
	PrizeStructureID  uuid.UUID
	Name              string
	Description       string
	Value             float64
	Quantity          int
	Position          int
	NumberOfRunnerUps int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         uuid.UUID
	UpdatedBy         uuid.UUID
}

// PrizeInput represents input for creating or updating a prize structure
type PrizeInput struct {
	ID                uuid.UUID
	Name              string
	Description       string
	StartDate         time.Time
	EndDate           time.Time
	Value             float64
	Quantity          int
	IsActive          bool
	NumberOfRunnerUps int
}

// PrizeStructure represents a prize structure in the system
type PrizeStructure struct {
	ID                uuid.UUID
	Name              string
	Description       string
	StartDate         time.Time
	EndDate           time.Time
	Value             float64
	IsActive          bool
	NumberOfRunnerUps int
	Prizes            []Prize
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         uuid.UUID
	UpdatedBy         uuid.UUID
}

// PaginatedPrizeStructures represents a paginated list of prize structures
type PaginatedPrizeStructures struct {
	PrizeStructures []PrizeStructure
	TotalCount      int
	Page            int
	PageSize        int
	TotalPages      int
}

// AuditLog represents an audit log entry in the system
type AuditLog struct {
	ID          uuid.UUID
	Action      string
	EntityType  string
	EntityID    string
	Description string
	Details     string
	Metadata    map[string]interface{}
	UserID      uuid.UUID
	Username    string
	PerformedBy uuid.UUID
	PerformedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PaginatedAuditLogs represents a paginated list of audit logs
type PaginatedAuditLogs struct {
	AuditLogs  []AuditLog
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// Draw represents a draw entity
type Draw struct {
	ID                  uuid.UUID
	Name                string
	Description         string
	DrawDate            time.Time
	PrizeStructureID    uuid.UUID
	Status              string
	TotalEligibleMSISDNs int
	TotalEntries        int
	ExecutedByAdminID   uuid.UUID
	ExecutedBy          uuid.UUID
	RunnerUpsCount      int
	CreatedBy           uuid.UUID
	Winners             []Winner
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// DrawWithWinners represents a draw with its winners
type DrawWithWinners struct {
	Draw    Draw
	Winners []Winner
}

// PaginatedDraws represents a paginated list of draws
type PaginatedDraws struct {
	Draws      []Draw
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// Winner represents a winner in a draw
type Winner struct {
	ID            uuid.UUID
	DrawID        uuid.UUID
	MSISDN        string
	MaskedMSISDN  string
	PrizeID       uuid.UUID
	PrizeTierID   uuid.UUID
	PrizeTierName string
	PrizeName     string
	PrizeValue    float64
	Status        string
	PaymentStatus string
	PaymentNotes  string
	PaidAt        *time.Time
	IsRunnerUp    bool
	RunnerUpRank  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PaginatedWinners represents a paginated list of winners
type PaginatedWinners struct {
	Winners    []Winner
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// EligibilityStats represents statistics about eligible participants for a draw
type EligibilityStats struct {
	TotalEligible int
	TotalPoints   int
	TotalEligibleMSISDNs int
	TotalEntries  int
	DrawDate      time.Time
	LastUpdated   time.Time
}

// RunnerUpInvocationResult represents the result of invoking a runner-up
type RunnerUpInvocationResult struct {
	Success        bool
	Message        string
	OriginalWinner Winner
	NewWinner      Winner
}
