package handler

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/adapter"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
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
	prizeTiers := make([]prize.PrizeTier, len(req.Prizes))
	for i, p := range req.Prizes {
		prizeTiers[i] = prize.PrizeTier{
			ID:                uuid.New(),
			Rank:              p.Rank,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			CurrencyCode:      p.CurrencyCode,
			ValueNGN:          p.Value, // Default to same value if conversion not available
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		}
	}
	
	// Create prize structure
	prizeStructure := &prize.PrizeStructure{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		ValidFrom:   req.ValidFrom,
		ValidTo:     req.ValidTo,
		StartDate:   req.ValidFrom,
		EndDate:     req.ValidFrom, // Default to ValidFrom if ValidTo is nil
		CreatedBy:   userUUID,
		UpdatedBy:   userUUID,
		Prizes:      prizeTiers,
		CreatedAt:   c.Request.Context().Value("now").(time.Time),
		UpdatedAt:   c.Request.Context().Value("now").(time.Time),
	}
	
	if req.ValidTo != nil {
		prizeStructure.EndDate = *req.ValidTo
	}
	
	// Call service to create prize structure
	prizeStructureID, err := h.prizeService.CreatePrizeStructure(
		c.Request.Context(),
		prizeStructure.Name,
		prizeStructure.Description,
		prizeStructure.IsActive,
		prizeStructure.StartDate,
		prizeStructure.EndDate,
		userUUID,
		prizeTiers,
	)
	
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
	prizeTiers := make([]response.PrizeResponse, len(prizeStructure.Prizes))
	for i, p := range prizeStructure.Prizes {
		prizeTiers[i] = response.PrizeResponse{
			ID:                p.ID.String(),
			PrizeStructureID:  p.PrizeStructureID.String(),
			Rank:              p.Rank,
			Name:              p.Name,
			PrizeType:         "Cash", // Default type
			Description:       p.Description,
			Value:             p.Value,
			CurrencyCode:      p.CurrencyCode,
			ValueNGN:          p.ValueNGN,
			Quantity:          p.Quantity,
			Order:             p.Rank, // Map Rank to Order
			NumberOfRunnerUps: p.NumberOfRunnerUps,
			CreatedAt:         p.CreatedAt,
			UpdatedAt:         p.UpdatedAt,
		}
	}
	
	resp := response.PrizeStructureResponse{
		ID:          prizeStructure.ID.String(),
		Name:        prizeStructure.Name,
		Description: prizeStructure.Description,
		IsActive:    prizeStructure.IsActive,
		ValidFrom:   prizeStructure.ValidFrom.Format("2006-01-02"),
		ValidTo:     prizeStructure.ValidTo.Format("2006-01-02"),
		Prizes:      prizeTiers,
		CreatedAt:   prizeStructure.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   prizeStructure.UpdatedAt.Format("2006-01-02 15:04:05"),
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
	for i, ps := range prizeStructures {
		prizeTiers := make([]response.PrizeResponse, len(ps.Prizes))
		for j, p := range ps.Prizes {
			prizeTiers[j] = response.PrizeResponse{
				ID:                p.ID.String(),
				PrizeStructureID:  p.PrizeStructureID.String(),
				Rank:              p.Rank,
				Name:              p.Name,
				PrizeType:         "Cash", // Default type
				Description:       p.Description,
				Value:             p.Value,
				CurrencyCode:      p.CurrencyCode,
				ValueNGN:          p.ValueNGN,
				Quantity:          p.Quantity,
				Order:             p.Rank, // Map Rank to Order
				NumberOfRunnerUps: p.NumberOfRunnerUps,
				CreatedAt:         p.CreatedAt,
				UpdatedAt:         p.UpdatedAt,
			}
		}
		
		resp[i] = response.PrizeStructureResponse{
			ID:          ps.ID.String(),
			Name:        ps.Name,
			Description: ps.Description,
			IsActive:    ps.IsActive,
			ValidFrom:   ps.ValidFrom.Format("2006-01-02"),
			ValidTo:     ps.ValidTo.Format("2006-01-02"),
			Prizes:      prizeTiers,
			CreatedAt:   ps.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   ps.UpdatedAt.Format("2006-01-02 15:04:05"),
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
	prizeTiers := make([]prize.PrizeTier, len(req.Prizes))
	for i, p := range req.Prizes {
		var prizeID uuid.UUID
		if p.ID != "" {
			prizeID, err = uuid.Parse(p.ID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize ID"})
				return
			}
		} else {
			prizeID = uuid.New()
		}
		
		prizeTiers[i] = prize.PrizeTier{
			ID:                prizeID,
			PrizeStructureID:  prizeStructureID,
			Rank:              p.Rank,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			CurrencyCode:      p.CurrencyCode,
			ValueNGN:          p.Value, // Default to same value if conversion not available
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		}
	}
	
	// Call service to update prize structure
	err = h.prizeService.UpdatePrizeStructure(
		c.Request.Context(),
		prizeStructureID,
		req.Name,
		req.Description,
		req.IsActive,
		req.ValidFrom,
		req.ValidTo,
		userUUID,
		prizeTiers,
	)
	
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
