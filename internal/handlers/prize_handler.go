package handlers

import (
	"errors" // For errors.Is
	"net/http"

	"github.com/ArowuTest/GP-Backend-Promo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // For UUID parsing
	"gorm.io/gorm"         // For gorm.ErrRecordNotFound
)

// PrizeHandler handles operations related to prize structures
type PrizeHandler struct {
	db *gorm.DB
}

// NewPrizeHandler creates a new PrizeHandler with the provided database connection
func NewPrizeHandler(db *gorm.DB) *PrizeHandler {
	return &PrizeHandler{
		db: db,
	}
}

// CreatePrizeStructure godoc
// @Summary Create a new prize structure
// @Description Create a new prize structure with a name and a list of prizes.
// @Tags PrizeStructures
// @Accept json
// @Produce json
// @Param prize_structure body models.PrizeStructure true "Prize Structure object to be created. IDs for structure and prizes will be auto-generated."
// @Success 201 {object} models.PrizeStructure
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/prize-structures [post]
func (h *PrizeHandler) CreatePrizeStructure(c *gin.Context) {
	var newPrizeStructure models.PrizeStructure
	// Bind only specific fields for creation
	var input struct {
		Name        string                    `json:"name" binding:"required"`
		Description string                    `json:"description,omitempty"`
		IsActive    bool                      `json:"is_active"` // Defaults to true in model if not provided
		Prizes      []models.CreatePrizeRequest `json:"prizes" binding:"required,dive"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	newPrizeStructure.Name = input.Name
	newPrizeStructure.Description = input.Description
	newPrizeStructure.IsActive = input.IsActive // Or set a default if not in input

	for _, pReq := range input.Prizes {
		newPrizeStructure.Prizes = append(newPrizeStructure.Prizes, models.Prize{
			Name:      pReq.Name,
			Value:     pReq.Value,
			PrizeType: pReq.PrizeType,
			Quantity:  pReq.Quantity,
			Order:     pReq.Order,
			// ID and PrizeStructureID will be handled by GORM/BeforeCreate hooks
		})
	}

	// Extract CreatedByAdminID from JWT claims if available and set it
	adminIDStr, exists := c.Get("userID") // Assuming userID is set in JWTMiddleware
	if exists {
		adminUUID, err := uuid.Parse(adminIDStr.(string))
		if err == nil {
			newPrizeStructure.CreatedByAdminID = adminUUID
		}
	}

	if err := h.db.Create(&newPrizeStructure).Error; err != nil {
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
func (h *PrizeHandler) ListPrizeStructures(c *gin.Context) {
	var prizeStructures []models.PrizeStructure
	if err := h.db.Preload("Prizes").Find(&prizeStructures).Error; err != nil {
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
// @Param id path string true "Prize Structure ID (UUID)"
// @Success 200 {object} models.PrizeStructure
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Router /admin/prize-structures/{id} [get]
func (h *PrizeHandler) GetPrizeStructure(c *gin.Context) {
	idStr := c.Param("id")
	structureID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format. Expected UUID."})
		return
	}

	var prizeStructure models.PrizeStructure
	if err := h.db.Preload("Prizes").First(&prizeStructure, structureID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, prizeStructure)
}

// UpdatePrizeStructure godoc
// @Summary Update an existing prize structure
// @Description Update details of an existing prize structure. This will replace all existing prizes for the structure with the provided list.
// @Tags PrizeStructures
// @Accept json
// @Produce json
// @Param id path string true "Prize Structure ID (UUID)"
// @Param prize_structure body object{name=string,description=string,is_active=bool,prizes=[]models.CreatePrizeRequest} true "Prize Structure object with updated fields"
// @Success 200 {object} models.PrizeStructure
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/prize-structures/{id} [put]
func (h *PrizeHandler) UpdatePrizeStructure(c *gin.Context) {
	idStr := c.Param("id")
	structureID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format. Expected UUID."})
		return
	}

	var existingPrizeStructure models.PrizeStructure
	if err := h.db.First(&existingPrizeStructure, structureID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve prize structure for update: " + err.Error()})
		}
		return
	}

	var input struct {
		Name        string                    `json:"name" binding:"required"`
		Description string                    `json:"description,omitempty"`
		IsActive    bool                      `json:"is_active"`
		Prizes      []models.CreatePrizeRequest `json:"prizes" binding:"omitempty,dive"` // Prizes are optional for update
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	existingPrizeStructure.Name = input.Name
	existingPrizeStructure.Description = input.Description
	existingPrizeStructure.IsActive = input.IsActive

	// If prizes are provided in the update, replace them
	if input.Prizes != nil { // Check if prizes field was actually in the payload
		if err := tx.Where("prize_structure_id = ?", existingPrizeStructure.ID).Delete(&models.Prize{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old prizes: " + err.Error()})
			return
		}

		var newPrizes []models.Prize
		for _, pReq := range input.Prizes {
			newPrizes = append(newPrizes, models.Prize{
				ID:               uuid.Nil, // Ensure new UUID is generated by BeforeCreate
				PrizeStructureID: existingPrizeStructure.ID,
				Name:             pReq.Name,
				Value:            pReq.Value,
				PrizeType:        pReq.PrizeType,
				Quantity:         pReq.Quantity,
				Order:            pReq.Order,
			})
		}
		existingPrizeStructure.Prizes = newPrizes // GORM will handle creating these new prizes due to association and Save
	} // If input.Prizes is nil, existing prizes are not touched unless explicitly handled otherwise

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
	h.db.Preload("Prizes").First(&finalStructure, existingPrizeStructure.ID)

	c.JSON(http.StatusOK, finalStructure)
}

// DeletePrizeStructure godoc
// @Summary Delete a prize structure by ID
// @Description Delete a prize structure and its associated prizes by its ID.
// @Tags PrizeStructures
// @Produce json
// @Param id path string true "Prize Structure ID (UUID)"
// @Success 200 {object} gin.H{"message": string}
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /admin/prize-structures/{id} [delete]
func (h *PrizeHandler) DeletePrizeStructure(c *gin.Context) {
	idStr := c.Param("id")
	structureID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID format. Expected UUID."})
		return
	}

	// Check if structure exists
	var existingPrizeStructure models.PrizeStructure
	if err := h.db.First(&existingPrizeStructure, structureID).Error; err != nil {
	    if errors.Is(err, gorm.ErrRecordNotFound) {
	        c.JSON(http.StatusNotFound, gin.H{"error": "Prize structure not found"})
	    } else {
	        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking prize structure existence: " + err.Error()})
	    }
	    return
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// GORM will handle cascading delete for associated Prizes if constraints are set up correctly in the model
	// Or delete them manually if cascade is not reliable or not set:
	// if err := tx.Where("prize_structure_id = ?", structureID).Delete(&models.Prize{}).Error; err != nil {
	// 	tx.Rollback()
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated prizes: " + err.Error()})
	// 	return
	// }

	if err := tx.Select("Prizes").Delete(&existingPrizeStructure).Error; err != nil { // Select("Prizes") ensures associations are handled for deletion
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
