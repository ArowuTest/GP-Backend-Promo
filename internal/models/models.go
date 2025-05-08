package models

import (
	"time"

	"github.com/google/uuid"
)

// AdminUserRole defines the roles an admin user can have
type AdminUserRole string

const (
	SuperAdminRole    AdminUserRole = "SUPER_ADMIN"
	DrawAdminRole     AdminUserRole = "DRAW_ADMIN"
	ViewOnlyAdminRole AdminUserRole = "VIEW_ONLY_ADMIN"
)

// UserStatus defines the status of a user account
type UserStatus string

const (
	StatusActive  UserStatus = "ACTIVE"
	StatusInactive UserStatus = "INACTIVE"
	StatusLocked   UserStatus = "LOCKED"
)

// AdminUser represents an administrator in the system (FR-BE-DM-001)
type AdminUser struct {
	ID                  uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email               string        `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash        string        `json:"-" gorm:"not null"` // Not exposed in JSON
	Salt                string        `json:"-" gorm:"not null"` // Not exposed in JSON
	Role                AdminUserRole `json:"role" gorm:"type:varchar(50);not null"`
	FirstName           string        `json:"first_name"`
	LastName            string        `json:"last_name"`
	Status              UserStatus    `json:"status" gorm:"type:varchar(50);default:'ACTIVE'"`
	LastLoginAt         *time.Time    `json:"last_login_at"`
	FailedLoginAttempts int           `json:"failed_login_attempts" gorm:"default:0"`
	TwoFactorSecret     *string       `json:"-"` // Not exposed in JSON
	TwoFactorEnabled    bool          `json:"two_factor_enabled" gorm:"default:false"`
	CreatedAt           time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// DayType defines the type of day for a prize structure (FR-BE-DM-002)
type DayType string

const (
	DailyMonFri DayType = "DAILY_MON_FRI"
	WeeklySat   DayType = "WEEKLY_SAT"
)

// PrizeStructure represents the prize configuration for draws (FR-BE-DM-002)
type PrizeStructure struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name               string     `json:"name" gorm:"not null"`
	DayType            DayType    `json:"day_type" gorm:"type:varchar(50);not null"`
	IsActive           bool       `json:"is_active" gorm:"default:false"`
	EffectiveStartDate *time.Time `json:"effective_start_date"`
	EffectiveEndDate   *time.Time `json:"effective_end_date"`
	Version            int        `json:"version" gorm:"default:1"`
	CreatedByAdminID   uuid.UUID  `json:"created_by_admin_id" gorm:"type:uuid"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	PrizeTiers         []PrizeTier `json:"prize_tiers" gorm:"foreignKey:PrizeStructureID"` // Has Many relationship
}

// PrizeTier represents a specific prize tier within a structure (FR-BE-DM-003)
type PrizeTier struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PrizeStructureID uuid.UUID `json:"prize_structure_id" gorm:"type:uuid;index;not null"`
	TierName         string    `json:"tier_name" gorm:"not null"`
	TierDescription  *string   `json:"tier_description"`
	PrizeAmount      float64   `json:"prize_amount" gorm:"type:decimal(15,2);not null"`
	WinnerCount      int       `json:"winner_count" gorm:"not null"`
	SortOrder        int       `json:"sort_order" gorm:"not null"` // For display and processing order
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DrawStatus defines the status of a draw (FR-BE-DM-004)
type DrawStatus string

const (
	DrawStatusPending            DrawStatus = "PENDING"
	DrawStatusInProgress         DrawStatus = "IN_PROGRESS"
	DrawStatusCompleted          DrawStatus = "COMPLETED"
	DrawStatusFailed             DrawStatus = "FAILED"
	DrawStatusRedrawnInvalid     DrawStatus = "REDRAWN_INVALID"
	DrawStatusRedrawCompleted    DrawStatus = "REDRAW_COMPLETED"
)

// ExecutionType defines how a draw was executed (FR-BE-DM-004)
type ExecutionType string

const (
	ExecutionManual   ExecutionType = "MANUAL"
	ExecutionAutomated ExecutionType = "AUTOMATED"
)

// Draw represents a draw event (FR-BE-DM-004)
type Draw struct {
	ID                         uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DrawDate                   time.Time     `json:"draw_date" gorm:"type:date;index;not null"`
	PrizeStructureID           uuid.UUID     `json:"prize_structure_id" gorm:"type:uuid;not null"`
	Status                     DrawStatus    `json:"status" gorm:"type:varchar(50);not null"`
	EligibilityStartTimeUTC    time.Time     `json:"eligibility_start_time_utc"`
	EligibilityEndTimeUTC      time.Time     `json:"eligibility_end_time_utc"`
	ExecutedAtUTC              *time.Time    `json:"executed_at_utc"`
	ExecutedByAdminID          *uuid.UUID    `json:"executed_by_admin_id" gorm:"type:uuid"`
	ExecutionType              ExecutionType `json:"execution_type" gorm:"type:varchar(50)"`
	TotalOptedInMSISDNs        *int          `json:"total_opted_in_msisdns"`
	TotalEligibleMSISDNs       *int          `json:"total_eligible_msisdns"`
	TotalTickets               *int          `json:"total_tickets"`
	FailureReason              *string       `json:"failure_reason"`
	OriginalDrawIDForRedraw    *uuid.UUID    `json:"original_draw_id_for_redraw" gorm:"type:uuid"`
	CreatedAt                  time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                  time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	Winners                    []Winner      `json:"winners,omitempty" gorm:"foreignKey:DrawID"`
	PrizeStructure             PrizeStructure `json:"prize_structure,omitempty" gorm:"foreignKey:PrizeStructureID"`
}

// NotificationStatus defines the status of a winner notification (FR-BE-DM-005)
type NotificationStatus string

const (
	NotificationPending      NotificationStatus = "PENDING"
	NotificationSent         NotificationStatus = "SENT"
	NotificationFailed       NotificationStatus = "FAILED"
	NotificationDeliveredAck NotificationStatus = "DELIVERED_ACK"
	NotificationViewed       NotificationStatus = "VIEWED" // If richer status available
)

// PaymentStatus defines the payment status for a winner (FR-BE-DM-005)
type PaymentStatus string

const (
	PaymentPendingExport       PaymentStatus = "PENDING_EXPORT"
	PaymentExportedForPayment  PaymentStatus = "EXPORTED_FOR_PAYMENT"
	PaymentConfirmed           PaymentStatus = "PAYMENT_CONFIRMED"
	PaymentFailed              PaymentStatus = "PAYMENT_FAILED"
	PaymentRequiresVerification PaymentStatus = "REQUIRES_VERIFICATION"
)

// Winner represents a winner of a draw (FR-BE-DM-005)
type Winner struct {
	ID                      uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DrawID                  uuid.UUID          `json:"draw_id" gorm:"type:uuid;index;not null"`
	MSISDN                  string             `json:"msisdn" gorm:"index;not null"` // Store full, mask at presentation
	PrizeTierID             uuid.UUID          `json:"prize_tier_id" gorm:"type:uuid;not null"`
	PrizeAmountWon          float64            `json:"prize_amount_won" gorm:"type:decimal(15,2);not null"`
	SelectionOrderInTier    *int               `json:"selection_order_in_tier"`
	NotificationStatus      NotificationStatus `json:"notification_status" gorm:"type:varchar(50);default:'PENDING'"`
	NotificationSentAt      *time.Time         `json:"notification_sent_at"`
	PaymentStatus           PaymentStatus      `json:"payment_status" gorm:"type:varchar(50);default:'PENDING_EXPORT'"`
	PaymentProcessedAt      *time.Time         `json:"payment_processed_at"`
	CreatedAt               time.Time          `json:"created_at" gorm:"autoCreateTime"`
	PrizeTier               PrizeTier          `json:"prize_tier,omitempty" gorm:"foreignKey:PrizeTierID"`
}

// AuditLogOutcome defines the outcome of an audited action (FR-BE-DM-009)
type AuditLogOutcome string

const (
	OutcomeSuccess AuditLogOutcome = "SUCCESS"
	OutcomeFailure AuditLogOutcome = "FAILURE"
)

// AuditLog represents an audit trail entry (FR-BE-DM-009)
type AuditLog struct {
	ID                      uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	AdminUserID             *uuid.UUID      `json:"admin_user_id" gorm:"type:uuid;index"`
	ImpersonatedByAdminID   *uuid.UUID      `json:"impersonated_by_admin_id" gorm:"type:uuid"`
	ActionType              string          `json:"action_type" gorm:"index;not null"` // e.g., "USER_LOGIN", "DRAW_EXECUTION"
	EntityType              *string         `json:"entity_type"`                     // e.g., "Draw", "User"
	EntityID                *string         `json:"entity_id" gorm:"index"`
	TimestampUTC            time.Time       `json:"timestamp_utc" gorm:"index;not null"`
	SourceIPAddress         *string         `json:"source_ip_address"`
	UserAgent               *string         `json:"user_agent"`
	DetailsBefore           *string         `json:"details_before" gorm:"type:text"` // JSONB or Text
	DetailsAfter            *string         `json:"details_after" gorm:"type:text"`  // JSONB or Text
	Outcome                 AuditLogOutcome `json:"outcome" gorm:"type:varchar(50);not null"`
	FailureReasonShort      *string         `json:"failure_reason_short"`
	Comments                *string         `json:"comments" gorm:"type:text"`
}

// Mock external data structures (conceptual, data comes from MTN/PostHog)

// RechargeData represents a recharge transaction (FR-BE-DM-006)
type RechargeData struct {
	MSISDN             string    `json:"msisdn"`
	RechargeAmount     float64   `json:"recharge_amount"`
	RechargeTimestampUTC time.Time `json:"recharge_timestamp_utc"`
	TransactionID      string    `json:"transaction_id"`
}

// OptInData represents an opt-in record (FR-BE-DM-007)
type OptInData struct {
	MSISDN           string    `json:"msisdn"`
	OptInTimestampUTC time.Time `json:"opt_in_timestamp_utc"`
	OptInStatus      string    `json:"opt_in_status"` // e.g., "ACTIVE", "INACTIVE"
}

// BlacklistEntry represents a blacklisted MSISDN (FR-BE-DM-008)
type BlacklistEntry struct {
	MSISDN             string    `json:"msisdn"`
	Reason             *string   `json:"reason"`
	BlacklistedSinceUTC time.Time `json:"blacklisted_since_utc"`
}




// EligibleParticipant represents an MSISDN eligible for a draw with their points
type EligibleParticipant struct {
	MSISDN string `json:"msisdn"`
	Points int    `json:"points"`
}

// CreatePrizeTierRequest defines the payload for creating a prize tier
// This is often part of a CreatePrizeStructureRequest
type CreatePrizeTierRequest struct {
	TierName        string  `json:"tier_name" binding:"required"`
	TierDescription *string `json:"tier_description"`
	PrizeAmount     float64 `json:"prize_amount" binding:"required,gt=0"`
	WinnerCount     int     `json:"winner_count" binding:"required,gt=0"`
	SortOrder       int     `json:"sort_order" binding:"required,gte=0"`
}

