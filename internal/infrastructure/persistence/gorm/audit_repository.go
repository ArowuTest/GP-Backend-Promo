package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/audit"
)

// GormAuditRepository implements the audit.AuditRepository interface using GORM
type GormAuditRepository struct {
	db *gorm.DB
}

// NewGormAuditRepository creates a new GormAuditRepository
func NewGormAuditRepository(db *gorm.DB) *GormAuditRepository {
	return &GormAuditRepository{
		db: db,
	}
}

// AuditLogModel is the GORM model for audit logs
type AuditLogModel struct {
	ID                 string    `gorm:"primaryKey;type:uuid"`
	UserID             string    `gorm:"type:uuid;index"`
	Action             string
	ActionType         string    `gorm:"column:action_type;not null;default:'SYSTEM'"`
	EntityType         string
	EntityID           string    `gorm:"type:uuid"`
	Description        string
	Metadata           string    `gorm:"type:text"`
	IPAddress          string
	UserAgent          string
	CreatedAt          time.Time `gorm:"index"`
	TimestampUTC       time.Time `gorm:"column:timestamp_utc;not null"`
	Outcome            string    `gorm:"column:outcome;not null;default:'SUCCESS'"`
	FailureReasonShort string    `gorm:"column:failure_reason_short;default:''"`
}

// SystemAuditLogModel is the GORM model for system audit logs
type SystemAuditLogModel struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	Action      string
	Description string
	Severity    string
	Source      string
	Metadata    string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"index"`
}

// TableName returns the table name for the AuditLogModel
func (AuditLogModel) TableName() string {
	return "audit_logs"
}

// TableName returns the table name for the SystemAuditLogModel
func (SystemAuditLogModel) TableName() string {
	return "system_audit_logs"
}

// toModel converts a domain audit log entity to a GORM model
func toAuditLogModel(a *audit.AuditLog) *AuditLogModel {
	// Determine outcome based on action
	outcome := "SUCCESS"
	failureReasonShort := ""
	actionType := "SYSTEM"
	
	if a.Action == "LOGIN_FAILED" {
		outcome = "FAILURE"
		failureReasonShort = "Invalid password"
		actionType = "AUTH"
	} else if a.Action == "LOGIN_SUCCESS" {
		actionType = "AUTH"
	}

	return &AuditLogModel{
		ID:                 a.ID.String(),
		UserID:             a.UserID.String(),
		Action:             a.Action,
		ActionType:         actionType,
		EntityType:         a.EntityType,
		EntityID:           a.EntityID,
		Description:        a.Description,
		Metadata:           fmt.Sprintf("%v", a.Metadata),
		IPAddress:          a.IPAddress,
		UserAgent:          a.UserAgent,
		CreatedAt:          a.CreatedAt,
		TimestampUTC:       time.Now(),
		Outcome:            outcome,
		FailureReasonShort: failureReasonShort,
	}
}

// toSystemAuditLogModel converts a domain system audit log entity to a GORM model
func toSystemAuditLogModel(a *audit.SystemAuditLog) *SystemAuditLogModel {
	return &SystemAuditLogModel{
		ID:          a.ID.String(),
		Action:      a.Action,
		Description: a.Description,
		Severity:    a.Severity,
		Source:      a.Source,
		Metadata:    fmt.Sprintf("%v", a.Metadata),
		CreatedAt:   a.CreatedAt,
	}
}

// toDomain converts a GORM model to a domain audit log entity
func (m *AuditLogModel) toDomain() (*audit.AuditLog, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	userID, err := uuid.Parse(m.UserID)
	if err != nil {
		return nil, err
	}
	
	// Create metadata map from string
	metadata := make(map[string]interface{})
	
	return &audit.AuditLog{
		ID:          id,
		UserID:      userID,
		Action:      m.Action,
		EntityType:  m.EntityType,
		EntityID:    m.EntityID,
		Description: m.Description,
		Metadata:    metadata,
		IPAddress:   m.IPAddress,
		UserAgent:   m.UserAgent,
		CreatedAt:   m.CreatedAt,
	}, nil
}

// toDomain converts a GORM model to a domain system audit log entity
func (m *SystemAuditLogModel) toDomain() (*audit.SystemAuditLog, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	// Create metadata map from string
	metadata := make(map[string]interface{})
	
	return &audit.SystemAuditLog{
		ID:          id,
		Action:      m.Action,
		Description: m.Description,
		Severity:    m.Severity,
		Source:      m.Source,
		Metadata:    metadata,
		CreatedAt:   m.CreatedAt,
	}, nil
}

// Create implements the audit.AuditRepository interface
func (r *GormAuditRepository) Create(auditLog *audit.AuditLog) error {
	model := toAuditLogModel(auditLog)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create audit log: %w", result.Error)
	}
	
	return nil
}

// CreateAuditLog implements the application.audit.Repository interface
func (r *GormAuditRepository) CreateAuditLog(ctx context.Context, auditLog *audit.AuditLog) error {
	// Delegate to the domain layer implementation
	return r.Create(auditLog)
}

// CreateSystemAuditLog implements the audit.AuditRepository interface
func (r *GormAuditRepository) CreateSystemAuditLog(systemAuditLog *audit.SystemAuditLog) error {
	model := toSystemAuditLogModel(systemAuditLog)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create system audit log: %w", result.Error)
	}
	
	return nil
}

// GetByID implements the audit.AuditRepository interface
func (r *GormAuditRepository) GetByID(id uuid.UUID) (*audit.AuditLog, error) {
	var model AuditLogModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {			
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, audit.NewAuditError(audit.ErrAuditLogNotFound, "Audit log not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", result.Error)
	}
	
	auditLogEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert audit log model to domain: %w", err)
	}
	
	return auditLogEntity, nil
}

// GetAuditLogByID implements the application.audit.Repository interface
func (r *GormAuditRepository) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*audit.AuditLog, error) {
	// Delegate to the domain layer implementation
	return r.GetByID(id)
}

// GetSystemAuditLogByID implements the audit.AuditRepository interface
func (r *GormAuditRepository) GetSystemAuditLogByID(id uuid.UUID) (*audit.SystemAuditLog, error) {
	var model SystemAuditLogModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, audit.NewAuditError(audit.ErrSystemAuditLogNotFound, "System audit log not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get system audit log: %w", result.Error)
	}
	
	systemAuditLogEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert system audit log model to domain: %w", err)
	}
	
	return systemAuditLogEntity, nil
}

// List implements the audit.AuditRepository interface
func (r *GormAuditRepository) List(filters audit.AuditLogFilters, page, pageSize int) ([]audit.AuditLog, int, error) {
	var models []AuditLogModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Build query with filters
	query := r.db.Model(&AuditLogModel{})
	
	if filters.UserID != uuid.Nil {
		query = query.Where("user_id = ?", filters.UserID.String())
	}
	
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	
	if filters.EntityType != "" {
		query = query.Where("entity_type = ?", filters.EntityType)
	}
	
	if !filters.StartDate.IsZero() {
		query = query.Where("created_at >= ?", filters.StartDate)
	}
	
	if !filters.EndDate.IsZero() {
		query = query.Where("created_at <= ?", filters.EndDate)
	}
	
	// Get total count
	result := query.Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", result.Error)
	}
	
	// Get paginated audit logs
	result = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", result.Error)
	}
	
	auditLogs := make([]audit.AuditLog, 0, len(models))
	for _, model := range models {
		auditLogEntity, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert audit log model to domain: %w", err)
		}
		auditLogs = append(auditLogs, *auditLogEntity)
	}
	
	return auditLogs, int(total), nil
}

// ListAuditLogs implements the application.audit.Repository interface
func (r *GormAuditRepository) ListAuditLogs(ctx context.Context, filters audit.AuditLogFilters, page, pageSize int) ([]*audit.AuditLog, int, error) {
	// Call the domain layer implementation and convert the result
	auditLogs, total, err := r.List(filters, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert slice of values to slice of pointers
	result := make([]*audit.AuditLog, 0, len(auditLogs))
	for i := range auditLogs {
		result = append(result, &auditLogs[i])
	}
	
	return result, total, nil
}

// ListSystemAuditLogs implements the audit.AuditRepository interface
func (r *GormAuditRepository) ListSystemAuditLogs(filters map[string]interface{}, page, pageSize int) ([]audit.SystemAuditLog, int, error) {
	var models []SystemAuditLogModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Build query with filters
	query := r.db.Model(&SystemAuditLogModel{})
	
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}
	
	// Get total count
	result := query.Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count system audit logs: %w", result.Error)
	}
	
	// Get paginated system audit logs
	result = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list system audit logs: %w", result.Error)
	}
	
	systemAuditLogs := make([]audit.SystemAuditLog, 0, len(models))
	for _, model := range models {
		systemAuditLogEntity, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert system audit log model to domain: %w", err)
		}
		systemAuditLogs = append(systemAuditLogs, *systemAuditLogEntity)
	}
	
	return systemAuditLogs, int(total), nil
}

// GetByEntityID implements the audit.AuditRepository interface
func (r *GormAuditRepository) GetByEntityID(entityType string, entityID uuid.UUID) ([]audit.AuditLog, error) {
	var models []AuditLogModel
	result := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID.String()).
		Order("created_at DESC").
		Find(&models)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs by entity ID: %w", result.Error)
	}
	
	auditLogs := make([]audit.AuditLog, 0, len(models))
	for _, model := range models {
		auditLogEntity, err := model.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert audit log model to domain: %w", err)
		}
		auditLogs = append(auditLogs, *auditLogEntity)
	}
	
	return auditLogs, nil
}

// GetByUserID implements the audit.AuditRepository interface
func (r *GormAuditRepository) GetByUserID(userID uuid.UUID) ([]audit.AuditLog, error) {
	var models []AuditLogModel
	result := r.db.Where("user_id = ?", userID.String()).
		Order("created_at DESC").
		Find(&models)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs by user ID: %w", result.Error)
	}
	
	auditLogs := make([]audit.AuditLog, 0, len(models))
	for _, model := range models {
		auditLogEntity, err := model.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert audit log model to domain: %w", err)
		}
		auditLogs = append(auditLogs, *auditLogEntity)
	}
	
	return auditLogs, nil
}

// GetDataUploadAudits implements the application.audit.Repository interface
func (r *GormAuditRepository) GetDataUploadAudits(ctx context.Context, page, pageSize int, startDate, endDate time.Time) ([]*audit.DataUploadAudit, int, error) {
	var models []AuditLogModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Build query with filters
	query := r.db.Model(&AuditLogModel{}).Where("entity_type = ?", "PARTICIPANT_UPLOAD")
	
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	
	// Get total count
	result := query.Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count data upload audits: %w", result.Error)
	}
	
	// Get paginated data upload audits
	result = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list data upload audits: %w", result.Error)
	}
	
	// Convert to domain entities
	dataUploadAudits := make([]*audit.DataUploadAudit, 0, len(models))
	for _, model := range models {
		auditLogEntity, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert audit log model to domain: %w", err)
		}
		
		// Convert AuditLog to DataUploadAudit using the correct field names from the domain entity
		dataUploadAudit := &audit.DataUploadAudit{
			ID:                  auditLogEntity.ID,
			UploadedBy:          auditLogEntity.UserID,
			UploadedAt:          auditLogEntity.CreatedAt,
			FileName:            "", // Not available in AuditLog
			TotalUploaded:       0, // Not available in AuditLog
			SuccessfullyImported: 0, // Not available in AuditLog
			DuplicatesSkipped:   0, // Not available in AuditLog
			ErrorsEncountered:   0, // Not available in AuditLog
			Status:              "Completed", // Default status
			Details:             "", // Not available in AuditLog
			OperationType:       "Upload", // Default operation type
			CreatedAt:           auditLogEntity.CreatedAt,
			UpdatedAt:           time.Now(),
		}
		
		dataUploadAudits = append(dataUploadAudits, dataUploadAudit)
	}
	
	return dataUploadAudits, int(total), nil
}
