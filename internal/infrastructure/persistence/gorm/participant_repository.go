package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
)

// GormParticipantRepository implements the participant.ParticipantRepository interface using GORM
type GormParticipantRepository struct {
	db *gorm.DB
}

// NewGormParticipantRepository creates a new GormParticipantRepository
func NewGormParticipantRepository(db *gorm.DB) *GormParticipantRepository {
	return &GormParticipantRepository{
		db: db,
	}
}

// ParticipantModel is the GORM model for participants
type ParticipantModel struct {
	ID             string    `gorm:"primaryKey;type:uuid"`
	MSISDN         string    `gorm:"index"`
	Points         int
	RechargeAmount float64
	RechargeDate   time.Time `gorm:"index"`
	UploadID       string    `gorm:"type:uuid"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UploadAuditModel is the GORM model for upload audits
type UploadAuditModel struct {
	ID              string    `gorm:"primaryKey;type:uuid"`
	UploadedBy      string    `gorm:"type:uuid"`
	UploadDate      time.Time
	FileName        string
	Status          string
	TotalRows       int
	SuccessfulRows  int
	ErrorCount      int
	ErrorDetails    []string `gorm:"-"` // Not stored directly in the database
	ErrorDetailsStr string   `gorm:"column:error_details"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedBy       *string   `gorm:"type:uuid"`
	DeletedAt       *time.Time
}

// TableName returns the table name for the ParticipantModel
func (ParticipantModel) TableName() string {
	return "participants"
}

// TableName returns the table name for the UploadAuditModel
func (UploadAuditModel) TableName() string {
	return "upload_audits"
}

// toModel converts a domain participant entity to a GORM model
func toParticipantModel(p *participant.Participant) *ParticipantModel {
	return &ParticipantModel{
		ID:             p.ID.String(),
		MSISDN:         p.MSISDN,
		Points:         p.Points,
		RechargeAmount: p.RechargeAmount,
		RechargeDate:   p.RechargeDate,
		UploadID:       p.UploadID.String(),
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain participant entity
func (m *ParticipantModel) toDomain() (*participant.Participant, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	uploadID, err := uuid.Parse(m.UploadID)
	if err != nil {
		return nil, err
	}
	
	return &participant.Participant{
		ID:             id,
		MSISDN:         m.MSISDN,
		Points:         m.Points,
		RechargeAmount: m.RechargeAmount,
		RechargeDate:   m.RechargeDate,
		UploadID:       uploadID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}, nil
}

// toUploadAuditModel converts a domain upload audit entity to a GORM model
func toUploadAuditModel(a *participant.UploadAudit) *UploadAuditModel {
	// Convert error details slice to string for storage
	errorDetailsStr := ""
	for i, detail := range a.ErrorDetails {
		if i > 0 {
			errorDetailsStr += "\n"
		}
		errorDetailsStr += detail
	}
	
	return &UploadAuditModel{
		ID:              a.ID.String(),
		UploadedBy:      a.UploadedBy.String(),
		UploadDate:      a.UploadDate,
		FileName:        a.FileName,
		Status:          a.Status,
		TotalRows:       a.TotalRows,
		SuccessfulRows:  a.SuccessfulRows,
		ErrorCount:      a.ErrorCount,
		ErrorDetails:    a.ErrorDetails,
		ErrorDetailsStr: errorDetailsStr,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain upload audit entity
func (m *UploadAuditModel) toDomain() (*participant.UploadAudit, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	uploadedBy, err := uuid.Parse(m.UploadedBy)
	if err != nil {
		return nil, err
	}
	
	// Convert error details string to slice
	var errorDetails []string
	if m.ErrorDetailsStr != "" {
		errorDetails = []string{m.ErrorDetailsStr}
	} else {
		errorDetails = []string{}
	}
	
	return &participant.UploadAudit{
		ID:             id,
		UploadedBy:     uploadedBy,
		UploadDate:     m.UploadDate,
		FileName:       m.FileName,
		Status:         m.Status,
		TotalRows:      m.TotalRows,
		SuccessfulRows: m.SuccessfulRows,
		ErrorCount:     m.ErrorCount,
		ErrorDetails:   errorDetails,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}, nil
}

// Create implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) Create(participant *participant.Participant) error {
	model := toParticipantModel(participant)
	
	result := r.db.Create(&model)
	if result.Error != nil {
		return fmt.Errorf("failed to create participant: %w", result.Error)
	}
	
	return nil
}

// GetByMSISDN implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) GetByMSISDN(msisdn string) (*participant.Participant, error) {
	var model ParticipantModel
	result := r.db.Where("msisdn = ?", msisdn).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, participant.NewParticipantError(participant.ErrParticipantNotFound, "Participant not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get participant: %w", result.Error)
	}
	
	participantEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert participant model to domain: %w", err)
	}
	
	return participantEntity, nil
}

// GetParticipantByMSISDN implements the application.participant.Repository interface
func (r *GormParticipantRepository) GetParticipantByMSISDN(ctx context.Context, msisdn string) (*participant.Participant, error) {
	// Delegate to the domain layer implementation
	return r.GetByMSISDN(msisdn)
}

// GetByMSISDNAndDate implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) GetByMSISDNAndDate(msisdn string, date time.Time) (*participant.Participant, error) {
	var model ParticipantModel
	
	// Format date to match database format (without time component)
	formattedDate := date.Format("2006-01-02")
	
	result := r.db.Where("msisdn = ? AND DATE(recharge_date) = ?", msisdn, formattedDate).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, participant.NewParticipantError(participant.ErrParticipantNotFound, "Participant not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get participant: %w", result.Error)
	}
	
	participantEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert participant model to domain: %w", err)
	}
	
	return participantEntity, nil
}

// List implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) List(page, pageSize int) ([]participant.Participant, int, error) {
	var models []ParticipantModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Get total count
	result := r.db.Model(&ParticipantModel{}).Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count participants: %w", result.Error)
	}
	
	// Get paginated participants
	result = r.db.Order("recharge_date DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list participants: %w", result.Error)
	}
	
	participants := make([]participant.Participant, 0, len(models))
	for _, model := range models {
		participantEntity, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert participant model to domain: %w", err)
		}
		participants = append(participants, *participantEntity)
	}
	
	return participants, int(total), nil
}

// ListParticipants implements the application.participant.Repository interface
func (r *GormParticipantRepository) ListParticipants(ctx context.Context, page, pageSize int) ([]*participant.Participant, int, error) {
	// Call the domain layer implementation and convert the result
	participants, total, err := r.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert slice of values to slice of pointers
	result := make([]*participant.Participant, 0, len(participants))
	for i := range participants {
		result = append(result, &participants[i])
	}
	
	return result, total, nil
}

// ListByDate implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) ListByDate(date time.Time, page, pageSize int) ([]participant.Participant, int, error) {
	var models []ParticipantModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Format date to match database format (without time component)
	formattedDate := date.Format("2006-01-02")
	
	// Get total count
	result := r.db.Model(&ParticipantModel{}).Where("DATE(recharge_date) = ?", formattedDate).Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count participants: %w", result.Error)
	}
	
	// Get paginated participants
	result = r.db.Where("DATE(recharge_date) = ?", formattedDate).
		Order("recharge_date DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&models)
	
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list participants: %w", result.Error)
	}
	
	participants := make([]participant.Participant, 0, len(models))
	for _, model := range models {
		participantEntity, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert participant model to domain: %w", err)
		}
		participants = append(participants, *participantEntity)
	}
	
	return participants, int(total), nil
}

// GetStats implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) GetStats(date time.Time) (int, int, float64, error) {
	var totalParticipants int64
	var totalPoints int64
	var totalUploads int64
	
	// Format date to match database format (without time component)
	formattedDate := date.Format("2006-01-02")
	
	// Count distinct MSISDNs for the given date
	result := r.db.Model(&ParticipantModel{}).
		Where("DATE(recharge_date) <= ?", formattedDate).
		Distinct("msisdn").
		Count(&totalParticipants)
	
	if result.Error != nil {
		return 0, 0, 0, fmt.Errorf("failed to count participants: %w", result.Error)
	}
	
	// Sum points for the given date
	var err error
	err = r.db.Model(&ParticipantModel{}).
		Where("DATE(recharge_date) <= ?", formattedDate).
		Select("SUM(points)").
		Row().
		Scan(&totalPoints)
	
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to sum points: %w", err)
	}
	
	// Count distinct upload IDs
	result = r.db.Model(&ParticipantModel{}).
		Where("DATE(recharge_date) <= ?", formattedDate).
		Distinct("upload_id").
		Count(&totalUploads)
	
	if result.Error != nil {
		return 0, 0, 0, fmt.Errorf("failed to count uploads: %w", result.Error)
	}
	
	return int(totalParticipants), int(totalPoints), float64(totalUploads), nil
}

// GetParticipantStats implements the application.participant.Repository interface
func (r *GormParticipantRepository) GetParticipantStats(ctx context.Context, date time.Time) (int, int, float64, error) {
	// Delegate to the domain layer implementation
	return r.GetStats(date)
}

// CreateBatch implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) CreateBatch(participants []*participant.Participant) (int, []string, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return 0, nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	successCount := 0
	errorDetails := make([]string, 0)
	
	for _, participant := range participants {
		model := toParticipantModel(participant)
		result := tx.Create(model)
		if result.Error != nil {
			errorDetails = append(errorDetails, fmt.Sprintf("Failed to create participant with MSISDN %s: %s", participant.MSISDN, result.Error.Error()))
			continue
		}
		successCount++
	}
	
	if err := tx.Commit().Error; err != nil {
		return 0, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return successCount, errorDetails, nil
}

// DeleteByUploadID implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) DeleteByUploadID(uploadID uuid.UUID) error {
	// This would typically involve a join or subquery to identify participants from a specific upload
	// For demonstration purposes, we'll use a simple approach
	
	result := r.db.Where("upload_id = ?", uploadID.String()).Delete(&ParticipantModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete participants: %w", result.Error)
	}
	
	return nil
}

// DeleteUpload implements the application.participant.Repository interface
// Updated to include context parameter as required by the application layer
func (r *GormParticipantRepository) DeleteUpload(ctx context.Context, uploadID uuid.UUID, deletedBy uuid.UUID) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	// Delete participants with this upload ID
	if err := r.DeleteByUploadID(uploadID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete participants: %w", err)
	}
	
	// Mark the upload audit as deleted
	deletedByStr := deletedBy.String()
	now := time.Now()
	
	result := tx.Model(&UploadAuditModel{}).
		Where("id = ?", uploadID.String()).
		Updates(map[string]interface{}{
			"deleted_by": deletedByStr,
			"deleted_at": now,
			"status":     "Deleted",
		})
	
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update upload audit: %w", result.Error)
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetStatsByDate implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) GetStatsByDate(date time.Time) (int, int, error) {
	var totalParticipants int64
	var totalPoints int64
	
	// Format date to match database format (without time component)
	formattedDate := date.Format("2006-01-02")
	
	// Count distinct MSISDNs for the given date
	result := r.db.Model(&ParticipantModel{}).
		Where("DATE(recharge_date) = ?", formattedDate).
		Distinct("msisdn").
		Count(&totalParticipants)
	
	if result.Error != nil {
		return 0, 0, fmt.Errorf("failed to count participants: %w", result.Error)
	}
	
	// Sum points for the given date
	var err error
	err = r.db.Model(&ParticipantModel{}).
		Where("DATE(recharge_date) = ?", formattedDate).
		Select("SUM(points)").
		Row().
		Scan(&totalPoints)
	
	if err != nil {
		return 0, 0, fmt.Errorf("failed to sum points: %w", err)
	}
	
	return int(totalParticipants), int(totalPoints), nil
}

// BulkCreate implements the participant.ParticipantRepository interface
func (r *GormParticipantRepository) BulkCreate(participants []*participant.Participant) (int, []string, error) {
	// This is an alias for CreateBatch for backward compatibility
	return r.CreateBatch(participants)
}

// UploadParticipants implements the application.participant.Repository interface
func (r *GormParticipantRepository) UploadParticipants(ctx context.Context, participants []*participant.ParticipantInput, uploadedBy uuid.UUID, fileName string) (*participant.UploadAudit, error) {
	// Create a new upload audit record
	uploadID := uuid.New()
	now := time.Now()
	
	// Convert participant inputs to domain participants
	domainParticipants := make([]*participant.Participant, 0, len(participants))
	for _, p := range participants {
		domainParticipants = append(domainParticipants, &participant.Participant{
			ID:             uuid.New(),
			MSISDN:         p.MSISDN,
			Points:         p.Points,
			RechargeAmount: p.RechargeAmount,
			RechargeDate:   p.RechargeDate,
			UploadID:       uploadID,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	
	// Create participants in batch
	successCount, errorDetails, err := r.CreateBatch(domainParticipants)
	if err != nil {
		return nil, fmt.Errorf("failed to create participants: %w", err)
	}
	
	// Create upload audit record
	audit := &participant.UploadAudit{
		ID:             uploadID,
		UploadedBy:     uploadedBy,
		UploadDate:     now,
		FileName:       fileName,
		Status:         "Completed",
		TotalRows:      len(participants),
		SuccessfulRows: successCount,
		ErrorCount:     len(errorDetails),
		ErrorDetails:   errorDetails,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	
	// Save audit record to database
	auditModel := toUploadAuditModel(audit)
	result := r.db.Create(auditModel)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create upload audit: %w", result.Error)
	}
	
	return audit, nil
}

// ListUploadAudits implements the application.participant.Repository interface
func (r *GormParticipantRepository) ListUploadAudits(ctx context.Context, page, pageSize int) ([]*participant.UploadAudit, int, error) {
	var models []UploadAuditModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Get total count
	result := r.db.Model(&UploadAuditModel{}).Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count upload audits: %w", result.Error)
	}
	
	// Get paginated upload audits
	result = r.db.Order("upload_date DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list upload audits: %w", result.Error)
	}
	
	audits := make([]*participant.UploadAudit, 0, len(models))
	for _, model := range models {
		audit, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert upload audit model to domain: %w", err)
		}
		audits = append(audits, audit)
	}
	
	return audits, int(total), nil
}
