package gorm

import (
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
