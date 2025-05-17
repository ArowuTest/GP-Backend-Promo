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
	ID                 uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Username           string        `json:"username" gorm:"uniqueIndex;not null"`
	Email              string        `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash       string        `json:"-" gorm:"not null"`
	Salt               string        `json:"-" gorm:"not null"`
	FirstName          string        `json:"first_name,omitempty"`
	LastName           string        `json:"last_name,omitempty"`
	Role               AdminUserRole `json:"role" gorm:"type:admin_user_role;not null"`
	Status             UserStatus    `json:"status" gorm:"type:user_status;default:'Active'"`
	LastLoginAt        *time.Time    `json:"last_login_at,omitempty"`
	FailedLoginAttempts int           `json:"failed_login_attempts,omitempty" gorm:"default:0"`
}

// BeforeCreate will set a UUID for AdminUser
func (u *AdminUser) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

// CreatePrizeRequest defines the structure for a prize tier when creating/updating a prize structure.
// This was missing and caused build errors in prize_handler.go.
type CreatePrizeRequest struct {
	Name             string `json:"name" binding:"required"`
	Value            string `json:"value,omitempty"`
	PrizeType        string `json:"prize_type,omitempty"`
	Quantity         int    `json:"quantity" binding:"required,min=1"`
	Order            int    `json:"order,omitempty"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps,omitempty" binding:"min=0"`
}

// PrizeStructure represents the structure of prizes for a draw or period
type PrizeStructure struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name             string         `json:"name" gorm:"uniqueIndex;not null"`
	Description      string         `json:"description,omitempty"`
	IsActive         bool           `json:"is_active" gorm:"default:true"`
	ValidFrom        *time.Time     `json:"valid_from,omitempty"`
	ValidTo          *time.Time     `json:"valid_to,omitempty"`
	CreatedByAdminID uuid.UUID      `json:"created_by_admin_id,omitempty" gorm:"type:uuid"`
	DayType          string         `json:"day_type,omitempty" gorm:"column:day_type;not null"` // Added DayType field
	ApplicableDays   []string       `json:"applicable_days,omitempty" gorm:"-"`                 // Virtual field, not stored in DB
	Prizes           []Prize        `json:"prizes" gorm:"foreignKey:PrizeStructureID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	PrizeStructureID uuid.UUID      `json:"prize_structure_id" gorm:"type:uuid;not null"`
	Name             string         `json:"name" gorm:"not null"`
	Value            string         `json:"value"`            // Display value, e.g., "N1000 Airtime"
	PrizeType        string         `json:"prize_type,omitempty"` // e.g., Cash, Airtime, Data, Physical
	Quantity         int            `json:"quantity" gorm:"not null;default:1"`
	Order            int            `json:"order,omitempty"`
	NumberOfRunnerUps int           `json:"numberOfRunnerUps,omitempty" gorm:"default:1"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
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
	ID                     uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DrawDate               time.Time      `json:"draw_date" gorm:"not null"`
	PrizeStructureID       uuid.UUID      `json:"prize_structure_id" gorm:"type:uuid"`
	PrizeStructure         PrizeStructure `json:"prize_structure,omitempty"`
	ExecutedByUserID       uuid.UUID      `json:"executed_by_user_id" gorm:"type:uuid"`
	ExecutedByUser         AdminUser      `json:"executed_by_user,omitempty" gorm:"foreignKey:ExecutedByUserID"`
	Status                 string         `json:"status,omitempty"`
	EligibleParticipantsCount int         `json:"eligible_participants_count,omitempty"`
	TotalPointsInDraw      int            `json:"total_points_in_draw,omitempty"`
	Winners                []DrawWinner   `json:"winners,omitempty" gorm:"foreignKey:DrawID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	DrawID            uuid.UUID      `json:"draw_id" gorm:"type:uuid;not null"`
	PrizeID           uuid.UUID      `json:"prize_id" gorm:"type:uuid;not null"`
	Prize             Prize          `json:"prize,omitempty"`
	MSISDN            string         `json:"msisdn"`
	IsRunnerUp        bool           `json:"is_runner_up" gorm:"default:false"`
	RunnerUpRank      int            `json:"runner_up_rank,omitempty"`
	OriginalWinnerID  *uuid.UUID     `json:"original_winner_id,omitempty" gorm:"type:uuid"`
	PointsAtWin       int            `json:"points_at_win,omitempty"`
	NotificationStatus string         `json:"notification_status,omitempty"`
	ClaimStatus       string         `json:"claim_status" gorm:"default:'Pending'"`
	NotifiedAt        *time.Time     `json:"notified_at,omitempty"`
	ClaimedAt         *time.Time     `json:"claimed_at,omitempty"`
	ForfeitedAt       *time.Time     `json:"forfeited_at,omitempty"`
	Notes             string         `json:"notes,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
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
	OptInDate *time.Time     `json:"opt_in_date,omitempty"`
}

// BeforeCreate will set a UUID for Participant
func (pt *Participant) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return
}

// ParticipantEvent represents an individual point-earning event from a CSV upload.
type ParticipantEvent struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt           time.Time  `json:"created_at"`
	MSISDN              string     `json:"msisdn" gorm:"index;not null"`
	Amount              string     `json:"amount,omitempty"`
	RechargeAmount      float64    `json:"recharge_amount,omitempty"` // Added field to match draw_data_service.go
	OptInStatus         string     `json:"opt_in_status,omitempty"`
	PointsEarned        int        `json:"points_earned" gorm:"not null"`
	TransactionTimestamp *time.Time `json:"transaction_timestamp,omitempty" gorm:"index"`
	UploadAuditID       uuid.UUID  `json:"upload_audit_id" gorm:"type:uuid"`
	IsDuplicate         bool       `json:"is_duplicate" gorm:"default:false"`
	IsEligible          bool       `json:"is_eligible" gorm:"default:true"` // Added field for eligibility check
	Notes               string     `json:"notes,omitempty"`
}

// BeforeCreate will set a UUID for ParticipantEvent
func (pe *ParticipantEvent) BeforeCreate(tx *gorm.DB) (err error) {
	if pe.ID == uuid.Nil {
		pe.ID = uuid.New()
	}
	return
}

// ParticipantPoints represents the current point balance for a participant
type ParticipantPoints struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	MSISDN    string         `json:"msisdn" gorm:"uniqueIndex;not null"`
	Points    int            `json:"points" gorm:"not null;default:0"`
	LastEventID uuid.UUID    `json:"last_event_id,omitempty" gorm:"type:uuid"`
}

// BeforeCreate will set a UUID for ParticipantPoints
func (pp *ParticipantPoints) BeforeCreate(tx *gorm.DB) (err error) {
	if pp.ID == uuid.Nil {
		pp.ID = uuid.New()
	}
	return
}

// BlacklistedMSISDN represents a blacklisted phone number that should be excluded from draws
type BlacklistedMSISDN struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	MSISDN      string         `json:"msisdn" gorm:"uniqueIndex;not null"`
	Reason      string         `json:"reason,omitempty"`
	AddedByUserID uuid.UUID    `json:"added_by_user_id,omitempty" gorm:"type:uuid"`
}

// BeforeCreate will set a UUID for BlacklistedMSISDN
func (b *BlacklistedMSISDN) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

// DataUploadAudit represents an audit entry for data uploads
type DataUploadAudit struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UploadedByUserID    uuid.UUID `json:"uploaded_by_user_id" gorm:"type:uuid;not null"`
	FileName            string    `json:"file_name"`
	Status              string    `json:"status" gorm:"not null;default:'Pending'"`
	Notes               string    `json:"notes"`
	// Re-added fields that were missing and causing build errors
	RecordCount         int       `json:"record_count"`
	OperationType       string    `json:"operation_type"`
	UploadTimestamp     time.Time `json:"upload_timestamp"`
	SuccessfullyImported int      `json:"successfully_imported"`
	DuplicatesSkipped   int       `json:"duplicates_skipped"`
}

// BeforeCreate will set a UUID for DataUploadAudit
func (da *DataUploadAudit) BeforeCreate(tx *gorm.DB) (err error) {
	if da.ID == uuid.Nil {
		da.ID = uuid.New()
	}
	return
}

// AuditLog represents a general audit log entry
type AuditLog struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time      `json:"created_at"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Action    string         `json:"action" gorm:"not null"`
	EntityType string        `json:"entity_type" gorm:"not null"`
	EntityID  uuid.UUID      `json:"entity_id" gorm:"type:uuid;not null"`
	Details   string         `json:"details"`
	IPAddress string         `json:"ip_address"`
}

// BeforeCreate will set a UUID for AuditLog
func (al *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return
}
