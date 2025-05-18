package gorm

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// GormDrawRepository implements the draw.DrawRepository interface using GORM
type GormDrawRepository struct {
	db *gorm.DB
}

// NewGormDrawRepository creates a new GormDrawRepository
func NewGormDrawRepository(db *gorm.DB) *GormDrawRepository {
	return &GormDrawRepository{
		db: db,
	}
}

// DrawModel is the GORM model for draws
type DrawModel struct {
	ID                    string    `gorm:"primaryKey;type:uuid"`
	DrawDate              time.Time `gorm:"index"`
	PrizeStructureID      string    `gorm:"type:uuid"`
	Status                string
	TotalEligibleMSISDNs  int
	TotalEntries          int
	ExecutedByAdminID     string    `gorm:"type:uuid"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// WinnerModel is the GORM model for winners
type WinnerModel struct {
	ID            string    `gorm:"primaryKey;type:uuid"`
	DrawID        string    `gorm:"type:uuid;index"`
	MSISDN        string
	PrizeTierID   string    `gorm:"type:uuid"`
	Status        string
	PaymentStatus string
	PaymentNotes  string
	PaidAt        *time.Time
	IsRunnerUp    bool
	RunnerUpRank  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName returns the table name for the DrawModel
func (DrawModel) TableName() string {
	return "draws"
}

// TableName returns the table name for the WinnerModel
func (WinnerModel) TableName() string {
	return "winners"
}

// toModel converts a domain draw entity to a GORM model
func toDrawModel(d *draw.Draw) *DrawModel {
	return &DrawModel{
		ID:                    d.ID.String(),
		DrawDate:              d.DrawDate,
		PrizeStructureID:      d.PrizeStructureID.String(),
		Status:                d.Status,
		TotalEligibleMSISDNs:  d.TotalEligibleMSISDNs,
		TotalEntries:          d.TotalEntries,
		ExecutedByAdminID:     d.ExecutedByAdminID.String(),
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain draw entity
func (m *DrawModel) toDomain() (*draw.Draw, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	prizeStructureID, err := uuid.Parse(m.PrizeStructureID)
	if err != nil {
		return nil, err
	}
	
	executedByAdminID, err := uuid.Parse(m.ExecutedByAdminID)
	if err != nil {
		return nil, err
	}
	
	return &draw.Draw{
		ID:                    id,
		DrawDate:              m.DrawDate,
		PrizeStructureID:      prizeStructureID,
		Status:                m.Status,
		TotalEligibleMSISDNs:  m.TotalEligibleMSISDNs,
		TotalEntries:          m.TotalEntries,
		ExecutedByAdminID:     executedByAdminID,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
		Winners:               []draw.Winner{}, // Will be populated separately
	}, nil
}

// toWinnerModel converts a domain winner entity to a GORM model
func toWinnerModel(w *draw.Winner) *WinnerModel {
	return &WinnerModel{
		ID:            w.ID.String(),
		DrawID:        w.DrawID.String(),
		MSISDN:        w.MSISDN,
		PrizeTierID:   w.PrizeTierID.String(),
		Status:        w.Status,
		PaymentStatus: w.PaymentStatus,
		PaymentNotes:  w.PaymentNotes,
		PaidAt:        w.PaidAt,
		IsRunnerUp:    w.IsRunnerUp,
		RunnerUpRank:  w.RunnerUpRank,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

// toDomain converts a GORM model to a domain winner entity
func (m *WinnerModel) toDomain() (*draw.Winner, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	
	drawID, err := uuid.Parse(m.DrawID)
	if err != nil {
		return nil, err
	}
	
	prizeTierID, err := uuid.Parse(m.PrizeTierID)
	if err != nil {
		return nil, err
	}
	
	return &draw.Winner{
		ID:            id,
		DrawID:        drawID,
		MSISDN:        m.MSISDN,
		PrizeTierID:   prizeTierID,
		Status:        m.Status,
		PaymentStatus: m.PaymentStatus,
		PaymentNotes:  m.PaymentNotes,
		PaidAt:        m.PaidAt,
		IsRunnerUp:    m.IsRunnerUp,
		RunnerUpRank:  m.RunnerUpRank,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}

// Create implements the draw.DrawRepository interface
func (r *GormDrawRepository) Create(d *draw.Draw) error {
	model := toDrawModel(d)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create draw: %w", result.Error)
	}
	
	// Create winners if any
	for _, winner := range d.Winners {
		winnerModel := toWinnerModel(&winner)
		result := r.db.Create(winnerModel)
		if result.Error != nil {
			return fmt.Errorf("failed to create winner: %w", result.Error)
		}
	}
	
	return nil
}

// GetByID implements the draw.DrawRepository interface
func (r *GormDrawRepository) GetByID(id uuid.UUID) (*draw.Draw, error) {
	var model DrawModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, draw.NewDrawError(draw.ErrDrawNotFound, "Draw not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get draw: %w", result.Error)
	}
	
	drawEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert draw model to domain: %w", err)
	}
	
	// Get winners for this draw
	var winnerModels []WinnerModel
	result = r.db.Where("draw_id = ?", id.String()).Find(&winnerModels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get winners: %w", result.Error)
	}
	
	winners := make([]draw.Winner, 0, len(winnerModels))
	for _, winnerModel := range winnerModels {
		winner, err := winnerModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert winner model to domain: %w", err)
		}
		winners = append(winners, *winner)
	}
	
	drawEntity.Winners = winners
	
	return drawEntity, nil
}

// List implements the draw.DrawRepository interface
func (r *GormDrawRepository) List(page, pageSize int) ([]draw.Draw, int, error) {
	var models []DrawModel
	var total int64
	
	offset := (page - 1) * pageSize
	
	// Get total count
	result := r.db.Model(&DrawModel{}).Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count draws: %w", result.Error)
	}
	
	// Get paginated draws
	result = r.db.Order("draw_date DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list draws: %w", result.Error)
	}
	
	draws := make([]draw.Draw, 0, len(models))
	for _, model := range models {
		drawEntity, err := model.toDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert draw model to domain: %w", err)
		}
		draws = append(draws, *drawEntity)
	}
	
	return draws, int(total), nil
}

// GetByDate implements the draw.DrawRepository interface
func (r *GormDrawRepository) GetByDate(date time.Time) (*draw.Draw, error) {
	var model DrawModel
	
	// Format date to match database format (without time component)
	formattedDate := date.Format("2006-01-02")
	
	result := r.db.Where("DATE(draw_date) = ?", formattedDate).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No draw found for this date, which is not an error
		}
		return nil, fmt.Errorf("failed to get draw by date: %w", result.Error)
	}
	
	drawEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert draw model to domain: %w", err)
	}
	
	// Get winners for this draw
	var winnerModels []WinnerModel
	result = r.db.Where("draw_id = ?", model.ID).Find(&winnerModels)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get winners: %w", result.Error)
	}
	
	winners := make([]draw.Winner, 0, len(winnerModels))
	for _, winnerModel := range winnerModels {
		winner, err := winnerModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert winner model to domain: %w", err)
		}
		winners = append(winners, *winner)
	}
	
	drawEntity.Winners = winners
	
	return drawEntity, nil
}

// Update implements the draw.DrawRepository interface
func (r *GormDrawRepository) Update(d *draw.Draw) error {
	model := toDrawModel(d)
	result := r.db.Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update draw: %w", result.Error)
	}
	
	return nil
}

// GetEligibilityStats implements the draw.DrawRepository interface
func (r *GormDrawRepository) GetEligibilityStats(date time.Time) (int, int, error) {
	// This would typically involve a complex query to the participants table
	// For demonstration purposes, we'll return placeholder values
	
	var totalEligibleMSISDNs int64
	var totalEntries int64
	
	// Format date to match database format (without time component)
	formattedDate := date.Format("2006-01-02")
	
	// Count distinct MSISDNs for the given date
	result := r.db.Table("participants").
		Where("DATE(recharge_date) <= ?", formattedDate).
		Distinct("msisdn").
		Count(&totalEligibleMSISDNs)
	
	if result.Error != nil {
		return 0, 0, fmt.Errorf("failed to count eligible MSISDNs: %w", result.Error)
	}
	
	// Sum points for the given date
	var err error
	err = r.db.Table("participants").
		Where("DATE(recharge_date) <= ?", formattedDate).
		Select("SUM(points)").
		Row().
		Scan(&totalEntries)
	
	if err != nil {
		return 0, 0, fmt.Errorf("failed to sum points: %w", err)
	}
	
	return int(totalEligibleMSISDNs), int(totalEntries), nil
}

// CreateWinner implements the draw.DrawRepository interface
func (r *GormDrawRepository) CreateWinner(w *draw.Winner) error {
	model := toWinnerModel(w)
	result := r.db.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create winner: %w", result.Error)
	}
	
	return nil
}

// GetWinnerByID implements the draw.DrawRepository interface
func (r *GormDrawRepository) GetWinnerByID(id uuid.UUID) (*draw.Winner, error) {
	var model WinnerModel
	result := r.db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, draw.NewDrawError(draw.ErrWinnerNotFound, "Winner not found", result.Error)
		}
		return nil, fmt.Errorf("failed to get winner: %w", result.Error)
	}
	
	winnerEntity, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert winner model to domain: %w", err)
	}
	
	return winnerEntity, nil
}

// UpdateWinner implements the draw.DrawRepository interface
func (r *GormDrawRepository) UpdateWinner(w *draw.Winner) error {
	model := toWinnerModel(w)
	result := r.db.Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update winner: %w", result.Error)
	}
	
	return nil
}

// GetRunnerUps implements the draw.DrawRepository interface
func (r *GormDrawRepository) GetRunnerUps(drawID uuid.UUID, prizeTierID uuid.UUID, limit int) ([]draw.Winner, error) {
	var models []WinnerModel
	result := r.db.Where("draw_id = ? AND prize_tier_id = ? AND is_runner_up = ?", 
		drawID.String(), prizeTierID.String(), true).
		Order("runner_up_rank ASC").
		Limit(limit).
		Find(&models)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get runner ups: %w", result.Error)
	}
	
	if len(models) == 0 {
		return nil, draw.NewDrawError(draw.ErrNoRunnerUpsAvailable, "No runner ups available", nil)
	}
	
	runnerUps := make([]draw.Winner, 0, len(models))
	for _, model := range models {
		runnerUp, err := model.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert runner up model to domain: %w", err)
		}
		runnerUps = append(runnerUps, *runnerUp)
	}
	
	return runnerUps, nil
}
