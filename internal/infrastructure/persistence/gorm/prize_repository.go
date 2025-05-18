package infrastructure

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// GormPrizeRepository implements the prize.PrizeRepository interface using GORM
type GormPrizeRepository struct {
	db *gorm.DB
}

// NewGormPrizeRepository creates a new GormPrizeRepository
func NewGormPrizeRepository(db *gorm.DB) *GormPrizeRepository {
	return &GormPrizeRepository{
		db: db,
	}
}

// PrizeStructureModel is the GORM model for prize structures
type PrizeStructureModel struct {
	ID          string     `gorm:"primaryKey;type:uuid"`
	Name        string
	Description string
	IsActive    bool
	ValidFrom   time.Time  `gorm:"index"`
	ValidTo     *time.Time `gorm:"index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PrizeTierModel is the GORM model for prize tiers
type PrizeTierModel struct {
	ID               string    `gorm:"primaryKey;type:uuid"`
	PrizeStructureID string    `gorm:"type:uuid;index"`
	Rank             int
	Name             string
	Description      string
	Value            string
	ValueNGN         float64
	Quantity         int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TableName returns the table name for the PrizeStructureModel
func (PrizeStructureModel) TableName() string {
	return "prize_structures"
}

// TableName returns the table name for the PrizeTierModel
func (PrizeTierModel) TableName() string {
	return "prize_tiers"
}

// toPrizeStructureModel converts a domain prize structure entity to a GORM model
func toPrizeStructureModel(ps *prize.PrizeStructure) *PrizeStructureModel {
	return &PrizeStructureModel{
		ID:          ps.ID.String(),
		Name:        ps.Name,
		Description: ps.Description,
		IsActive:    ps.IsActive,
		ValidFrom:   ps.ValidFrom,
		ValidTo:     ps.ValidTo,
		CreatedAt:   ps.CreatedAt,
		UpdatedAt:   ps.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain prize structure entity
func (m *PrizeStructureModel) toDomain(tiers []prize.PrizeTier) (*prize.PrizeStructure, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	return &prize.PrizeStructure{
		ID:          id,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		ValidFrom:   m.ValidFrom,
		ValidTo:     m.ValidTo,
		Prizes:      tiers,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// toPrizeTierModel converts a domain prize tier entity to a GORM model
func toPrizeTierModel(pt *prize.PrizeTier) *PrizeTierModel {
	return &PrizeTierModel{
		ID:               pt.ID.String(),
		PrizeStructureID: pt.PrizeStructureID.String(),
		Rank:             pt.Rank,
		Name:             pt.Name,
		Description:      pt.Description,
		Value:            pt.Value,
		ValueNGN:         pt.ValueNGN,
		Quantity:         pt.Quantity,
		CreatedAt:        pt.CreatedAt,
		UpdatedAt:        pt.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain prize tier entity
func (m *PrizeTierModel) toDomain() (*prize.PrizeTier, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	prizeStructureID, err := uuid.Parse(m.PrizeStructureID)
	if err != nil {
		return nil, err
	}
	
	return &prize.PrizeTier{
		ID:               id,
		PrizeStructureID: prizeStructureID,
		Rank:             m.Rank,
		Name:             m.Name,
		Description:      m.Description,
		Value:            m.Value,
		ValueNGN:         m.ValueNGN,
		Quantity:         m.Quantity,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, nil
}

// CreatePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) CreatePrizeStructure(ps *prize.PrizeStructure) error {
	// Start a transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	// Create prize structure
	structureModel := toPrizeStructureModel(ps)
	result := tx.Create(structureModel)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create prize structure: %w", result.Error)
	}
	
	// Create prize tiers
	for _, tier := range ps.Prizes {
		tierModel := toPrizeTierModel(&tier)
		result := tx.Create(tierModel)
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create prize tier: %w", result.Error)
		}
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetPrizeStructureByID implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) GetPrizeStructureByID(id uuid.UUID) (*prize.PrizeStructure, error) {
	var model PrizeStructureModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, prize.NewPrizeError(prize.ErrPrizeStructureNotFound, "Prize structure not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get prize structure: %w", result.Error)
	}
	
	// Get prize tiers for this structure
	var tierModels []PrizeTierModel
	result = r.db.Where("prize_structure_id = ?", id.String()).Order("rank ASC").Find(&tierModels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get prize tiers: %w", result.Error)
	}
	
	tiers := make([]prize.PrizeTier, 0, len(tierModels))
	for _, tierModel := range tierModels {
		tier, err := tierModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
		}
		tiers = append(tiers, *tier)
	}
	
	prizeStructure, err := model.toDomain(tiers)
	if err != nil {
		return nil, fmt.Errorf("failed to convert prize structure model to domain: %w", err)
	}
	
	return prizeStructure, nil
}

// ListPrizeStructures implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) ListPrizeStructures(page, pageSize int) ([]prize.PrizeStructure, int, error) {
	var models []PrizeStructureModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Get total count
	result := r.db.Model(&PrizeStructureModel{}).Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count prize structures: %w", result.Error)
	}
	
	// Get paginated prize structures
	result = r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list prize structures: %w", result.Error)
	}
	
	prizeStructures := make([]prize.PrizeStructure, 0, len(models))
	for _, model := range models {
		// Get prize tiers for this structure
		var tierModels []PrizeTierModel
		result = r.db.Where("prize_structure_id = ?", model.ID).Order("rank ASC").Find(&tierModels)
		if result.Error != nil {
			return nil, 0, fmt.Errorf("failed to get prize tiers: %w", result.Error)
		}
		
		tiers := make([]prize.PrizeTier, 0, len(tierModels))
		for _, tierModel := range tierModels {
			tier, err := tierModel.toDomain()
			if err != nil {
				return nil, 0, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
			}
			tiers = append(tiers, *tier)
		}
		
		prizeStructure, err := model.toDomain(tiers)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert prize structure model to domain: %w", err)
		}
		
		prizeStructures = append(prizeStructures, *prizeStructure)
	}
	
	return prizeStructures, int(total), nil
}

// UpdatePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) UpdatePrizeStructure(ps *prize.PrizeStructure) error {
	// Start a transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	// Update prize structure
	structureModel := toPrizeStructureModel(ps)
	result := tx.Save(structureModel)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update prize structure: %w", result.Error)
	}
	
	// Delete existing prize tiers
	result = tx.Where("prize_structure_id = ?", ps.ID.String()).Delete(&PrizeTierModel{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing prize tiers: %w", result.Error)
	}
	
	// Create new prize tiers
	for _, tier := range ps.Prizes {
		tierModel := toPrizeTierModel(&tier)
		result := tx.Create(tierModel)
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create prize tier: %w", result.Error)
		}
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// DeletePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) DeletePrizeStructure(id uuid.UUID) error {
	// Start a transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	// Delete prize tiers
	result := tx.Where("prize_structure_id = ?", id.String()).Delete(&PrizeTierModel{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete prize tiers: %w", result.Error)
	}
	
	// Delete prize structure
	result = tx.Delete(&PrizeStructureModel{}, "id = ?", id.String())
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete prize structure: %w", result.Error)
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetActivePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) GetActivePrizeStructure(date time.Time) (*prize.PrizeStructure, error) {
	var model PrizeStructureModel
	
	// Format date to match database format
	formattedDate := date.Format("2006-01-02 15:04:05")
	
	// Find active prize structure for the given date
	result := r.db.Where("is_active = ? AND valid_from <= ? AND (valid_to IS NULL OR valid_to >= ?)", 
		true, formattedDate, formattedDate).First(&model)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, prize.NewPrizeError(prize.ErrNoPrizeStructureActive, "No active prize structure found for the given date", result.Error)
		}
		return nil, fmt.Errorf("failed to get active prize structure: %w", result.Error)
	}
	
	// Get prize tiers for this structure
	var tierModels []PrizeTierModel
	result = r.db.Where("prize_structure_id = ?", model.ID).Order("rank ASC").Find(&tierModels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get prize tiers: %w", result.Error)
	}
	
	tiers := make([]prize.PrizeTier, 0, len(tierModels))
	for _, tierModel := range tierModels {
		tier, err := tierModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
		}
		tiers = append(tiers, *tier)
	}
	
	prizeStructure, err := model.toDomain(tiers)
	if err != nil {
		return nil, fmt.Errorf("failed to convert prize structure model to domain: %w", err)
	}
	
	return prizeStructure, nil
}

// CreatePrizeTier implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) CreatePrizeTier(pt *prize.PrizeTier) error {
	model := toPrizeTierModel(pt)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create prize tier: %w", result.Error)
	}
	
	return nil
}

// GetPrizeTierByID implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) GetPrizeTierByID(id uuid.UUID) (*prize.PrizeTier, error) {
	var model PrizeTierModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, prize.NewPrizeError(prize.ErrPrizeTierNotFound, "Prize tier not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get prize tier: %w", result.Error)
	}
	
	prizeTier, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
	}
	
	return prizeTier, nil
}

// ListPrizeTiersByStructureID implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) ListPrizeTiersByStructureID(structureID uuid.UUID) ([]prize.PrizeTier, error) {
	var models []PrizeTierModel
	result := r.db.Where("prize_structure_id = ?", structureID.String()).Order("rank ASC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list prize tiers: %w", result.Error)
	}
	
	prizeTiers := make([]prize.PrizeTier, 0, len(models))
	for _, model := range models {
		prizeTier, err := model.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
		}
		prizeTiers = append(prizeTiers, *prizeTier)
	}
	
	return prizeTiers, nil
}

// UpdatePrizeTier implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) UpdatePrizeTier(pt *prize.PrizeTier) error {
	model := toPrizeTierModel(pt)
	result := r.db.Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update prize tier: %w", result.Error)
	}
	
	return nil
}

// DeletePrizeTier implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) DeletePrizeTier(id uuid.UUID) error {
	result := r.db.Delete(&PrizeTierModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("failed to delete prize tier: %w", result.Error)
	}
	
	return nil
}
