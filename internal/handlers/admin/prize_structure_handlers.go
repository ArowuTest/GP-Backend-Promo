package admin

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause" // Correct import for GORM clauses
)

// CreatePrizeStructureRequest defines the structure for creating a prize structure
type CreatePrizeStructureRequest struct {
	Name        string                      `json:"name" binding:"required"`
	Description string                      `json:"description,omitempty"`
	IsActive    bool                        `json:"is_active"` // Default is true in model
	ValidFrom   *time.Time                  `json:"valid_from,omitempty"`
	ValidTo     *time.Time                  `json:"valid_to,omitempty"`
	Prizes      []models.CreatePrizeRequest `json:"prizes" binding:"required,dive"`
}

// CreatePrizeStructure handles the creation of a new prize structure
func CreatePrizeStructure(c *gin.Context) {
	var req CreatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	if req.ValidFrom != nil && req.ValidTo != nil && req.ValidTo.Before(*req.ValidFrom) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidTo cannot be before ValidFrom"})
		return
	}

	adminIDClaim, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token"})
		return
	}
	adminIDStr, ok := adminIDClaim.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin user ID in token is not a string"})
		return
	}
	parsedAdminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin user ID format in token"})
		return
	}

	var createdPrizeStructure models.PrizeStructure

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		prizeStructure := models.PrizeStructure{
			Name:             req.Name,
			Description:      req.Description,
			IsActive:         req.IsActive, // Model has default true, this will override if false
			ValidFrom:        req.ValidFrom,
			ValidTo:          req.ValidTo,
			CreatedByAdminID: parsedAdminID,
		}

		if err := tx.Create(&prizeStructure).Error; err != nil {
			return fmt.Errorf("failed to create prize structure: %w", err)
		}

		for _, prizeReq := range req.Prizes {
			prize := models.Prize{
				PrizeStructureID: prizeStructure.ID,
				Name:             prizeReq.Name,
				Value:            prizeReq.Value,
				PrizeType:        prizeReq.PrizeType,
				Quantity:         prizeReq.Quantity,
				Order:            prizeReq.Order,
			}
			if err := tx.Create(&prize).Error; err != nil {
				return fmt.Errorf("failed to create prize tier %s: %w", prizeReq.Name, err)
			}
		}

		if err := tx.Preload("Prizes").First(&prizeStructure, prizeStructure.ID).Error; err != nil {
			return fmt.Errorf("failed to reload prize structure with prizes: %w", err)
		}
		createdPrizeStructure = prizeStructure
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdPrizeStructure)
}

// ListPrizeStructures handles listing all prize structures
func ListPrizeStructures(c *gin.Context) {
	var prizeStructures []models.PrizeStructure
	result := config.DB.Preload("Prizes").Order("created_at desc").Find(&prizeStructures)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structures: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, prizeStructures)
}

// GetPrizeStructure handles retrieving a single prize structure by ID
func GetPrizeStructure(c *gin.Context) {
	structureIDStr := c.Param("id")
	structureID, err := uuid.Parse(structureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var prizeStructure models.PrizeStructure
	result := config.DB.Preload("Prizes").First(&prizeStructure, "id = ?", structureID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, prizeStructure)
}

// UpdatePrizeStructureRequest defines the structure for updating a prize structure
type UpdatePrizeStructureRequest struct {
	Name        *string                     `json:"name,omitempty"`
	Description *string                     `json:"description,omitempty"`
	IsActive    *bool                       `json:"is_active,omitempty"`
	ValidFrom   *time.Time                  `json:"valid_from,omitempty"`
	ValidTo     **time.Time                 `json:"valid_to,omitempty"` // Pointer to pointer for explicit null
	Prizes      *[]models.CreatePrizeRequest `json:"prizes,omitempty"`
}

// UpdatePrizeStructure handles updating an existing prize structure
func UpdatePrizeStructure(c *gin.Context) {
	structureIDStr := c.Param("id")
	structureID, err := uuid.Parse(structureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var req UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	var updatedPrizeStructure models.PrizeStructure

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		var prizeStructure models.PrizeStructure
		if err := tx.Preload("Prizes").First(&prizeStructure, "id = ?", structureID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("prize structure not found")
			}
			return fmt.Errorf("failed to find prize structure: %w", err)
		}

		updates := make(map[string]interface{})
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}
		if req.ValidFrom != nil {
			updates["valid_from"] = *req.ValidFrom
		}
		if req.ValidTo != nil {
			if *req.ValidTo == nil {
				updates["valid_to"] = clause.Expr{SQL: "NULL"} // Correct way to set NULL with GORM
			} else {
				updates["valid_to"] = **req.ValidTo
			}
		}

		currentValidFrom := prizeStructure.ValidFrom
		if val, ok := updates["valid_from"].(time.Time); ok {
			currentValidFrom = &val
		}

		var currentValidTo *time.Time = prizeStructure.ValidTo
		if val, ok := updates["valid_to"].(time.Time); ok {
			currentValidTo = &val
		} else if _, ok := updates["valid_to"].(clause.Expr); ok {
			currentValidTo = nil // Being set to NULL
		}

		if currentValidFrom != nil && currentValidTo != nil && currentValidTo.Before(*currentValidFrom) {
			return errors.New("ValidTo cannot be before ValidFrom")
		}

		if len(updates) > 0 {
			if err := tx.Model(&prizeStructure).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update prize structure details: %w", err)
			}
		}

		if req.Prizes != nil {
			if err := tx.Where("prize_structure_id = ?", prizeStructure.ID).Delete(&models.Prize{}).Error; err != nil {
				return fmt.Errorf("failed to delete existing prizes: %w", err)
			}
			for _, prizeReq := range *req.Prizes {
				newPrize := models.Prize{
					PrizeStructureID: prizeStructure.ID,
					Name:             prizeReq.Name,
					Value:            prizeReq.Value,
					PrizeType:        prizeReq.PrizeType,
					Quantity:         prizeReq.Quantity,
					Order:            prizeReq.Order,
				}
				if err := tx.Create(&newPrize).Error; err != nil {
					return fmt.Errorf("failed to create new prize tier %s: %w", prizeReq.Name, err)
				}
			}
		}

		if err := tx.Preload("Prizes").First(&updatedPrizeStructure, "id = ?", structureID).Error; err != nil {
			return fmt.Errorf("failed to reload updated prize structure: %w", err)
		}
		return nil
	})

	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) || txErr.Error() == "prize structure not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedPrizeStructure)
}

// DeletePrizeStructure handles deleting a prize structure (soft delete)
func DeletePrizeStructure(c *gin.Context) {
	structureIDStr := c.Param("id")
	structureID, err := uuid.Parse(structureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var drawCount int64
	config.DB.Model(&models.Draw{}).Where("prize_structure_id = ?", structureID).Count(&drawCount)
	if drawCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete prize structure: It is associated with existing draws."})
		return
	}

	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("prize_structure_id = ?", structureID).Delete(&models.Prize{}).Error; err != nil {
			return fmt.Errorf("failed to soft delete prizes: %w", err)
		}
		if err := tx.Delete(&models.PrizeStructure{}, "id = ?", structureID).Error; err != nil {
			return fmt.Errorf("failed to soft delete prize structure: %w", err)
		}
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prize structure: " + txErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prize structure and its prizes deleted successfully"})
}

