package gorm

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	prizeDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
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
	ID          string    `gorm:"primaryKey;type:uuid"`
	Name        string
	Description string
	IsActive    bool
	ValidFrom   time.Time
	ValidTo     *time.Time
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
	Value            float64
	ValueNGN         float64
	Quantity         int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PrizeModel is the GORM model for prizes
type PrizeModel struct {
	ID               string    `gorm:"primaryKey;type:uuid"`
	PrizeStructureID string    `gorm:"type:uuid;index"`
	Name             string
	Description      string
	Value            float64
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

// TableName returns the table name for the PrizeModel
func (PrizeModel) TableName() string {
	return "prizes"
}

// toPrizeStructureModel converts a domain prize structure entity to a GORM model
func toPrizeStructureModel(ps *prizeDomain.PrizeStructure) *PrizeStructureModel {
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
func (m *PrizeStructureModel) toDomain(prizeTiers []prizeDomain.PrizeTier) (*prizeDomain.PrizeStructure, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	return &prizeDomain.PrizeStructure{
		ID:          id,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		ValidFrom:   m.ValidFrom,
		ValidTo:     m.ValidTo,
		Prizes:      prizeTiers,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// toPrizeTierModel converts a domain prize tier entity to a GORM model
func toPrizeTierModel(pt *prizeDomain.PrizeTier) *PrizeTierModel {
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
func (m *PrizeTierModel) toDomain() (*prizeDomain.PrizeTier, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	prizeStructureID, err := uuid.Parse(m.PrizeStructureID)
	if err != nil {
		return nil, err
	}
	
	return &prizeDomain.PrizeTier{
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

// toPrizeModel converts a domain prize entity to a GORM model
func toPrizeModel(p *prizeDomain.Prize) *PrizeModel {
	return &PrizeModel{
		ID:               p.ID.String(),
		PrizeStructureID: p.PrizeStructureID.String(),
		Name:             p.Name,
		Description:      p.Description,
		Value:            p.Value,
		Quantity:         p.Quantity,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain prize entity
func (m *PrizeModel) toDomain() (*prizeDomain.Prize, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	prizeStructureID, err := uuid.Parse(m.PrizeStructureID)
	if err != nil {
		return nil, err
	}
	
	return &prizeDomain.Prize{
		ID:               id,
		PrizeStructureID: prizeStructureID,
		Name:             m.Name,
		Description:      m.Description,
		Value:            m.Value,
		Quantity:         m.Quantity,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, nil
}

// CreatePrize implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) CreatePrize(prize *prizeDomain.Prize) error {
	model := toPrizeModel(prize)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create prize: %w", result.Error)
	}
	
	return nil
}

// GetPrizeByID implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) GetPrizeByID(id uuid.UUID) (*prizeDomain.Prize, error) {
	var model PrizeModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, prizeDomain.NewPrizeError(prizeDomain.ErrPrizeNotFound, "Prize not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get prize: %w", result.Error)
	}
	
	prize, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert prize model to domain: %w", err)
	}
	
	return prize, nil
}

// ListPrizesByStructureID implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) ListPrizesByStructureID(structureID uuid.UUID) ([]prizeDomain.Prize, error) {
	var models []PrizeModel
	result := r.db.Where("prize_structure_id = ?", structureID.String()).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list prizes: %w", result.Error)
	}
	
	prizes := make([]prizeDomain.Prize, 0, len(models))
	for _, model := range models {
		prize, err := model.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert prize model to domain: %w", err)
		}
		prizes = append(prizes, *prize)
	}
	
	return prizes, nil
}

// UpdatePrize implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) UpdatePrize(prize *prizeDomain.Prize) error {
	model := toPrizeModel(prize)
	result := r.db.Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update prize: %w", result.Error)
	}
	
	return nil
}

// DeletePrize implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) DeletePrize(id uuid.UUID) error {
	result := r.db.Delete(&PrizeModel{}, "id = ?", id.String())
	if result.Error != nil {
		return fmt.Errorf("failed to delete prize: %w", result.Error)
	}
	
	return nil
}

// CreatePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) CreatePrizeStructure(prizeStructure *prizeDomain.PrizeStructure) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	// Create prize structure
	model := toPrizeStructureModel(prizeStructure)
	result := tx.Create(model)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create prize structure: %w", result.Error)
	}
	
	// Create prize tiers
	for i := range prizeStructure.Prizes {
		prizeTier := &prizeStructure.Prizes[i]
		prizeTierModel := toPrizeTierModel(prizeTier)
		result := tx.Create(prizeTierModel)
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create prize tier: %w", result.Error)
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetPrizeStructureByID implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) GetPrizeStructureByID(id uuid.UUID) (*prizeDomain.PrizeStructure, error) {
	var model PrizeStructureModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, prizeDomain.NewPrizeError(prizeDomain.ErrPrizeStructureNotFound, "Prize structure not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get prize structure: %w", result.Error)
	}
	
	// Get prize tiers for this structure
	var prizeTierModels []PrizeTierModel
	result = r.db.Where("prize_structure_id = ?", id.String()).Order("rank ASC").Find(&prizeTierModels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get prize tiers: %w", result.Error)
	}
	
	prizeTiers := make([]prizeDomain.PrizeTier, 0, len(prizeTierModels))
	for _, prizeTierModel := range prizeTierModels {
		prizeTier, err := prizeTierModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
		}
		prizeTiers = append(prizeTiers, *prizeTier)
	}
	
	prizeStructure, err := model.toDomain(prizeTiers)
	if err != nil {
		return nil, fmt.Errorf("failed to convert prize structure model to domain: %w", err)
	}
	
	return prizeStructure, nil
}

// ListPrizeStructures implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) ListPrizeStructures(page, pageSize int) ([]prizeDomain.PrizeStructure, int, error) {
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
	
	prizeStructures := make([]prizeDomain.PrizeStructure, 0, len(models))
	for _, model := range models {
		// Get prize tiers for this structure
		var prizeTierModels []PrizeTierModel
		result = r.db.Where("prize_structure_id = ?", model.ID).Order("rank ASC").Find(&prizeTierModels)
		if result.Error != nil {
			return nil, 0, fmt.Errorf("failed to get prize tiers: %w", result.Error)
		}
		
		prizeTiers := make([]prizeDomain.PrizeTier, 0, len(prizeTierModels))
		for _, prizeTierModel := range prizeTierModels {
			prizeTier, err := prizeTierModel.toDomain()
			if err != nil {
				return nil, 0, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
			}
			prizeTiers = append(prizeTiers, *prizeTier)
		}
		
		prizeStructure, err := model.toDomain(prizeTiers)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert prize structure model to domain: %w", err)
		}
		
		prizeStructures = append(prizeStructures, *prizeStructure)
	}
	
	return prizeStructures, int(total), nil
}

// UpdatePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) UpdatePrizeStructure(prizeStructure *prizeDomain.PrizeStructure) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	// Update prize structure
	model := toPrizeStructureModel(prizeStructure)
	result := tx.Save(model)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update prize structure: %w", result.Error)
	}
	
	// Delete existing prize tiers
	result = tx.Where("prize_structure_id = ?", prizeStructure.ID.String()).Delete(&PrizeTierModel{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing prize tiers: %w", result.Error)
	}
	
	// Create new prize tiers
	for i := range prizeStructure.Prizes {
		prizeTier := &prizeStructure.Prizes[i]
		prizeTierModel := toPrizeTierModel(prizeTier)
		result := tx.Create(prizeTierModel)
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create prize tier: %w", result.Error)
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// DeletePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) DeletePrizeStructure(id uuid.UUID) error {
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
	
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetActivePrizeStructure implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) GetActivePrizeStructure(date time.Time) (*prizeDomain.PrizeStructure, error) {
	var model PrizeStructureModel
	
	// Format date to match database format
	formattedDate := date.Format("2006-01-02 15:04:05")
	
	// Get active prize structure for the given date
	result := r.db.Where("is_active = ? AND valid_from <= ? AND (valid_to IS NULL OR valid_to >= ?)",
		true, formattedDate, formattedDate).
		Order("valid_from DESC").
		First(&model)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, prizeDomain.NewPrizeError(prizeDomain.ErrNoPrizeStructureActive, "No active prize structure found", result.Error)
		}
		return nil, fmt.Errorf("failed to get active prize structure: %w", result.Error)
	}
	
	// Get prize tiers for this structure
	var prizeTierModels []PrizeTierModel
	result = r.db.Where("prize_structure_id = ?", model.ID).Order("rank ASC").Find(&prizeTierModels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get prize tiers: %w", result.Error)
	}
	
	prizeTiers := make([]prizeDomain.PrizeTier, 0, len(prizeTierModels))
	for _, prizeTierModel := range prizeTierModels {
		prizeTier, err := prizeTierModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert prize tier model to domain: %w", err)
		}
		prizeTiers = append(prizeTiers, *prizeTier)
	}
	
	prizeStructure, err := model.toDomain(prizeTiers)
	if err != nil {
		return nil, fmt.Errorf("failed to convert prize structure model to domain: %w", err)
	}
	
	return prizeStructure, nil
}

// CreatePrizeTier implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) CreatePrizeTier(prizeTier *prizeDomain.PrizeTier) error {
	model := toPrizeTierModel(prizeTier)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create prize tier: %w", result.Error)
	}
	
	return nil
}

// GetPrizeTierByID implements the prize.PrizeRepository interface
func (r *GormPrizeRepository) GetPrizeTierByID(id uuid.UUID) (*prizeDomain.PrizeTier, error) {
	var model PrizeTierModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, prizeDomain.NewPrizeError(prizeDomain.ErrPrizeTierNotFound, "Prize tier not found", result.Error)
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
func (r *GormPrizeRepository) ListPrizeTiersByStructureID(structureID uuid.UUID) ([]prizeDomain.PrizeTier, error) {
	var models []PrizeTierModel
	result := r.db.Where("prize_structure_id = ?", structureID.String()).Order("rank ASC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list prize tiers: %w", result.Error)
	}
	
	prizeTiers := make([]prizeDomain.PrizeTier, 0, len(models))
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
func (r *GormPrizeRepository) UpdatePrizeTier(prizeTier *prizeDomain.PrizeTier) error {
	model := toPrizeTierModel(prizeTier)
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
