package handlers

import (
	"net/http"
	"strconv"

	"github.com/ArowuTest/GP-Backend-Promo/internal/config"
	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
)

// CreatePrizeStructure godoc
// @Summary Create a new prize structure
// @Description Create a new prize structure with a name and a list of prizes.
// @Tags PrizeStructures
// @Accept json
// @Produce json
// @Param prize_structure body models.PrizeStructure true "Prize Structure object to be created"
// @Success 201 {object} models.PrizeStructure
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/prize-structures [post]
func CreatePrizeStructure(c *gin.Context) {
	var newPrizeStructure models.PrizeStructure
	if err := c.ShouldBindJSON(&newPrizeStructure); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Basic validation
	if newPrizeStructure.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prize structure name cannot be empty"})
		return
	}
	if len(newPrizeStructure.Prizes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prize structure must contain at least one prize"})
		return
	}

	for _, prize := range newPrizeStructure.Prizes {
		if prize.Name == "" || prize.Value == "" || prize.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize details: name, value, and quantity are required and quantity must be positive"})
			return
		}
	}

	if err := db.DB.Create(&newPrizeStructure).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prize structure: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newPrizeStructure)
}

// ListPrizeStructures godoc
// @Summary List all prize structures
// @Description Get a list of all prize structures, including their prizes.
// @Tags PrizeStructures
// @Produce json
// @Success 200 {array} models.PrizeStructure
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/prize-structures [get]
func ListPrizeStructures(c *gin.Context) {
	var prizeStructures []models.PrizeStructure
	// Preload Prizes to get associated prizes
	if err := db.DB.Preload("Prizes").Find(&prizeStructures).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structures: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, prizeStructures)
}

// GetPrizeStructure godoc
// @Summary Get a prize structure by ID
// @Description Get details of a specific prize structure by its ID, including its prizes.
// @Tags PrizeStructures
// @Produce json
// @Param id path int true "Prize Structure ID"
// @Success 200 {object} models.PrizeStructure
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Router /admin/prize-structures/{id} [get]
func GetPrizeStructure(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var prizeStructure models.PrizeStructure
	// Preload Prizes to get associated prizes
	if err := db.DB.Preload("Prizes").First(&prizeStructure, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
		return
	}
	c.JSON(http.StatusOK, prizeStructure)
}

// UpdatePrizeStructure godoc
// @Summary Update an existing prize structure
// @Description Update details of an existing prize structure, including its name, active status, and prizes. 
// @Description Note: This will replace all existing prizes for the structure with the provided list.
// @Tags PrizeStructures
// @Accept json
// @Produce json
// @Param id path int true "Prize Structure ID"
// @Param prize_structure body models.PrizeStructure true "Prize Structure object with updated fields"
// @Success 200 {object} models.PrizeStructure
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/prize-structures/{id} [put]
func UpdatePrizeStructure(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	var existingPrizeStructure models.PrizeStructure
	if err := db.DB.Preload("Prizes").First(&existingPrizeStructure, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
		return
	}

	var updatedInfo models.PrizeStructure
	if err := c.ShouldBindJSON(&updatedInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Basic validation for updated info
	if updatedInfo.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prize structure name cannot be empty"})
		return
	}
	// If prizes are provided in the update, validate them
	if len(updatedInfo.Prizes) > 0 {
		for _, prize := range updatedInfo.Prizes {
			if prize.Name == "" || prize.Value == "" || prize.Quantity <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize details: name, value, and quantity are required and quantity must be positive"})
				return
			}
		}
	} else {
        // If no prizes are provided in the update payload, it implies removing all prizes, which might be undesirable.
        // Depending on business logic, you might want to prevent this or handle it specifically.
        // For now, we'll allow it, but it means the structure will have no prizes.
    }

	tx := db.DB.Begin()

	// Update prize structure fields
	existingPrizeStructure.Name = updatedInfo.Name
	existingPrizeStructure.IsActive = updatedInfo.IsActive

	// Delete old prizes associated with this structure
	if err := tx.Where("prize_structure_id = ?", existingPrizeStructure.ID).Delete(&models.Prize{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old prizes: " + err.Error()})
		return
	}

	// Add new prizes (if any)
	if len(updatedInfo.Prizes) > 0 {
		for i := range updatedInfo.Prizes {
			updatedInfo.Prizes[i].PrizeStructureID = existingPrizeStructure.ID // Ensure association
            updatedInfo.Prizes[i].ID = 0 // Ensure GORM creates new prize records
		}
		existingPrizeStructure.Prizes = updatedInfo.Prizes
		if err := tx.Create(&existingPrizeStructure.Prizes).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new prizes: " + err.Error()})
			return
		}
	} else {
        existingPrizeStructure.Prizes = []models.Prize{}
    }


	if err := tx.Save(&existingPrizeStructure).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure: " + err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error: " + err.Error()})
        return
    }

    // Refetch to ensure all associations are correctly loaded for the response
    var finalStructure models.PrizeStructure
    db.DB.Preload("Prizes").First(&finalStructure, existingPrizeStructure.ID)

	c.JSON(http.StatusOK, finalStructure)
}

// DeletePrizeStructure godoc
// @Summary Delete a prize structure by ID
// @Description Delete a prize structure and its associated prizes by its ID.
// @Tags PrizeStructures
// @Produce json
// @Param id path int true "Prize Structure ID"
// @Success 200 {object} gin.H{"message": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/prize-structures/{id} [delete]
func DeletePrizeStructure(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format"})
		return
	}

	tx := db.DB.Begin()

	// First, delete associated prizes
	if err := tx.Where("prize_structure_id = ?", uint(id)).Delete(&models.Prize{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated prizes: " + err.Error()})
		return
	}

	// Then, delete the prize structure itself
	if err := tx.Delete(&models.PrizeStructure{}, uint(id)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prize structure: " + err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error: " + err.Error()})
        return
    }

	c.JSON(http.StatusOK, gin.H{"message": "Prize structure and associated prizes deleted successfully"})
}


