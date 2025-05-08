package admin

import (
	"fmt" // Added missing fmt import
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mynumba_don_win_draw_system_backend_internal_config "mynumba-don-win-draw-system/backend/internal/config"
	mynumba_don_win_draw_system_backend_internal_models "mynumba-don-win-draw-system/backend/internal/models"
	"gorm.io/gorm"
)

// CreatePrizeStructureRequest defines the structure for creating a prize structure
type CreatePrizeStructureRequest struct {
	Name               string                                                                    `json:"name" binding:"required"`
	DayType            mynumba_don_win_draw_system_backend_internal_models.DayType               `json:"day_type" binding:"required"`
	IsActive           bool                                                                      `json:"isActive"` // Default to false, activate explicitly
	EffectiveStartDate *time.Time                                                                `json:"effective_start_date,omitempty"`
	EffectiveEndDate   *time.Time                                                                `json:"effective_end_date,omitempty"`
	PrizeTiers         []mynumba_don_win_draw_system_backend_internal_models.CreatePrizeTierRequest `json:"prize_tiers" binding:"required,dive"`
}

// CreatePrizeStructure handles the creation of a new prize structure
func CreatePrizeStructure(c *gin.Context) {
	var req CreatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Basic validation for dates
	if req.EffectiveStartDate != nil && req.EffectiveEndDate != nil && req.EffectiveEndDate.Before(*req.EffectiveStartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "EffectiveEndDate cannot be before EffectiveStartDate"})
		return
	}

	// Create PrizeStructure and its Tiers within a transaction
	txErr := mynumba_don_win_draw_system_backend_internal_config.DB.Transaction(func(tx *gorm.DB) error {
		prizeStructure := mynumba_don_win_draw_system_backend_internal_models.PrizeStructure{
			Name:               req.Name,
			DayType:            req.DayType,
			IsActive:           req.IsActive, 
			EffectiveStartDate: req.EffectiveStartDate,
			EffectiveEndDate:   req.EffectiveEndDate,
			// CreatedByAdminID should be set from JWT context
		}

		// Get AdminID from context (set by JWTMiddleware)
		adminIDInterface, exists := c.Get("userID")
		if !exists {
			// This should ideally not happen if middleware is correctly applied
			return fmt.Errorf("admin ID not found in context")
		}
		adminIDStr, ok := adminIDInterface.(string)
		if !ok {
			return fmt.Errorf("admin ID in context is not a string")
		}
		parsedAdminID, err := uuid.Parse(adminIDStr)
		if err != nil {
			return fmt.Errorf("failed to parse admin ID from context: %v", err)
		}
		prizeStructure.CreatedByAdminID = parsedAdminID

		if err := tx.Create(&prizeStructure).Error; err != nil {
			return err
		}

		for _, tierReq := range req.PrizeTiers {
			tier := mynumba_don_win_draw_system_backend_internal_models.PrizeTier{
				PrizeStructureID: prizeStructure.ID,
				TierName:         tierReq.TierName, // Corrected from Name
				TierDescription:  tierReq.TierDescription, // Corrected from PrizeType
				PrizeAmount:      tierReq.PrizeAmount, // Corrected from ValueNGN
				WinnerCount:      tierReq.WinnerCount,
				SortOrder:        tierReq.SortOrder, // Corrected from Order
			}
			if err := tx.Create(&tier).Error; err != nil {
				return err // Rollback transaction
			}
		}
		// Load the created prize structure with its tiers for the response
		if err := tx.Preload("PrizeTiers").First(&prizeStructure, prizeStructure.ID).Error; err != nil {
		    return err
		}
		c.Set("createdPrizeStructure", prizeStructure) // Pass to outer scope
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prize structure: " + txErr.Error()})
		return
	}

	createdPrizeStructure, _ := c.Get("createdPrizeStructure")
	c.JSON(http.StatusCreated, createdPrizeStructure)
}

// ListPrizeStructures handles listing all prize structures
func ListPrizeStructures(c *gin.Context) {
	var prizeStructures []mynumba_don_win_draw_system_backend_internal_models.PrizeStructure
	// Add pagination, filtering by active status, date range etc. later
	result := mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeTiers").Order("created_at desc").Find(&prizeStructures)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structures: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, prizeStructures)
}

// GetPrizeStructure handles retrieving a single prize structure by ID
func GetPrizeStructure(c *gin.Context) {
	structureID := c.Param("id")
	parsedStructureID, err := uuid.Parse(structureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var prizeStructure mynumba_don_win_draw_system_backend_internal_models.PrizeStructure
	result := mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeTiers").Where("id = ?", parsedStructureID).First(&prizeStructure)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, prizeStructure)
}


type UpdatePrizeStructureRequest struct {
	Name               *string                                                                   `json:"name,omitempty"`
	DayType            *mynumba_don_win_draw_system_backend_internal_models.DayType              `json:"day_type,omitempty"`
	IsActive           *bool                                                                     `json:"isActive,omitempty"`
	EffectiveStartDate *time.Time                                                                `json:"effective_start_date,omitempty"`
	EffectiveEndDate   *time.Time                                                                `json:"effective_end_date,omitempty"`
	// prizeTiers: For now, not handling direct update of tiers via this endpoint to keep it simpler.
}

// UpdatePrizeStructure handles updating an existing prize structure
func UpdatePrizeStructure(c *gin.Context) {
	structureID := c.Param("id")
	parsedStructureID, err := uuid.Parse(structureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var req UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	var prizeStructure mynumba_don_win_draw_system_backend_internal_models.PrizeStructure
	if mynumba_don_win_draw_system_backend_internal_config.DB.Where("id = ?", parsedStructureID).First(&prizeStructure).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.DayType != nil {
		updates["day_type"] = *req.DayType
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.EffectiveStartDate != nil {
		updates["effective_start_date"] = *req.EffectiveStartDate
	}
	if req.EffectiveEndDate != nil { 
		updates["effective_end_date"] = req.EffectiveEndDate
	} // Add logic if you need to explicitly set EffectiveEndDate to NULL

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update fields provided"})
		return
	}

    currentEffectiveStartDate := prizeStructure.EffectiveStartDate
    if val, ok := updates["effective_start_date"].(*time.Time); ok {
        currentEffectiveStartDate = val
    }
    currentEffectiveEndDate := prizeStructure.EffectiveEndDate
    if val, ok := updates["effective_end_date"].(*time.Time); ok {
        currentEffectiveEndDate = val
    }

    if currentEffectiveStartDate != nil && currentEffectiveEndDate != nil && currentEffectiveEndDate.Before(*currentEffectiveStartDate) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "EffectiveEndDate cannot be before EffectiveStartDate"})
        return
    }

	result := mynumba_don_win_draw_system_backend_internal_config.DB.Model(&prizeStructure).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure: " + result.Error.Error()})
		return
	}

	// Refetch to get updated data with tiers
	mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeTiers").First(&prizeStructure, "id = ?", parsedStructureID)
	c.JSON(http.StatusOK, prizeStructure)
}

// DeletePrizeStructure handles deleting a prize structure (soft delete is preferred)
func DeletePrizeStructure(c *gin.Context) {
	structureID := c.Param("id")
	parsedStructureID, err := uuid.Parse(structureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	// Check if the prize structure is associated with any draws. If so, prevent deletion or handle accordingly.
	var drawCount int64
    mynumba_don_win_draw_system_backend_internal_config.DB.Model(&mynumba_don_win_draw_system_backend_internal_models.Draw{}).Where("prize_structure_id = ?", parsedStructureID).Count(&drawCount)
    if drawCount > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete prize structure: It is associated with existing draws."})
        return
    }

	// Soft delete would be: DB.Model(&models.PrizeStructure{}).Where("id = ?", parsedStructureID).Update("deleted_at", time.Now())
	// For hard delete with cascading tier deletion:
	txErr := mynumba_don_win_draw_system_backend_internal_config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("prize_structure_id = ?", parsedStructureID).Delete(&mynumba_don_win_draw_system_backend_internal_models.PrizeTier{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&mynumba_don_win_draw_system_backend_internal_models.PrizeStructure{}, "id = ?", parsedStructureID).Error; err != nil {
			return err
		}
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prize structure: " + txErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prize structure and its tiers deleted successfully"})
}


type ActivatePrizeStructureRequest struct {
    IsActive bool `json:"isActive"`
}

func ActivatePrizeStructure(c *gin.Context) {
    structureID := c.Param("id")
    parsedStructureID, err := uuid.Parse(structureID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
        return
    }

    var req ActivatePrizeStructureRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
        return
    }

    var prizeStructure mynumba_don_win_draw_system_backend_internal_models.PrizeStructure
    if mynumba_don_win_draw_system_backend_internal_config.DB.Where("id = ?", parsedStructureID).First(&prizeStructure).Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
        return
    }

    result := mynumba_don_win_draw_system_backend_internal_config.DB.Model(&prizeStructure).Update("is_active", req.IsActive)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure status: " + result.Error.Error()})
        return
    }

    mynumba_don_win_draw_system_backend_internal_config.DB.Preload("PrizeTiers").First(&prizeStructure, "id = ?", parsedStructureID)
    c.JSON(http.StatusOK, prizeStructure)
}

