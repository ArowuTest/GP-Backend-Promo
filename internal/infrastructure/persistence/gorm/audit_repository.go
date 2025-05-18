package infrastructure

import (
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
	ID          string    `gorm:"primaryKey;type:uuid"`
	UserID      string    `gorm:"type:uuid;index"`
	Username    string
	Action      string    `gorm:"index"`
	EntityType  string    `gorm:"index"`
	EntityID    string    `gorm:"index"`
	Description string
	IPAddress   string
	UserAgent   string
	MetadataJSON string   `gorm:"column:metadata"`
	CreatedAt   time.Time `gorm:"index"`
}

// TableName returns the table name for the AuditLogModel
func (AuditLogModel) TableName() string {
	return "audit_logs"
}

// toModel converts a domain audit log entity to a GORM model
func toAuditLogModel(a *audit.AuditLog) *AuditLogModel {
	// In a real implementation, we would convert the Metadata map to JSON
	metadataJSON := "{}" // Simplified for this example
	
	return &AuditLogModel{
		ID:          a.ID.String(),
		UserID:      a.UserID.String(),
		Username:    a.Username,
		Action:      a.Action,
		EntityType:  a.EntityType,
		EntityID:    a.EntityID,
		Description: a.Description,
		IPAddress:   a.IPAddress,
		UserAgent:   a.UserAgent,
		MetadataJSON: metadataJSON,
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
	
	// In a real implementation, we would parse the JSON from MetadataJSON
	metadata := map[string]interface{}{} // Simplified for this example
	
	return &audit.AuditLog{
		ID:          id,
		UserID:      userID,
		Username:    m.Username,
		Action:      m.Action,
		EntityType:  m.EntityType,
		EntityID:    m.EntityID,
		Description: m.Description,
		IPAddress:   m.IPAddress,
		UserAgent:   m.UserAgent,
		Metadata:    metadata,
		CreatedAt:   m.CreatedAt,
	}, nil
}

// CreateAuditLog implements the audit.AuditRepository interface
func (r *GormAuditRepository) CreateAuditLog(a *audit.AuditLog) error {
	model := toAuditLogModel(a)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create audit log: %w", result.Error)
	}
	
	return nil
}

// GetAuditLogByID implements the audit.AuditRepository interface
func (r *GormAuditRepository) GetAuditLogByID(id uuid.UUID) (*audit.AuditLog, error) {
	var model AuditLogModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, audit.NewAuditError(audit.ErrAuditLogNotFound, "Audit log not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", result.Error)
	}
	
	auditLog, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert audit log model to domain: %w", err)
	}
	
	return auditLog, nil
}

// ListAuditLogs implements the audit.AuditRepository interface
func (r *GormAuditRepository) ListAuditLogs(filters audit.AuditLogFilters, page, pageSize int) ([]audit.AuditLog, int, error) {
	var models []AuditLogModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Build query with filters
	query := r.db.Model(&AuditLogModel{})
	
	if filters.UserID != nil {
		query = query.Where("user_id = ?", filters.UserID.String())
	}
	
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	
	if filters.EntityType != "" {
		query = query.Where("entity_type = ?", filters.EntityType)
	}
	
	if filters.EntityID != "" {
		query = query.Where("entity_id = ?", filters.EntityID)
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
		auditLog, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert audit log model to domain: %w", err)
		}
		auditLogs = append(auditLogs, *auditLog)
	}
	
	return auditLogs, int(total), nil
}

// GetUserActivitySummary implements the audit.AuditRepository interface
func (r *GormAuditRepository) GetUserActivitySummary(userID uuid.UUID, startDate, endDate time.Time) (map[string]int, error) {
	type ActionCount struct {
		Action string
		Count  int
	}
	
	var actionCounts []ActionCount
	
	result := r.db.Model(&AuditLogModel{}).
		Select("action, COUNT(*) as count").
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID.String(), startDate, endDate).
		Group("action").
		Find(&actionCounts)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user activity summary: %w", result.Error)
	}
	
	summary := make(map[string]int)
	for _, ac := range actionCounts {
		summary[ac.Action] = ac.Count
	}
	
	return summary, nil
}

// GetSystemActivitySummary implements the audit.AuditRepository interface
func (r *GormAuditRepository) GetSystemActivitySummary(startDate, endDate time.Time) (map[string]int, error) {
	type ActionCount struct {
		Action string
		Count  int
	}
	
	var actionCounts []ActionCount
	
	result := r.db.Model(&AuditLogModel{}).
		Select("action, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("action").
		Find(&actionCounts)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get system activity summary: %w", result.Error)
	}
	
	summary := make(map[string]int)
	for _, ac := range actionCounts {
		summary[ac.Action] = ac.Count
	}
	
	return summary, nil
}
