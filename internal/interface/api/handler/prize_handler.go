package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/adapter"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// PrizeHandler handles HTTP requests related to prize structures
type PrizeHandler struct {
	prizeService adapter.PrizeServiceAdapter
}

// NewPrizeHandler creates a new PrizeHandler
func NewPrizeHandler(prizeService adapter.PrizeServiceAdapter) *PrizeHandler {
	return &PrizeHandler{
		prizeService: prizeService,
	}
}

// CreatePrizeStructure handles the creation of a new prize structure
func (h *PrizeHandler) CreatePrizeStructure(c *gin.Context) {
	var req response.CreatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert user ID to UUID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create prize tiers from request
	prizeTiers := make([]struct {
		Rank              int
		Name              string
		Description       string
		Value             float64
		CurrencyCode      string // Added currency code field
		Quantity          int
		NumberOfRunnerUps int
	}, len(req.Prizes))

	for i, prize := range req.Prizes {
		prizeTiers[i] = struct {
			Rank              int
			Name              string
			Description       string
			Value             float64
			CurrencyCode      string // Added currency code field
			Quantity          int
			NumberOfRunnerUps int
		}{
			Rank:              prize.Rank,
			Name:              prize.Name,
			Description:       prize.Description,
			Value:             prize.Value,
			CurrencyCode:      prize.CurrencyCode, // Added currency code field
			Quantity:          prize.Quantity,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		}
	}

	// Call service to create prize structure
	prizeStructureID, err := h.prizeService.CreatePrizeStructure(c.Request.Context(), req.Name, req.Description, req.IsActive, req.ValidFrom, req.ValidTo, userUUID, userUUID, prizeTiers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": prizeStructureID})
}

// GetPrizeStructure handles the retrieval of a prize structure by ID
func (h *PrizeHandler) GetPrizeStructure(c *gin.Context) {
	// Get prize structure ID from URL
	id := c.Param("id")
	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID"})
		return
	}

	// Call service to get prize structure
	prizeStructure, err := h.prizeService.GetPrizeStructure(c.Request.Context(), prizeStructureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTO
	prizeTiers := make([]response.PrizeTierResponse, len(prizeStructure.Prizes))
	for i, prize := range prizeStructure.Prizes {
		prizeTiers[i] = response.PrizeTierResponse{
			ID:                prize.ID.String(),
			PrizeStructureID:  prize.PrizeStructureID.String(),
			Rank:              prize.Rank,
			Name:              prize.Name,
			Description:       prize.Description,
			Value:             prize.Value,
			CurrencyCode:      prize.CurrencyCode, // Added currency code field
			ValueNGN:          prize.ValueNGN,
			Quantity:          prize.Quantity,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
			CreatedAt:         prize.CreatedAt,
			UpdatedAt:         prize.UpdatedAt,
		}
	}

	resp := response.PrizeStructureResponse{
		ID:          prizeStructure.ID.String(),
		Name:        prizeStructure.Name,
		Description: prizeStructure.Description,
		IsActive:    prizeStructure.IsActive,
		ValidFrom:   prizeStructure.ValidFrom,
		ValidTo:     prizeStructure.ValidTo,
		Prizes:      prizeTiers,
		CreatedAt:   prizeStructure.CreatedAt,
		UpdatedAt:   prizeStructure.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

// ListPrizeStructures handles the retrieval of all prize structures
func (h *PrizeHandler) ListPrizeStructures(c *gin.Context) {
	// Get pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Call service to list prize structures
	prizeStructures, total, err := h.prizeService.ListPrizeStructures(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTOs
	resp := make([]response.PrizeStructureResponse, len(prizeStructures))
	for i, prizeStructure := range prizeStructures {
		prizeTiers := make([]response.PrizeTierResponse, len(prizeStructure.Prizes))
		for j, prize := range prizeStructure.Prizes {
			prizeTiers[j] = response.PrizeTierResponse{
				ID:                prize.ID.String(),
				PrizeStructureID:  prize.PrizeStructureID.String(),
				Rank:              prize.Rank,
				Name:              prize.Name,
				Description:       prize.Description,
				Value:             prize.Value,
				CurrencyCode:      prize.CurrencyCode, // Added currency code field
				ValueNGN:          prize.ValueNGN,
				Quantity:          prize.Quantity,
				NumberOfRunnerUps: prize.NumberOfRunnerUps,
				CreatedAt:         prize.CreatedAt,
				UpdatedAt:         prize.UpdatedAt,
			}
		}

		resp[i] = response.PrizeStructureResponse{
			ID:          prizeStructure.ID.String(),
			Name:        prizeStructure.Name,
			Description: prizeStructure.Description,
			IsActive:    prizeStructure.IsActive,
			ValidFrom:   prizeStructure.ValidFrom,
			ValidTo:     prizeStructure.ValidTo,
			Prizes:      prizeTiers,
			CreatedAt:   prizeStructure.CreatedAt,
			UpdatedAt:   prizeStructure.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  resp,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// UpdatePrizeStructure handles the update of a prize structure
func (h *PrizeHandler) UpdatePrizeStructure(c *gin.Context) {
	// Get prize structure ID from URL
	id := c.Param("id")
	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID"})
		return
	}

	var req response.UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert user ID to UUID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create prize tiers from request
	prizeTiers := make([]struct {
		ID                uuid.UUID
		Rank              int
		Name              string
		Description       string
		Value             float64
		CurrencyCode      string // Added currency code field
		Quantity          int
		NumberOfRunnerUps int
	}, len(req.Prizes))

	for i, prize := range req.Prizes {
		var prizeID uuid.UUID
		if prize.ID != "" {
			prizeID, err = uuid.Parse(prize.ID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize ID"})
				return
			}
		} else {
			prizeID = uuid.New()
		}

		prizeTiers[i] = struct {
			ID                uuid.UUID
			Rank              int
			Name              string
			Description       string
			Value             float64
			CurrencyCode      string // Added currency code field
			Quantity          int
			NumberOfRunnerUps int
		}{
			ID:                prizeID,
			Rank:              prize.Rank,
			Name:              prize.Name,
			Description:       prize.Description,
			Value:             prize.Value,
			CurrencyCode:      prize.CurrencyCode, // Added currency code field
			Quantity:          prize.Quantity,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		}
	}

	// Call service to update prize structure
	err = h.prizeService.UpdatePrizeStructure(c.Request.Context(), prizeStructureID, req.Name, req.Description, req.IsActive, req.ValidFrom, req.ValidTo, userUUID, prizeTiers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prize structure updated successfully"})
}

// DeletePrizeStructure handles the deletion of a prize structure
func (h *PrizeHandler) DeletePrizeStructure(c *gin.Context) {
	// Get prize structure ID from URL
	id := c.Param("id")
	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize structure ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert user ID to UUID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Call service to delete prize structure
	err = h.prizeService.DeletePrizeStructure(c.Request.Context(), prizeStructureID, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prize structure deleted successfully"})
}
