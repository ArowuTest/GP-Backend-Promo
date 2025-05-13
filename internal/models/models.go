package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminUserRole defines the possible roles for an admin user
type AdminUserRole string

const (
	RoleSuperAdmin        AdminUserRole = "SUPER_ADMIN"
	RoleAdmin             AdminUserRole = "ADMIN"
	RoleSeniorUser        AdminUserRole = "SENIOR_USER"
	RoleWinnerReportsUser AdminUserRole = "WINNER_REPORTS_USER"
	RoleAllReportUser     AdminUserRole = "ALL_REPORT_USER"
)

// UserStatus defines the possible statuses for an admin user
type UserStatus string

const (
	StatusActive   UserStatus = "Active"
	StatusInactive UserStatus = "Inactive"
	StatusLocked   UserStatus = "Locked"
)

// AdminUser represents the structure for an admin user
type AdminUser struct {
	ID                  uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Username            string         `json:"username" gorm:"uniqueIndex;not null"`
	Email               string         `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash        string         `json:"-" gorm:"not null"`
	Salt                string         `json:"-" gorm:"not null"`
	FirstName           string         `json:"first_name,omitempty"`
	LastName            string         `json:"last_name,omitempty"`
	Role                AdminUserRole  `json:"role" gorm:"type:admin_user_role;not null"`
	Status              UserStatus     `json:"status" gorm:"type:user_status;default:'Active'"`
	LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
	FailedLoginAttempts int            `json:"failed_login_attempts,omitempty" gorm:"default:0"`
}

// BeforeCreate will set a UUID for AdminUser
func (u *AdminUser) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

// CreatePrizeRequest defines the structure for creating a prize tier within a prize structure request
type CreatePrizeRequest struct {
	Name      string `json:"name" binding:"required"`
	Value     string `json:"value,omitempty"`
	PrizeType string `json:"prize_type,omitempty"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Order     int    `json:"order,omitempty"`
}

// PrizeStructure represents the structure of prizes for a draw or period
type PrizeStructure struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name              string         `json:"name" gorm:"uniqueIndex;not null"`
	Description       string         `json:"description,omitempty"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	ValidFrom         *time.Time     `json:"valid_from,omitempty"`
	ValidTo           *time.Time     `json:"valid_to,omitempty"`
	CreatedByAdminID  uuid.UUID      `json:"created_by_admin_id,omitempty" gorm:"type:uuid"`
	Prizes            []Prize        `json:"prizes" gorm:"foreignKey:PrizeStructureID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// BeforeCreate will set a UUID for PrizeStructure
func (ps *PrizeStructure) BeforeCreate(tx *gorm.DB) (err error) {
	if ps.ID == uuid.Nil {
		ps.ID = uuid.New()
	}
	return
}

// Prize represents a single prize within a prize structure
type Prize struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	PrizeStructureID  uuid.UUID      `json:"prize_structure_id" gorm:"type:uuid;not null"`
	Name              string         `json:"name" gorm:"not null"`
	Value             string         `json:"value"`
	PrizeType         string         `json:"prize_type,omitempty"`
	Quantity          int            `json:"quantity" gorm:"not null;default:1"`
	Order             int            `json:"order,omitempty"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate will set a UUID for Prize
func (p *Prize) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

// Draw represents a draw event
type Draw struct {
	ID                        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt                 time.Time      `json:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at"`
	DeletedAt                 gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DrawDate                  time.Time      `json:"draw_date" gorm:"not null"`
	PrizeStructureID          uuid.UUID      `json:"prize_structure_id" gorm:"type:uuid"`
	ExecutedByUserID          uuid.UUID      `json:"executed_by_user_id" gorm:"type:uuid"`
	Status                    string         `json:"status,omitempty"`
	EligibleParticipantsCount int            `json:"eligible_participants_count,omitempty"`
	TotalPointsInDraw         int            `json:"total_points_in_draw,omitempty"`
	Winners                   []DrawWinner   `json:"winners,omitempty" gorm:"foreignKey:DrawID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// BeforeCreate will set a UUID for Draw
func (d *Draw) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return
}

// DrawWinner represents a winner for a specific prize in a draw
type DrawWinner struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	DrawID        uuid.UUID      `json:"draw_id" gorm:"type:uuid;not null"`
	PrizeID       uuid.UUID      `json:"prize_id" gorm:"type:uuid;not null"`
	MSISDN        string         `json:"msisdn"`
	IsRunnerUp    bool           `json:"is_runner_up" gorm:"default:false"`
	RunnerUpRank  int            `json:"runner_up_rank,omitempty"`
	PointsAtWin   int            `json:"points_at_win,omitempty"`
	NotificationStatus string    `json:"notification_status,omitempty"`
	ClaimStatus   string         `json:"claim_status" gorm:"default:'Pending'"`
	NotifiedAt    *time.Time     `json:"notified_at,omitempty"`
	ClaimedAt     *time.Time     `json:"claimed_at,omitempty"`
	ForfeitedAt   *time.Time     `json:"forfeited_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate will set a UUID for DrawWinner
func (dw *DrawWinner) BeforeCreate(tx *gorm.DB) (err error) {
	if dw.ID == uuid.Nil {
		dw.ID = uuid.New()
	}
	return
}

// Participant represents an individual who can participate in draws (Master Record)
type Participant struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	MSISDN    string         `json:"msisdn" gorm:"uniqueIndex;not null"`
	// Points field is deprecated here; points are aggregated from ParticipantEvent records.
	// Points    int            `json:"points" gorm:"default:0"` 
	OptInDate *time.Time     `json:"opt_in_date,omitempty"` // Date the participant first opted in or became eligible
}

// BeforeCreate will set a UUID for Participant
func (pt *Participant) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return
}

// ParticipantEvent represents an individual point-earning event from a CSV upload.
// This table will store each row from the CSV as a separate event.
type ParticipantEvent struct {
	ID                   uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt            time.Time      `json:"created_at"` // Timestamp of when this event record was created in DB
	MSISDN               string         `json:"msisdn" gorm:"index;not null"`
	Amount               string         `json:"amount,omitempty"`      // Amount from CSV (can be string if it varies)
	OptInStatus          string         `json:"opt_in_status,omitempty"` // OptInStatus from CSV
	PointsEarned         int            `json:"points_earned" gorm:"not null"`
	TransactionTimestamp *time.Time     `json:"transaction_timestamp,omitempty" gorm:"index"` // Timestamp from CSV, represents when the event occurred
	UploadAuditID        uuid.UUID      `json:"upload_audit_id" gorm:"type:uuid"` // FK to DataUploadAudit
	IsDuplicate          bool           `json:"is_duplicate" gorm:"default:false"` // Flag to mark if this was identified as a true duplicate during upload but loaded due to override (future use)
	Notes                string         `json:"notes,omitempty"` // For any notes specific to this event processing
}

// BeforeCreate will set a UUID for ParticipantEvent
func (pe *ParticipantEvent) BeforeCreate(tx *gorm.DB) (err error) {
	if pe.ID == uuid.Nil {
		pe.ID = uuid.New()
	}
	return
}

// DataUploadAudit represents an audit trail for data uploads
type DataUploadAudit struct {
	ID                   uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	UploadedByUserID     uuid.UUID      `json:"uploaded_by_user_id" gorm:"type:uuid"`
	UploadTimestamp      time.Time      `json:"upload_timestamp" gorm:"not null"`
	FileName             string         `json:"file_name,omitempty"`
	RecordCount          int            `json:"record_count,omitempty"` // Total data rows in CSV
	SuccessfullyImported int            `json:"successfully_imported,omitempty"` // Count of unique events inserted
	DuplicatesSkipped    int            `json:"duplicates_skipped,omitempty"`    // Count of true duplicates skipped
	Status               string         `json:"status,omitempty"` // e.g., "Success", "Failed", "Partial Success"
	Notes                string         `json:"notes,omitempty"`      // For error messages or other details, including list of skipped duplicates
	OperationType        string         `json:"operation_type,omitempty"` // e.g., "ParticipantUpload", "BlacklistUpload"
	CreatedAt            time.Time      `json:"created_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate will set a UUID for DataUploadAudit
func (dua *DataUploadAudit) BeforeCreate(tx *gorm.DB) (err error) {
	if dua.ID == uuid.Nil {
		dua.ID = uuid.New()
	}
	return
}

