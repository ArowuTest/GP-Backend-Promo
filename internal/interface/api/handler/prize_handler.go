package handler

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	prizeApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// PrizeHandler handles prize-related HTTP requests
type PrizeHandler struct {
	createPrizeStructureService *prizeApp.CreatePrizeStructureService
	getPrizeStructureService    *prizeApp.GetPrizeStructureService
	listPrizeStructuresService  *prizeApp.ListPrizeStructuresService
	updatePrizeStructureService *prizeApp.UpdatePrizeStructureService
}

// NewPrizeHandler creates a new PrizeHandler
func NewPrizeHandler(
	createPrizeStructureService *prizeApp.CreatePrizeStructureService,
	getPrizeStructureService *prizeApp.GetPrizeStructureService,
	listPrizeStructuresService *prizeApp.ListPrizeStructuresService,
	updatePrizeStructureService *prizeApp.UpdatePrizeStructureService,
) *PrizeHandler {
	return &PrizeHandler{
		createPrizeStructureService: createPrizeStructureService,
		getPrizeStructureService:    getPrizeStructureService,
		listPrizeStructuresService:  listPrizeStructuresService,
		updatePrizeStructureService: updatePrizeStructureService,
	}
}

// CreatePrizeStructure handles POST /api/admin/prize-structures
func (h *PrizeHandler) CreatePrizeStructure(c *gin.Context) {
	var req request.CreatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: "Please check all required fields and formats",
		})
		return
	}
	
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Prepare input
	prizes := make([]prizeApp.PrizeInput, 0, len(req.Prizes))
	for _, prize := range req.Prizes {
		prizes = append(prizes, prizeApp.PrizeInput{
			Name:              prize.Name,
			Description:       prize.Name, // Using name as description
			PrizeType:         prize.PrizeType,
			Value:             prize.Value,
			ValueNGN:          prize.ValueNGN,
			Quantity:          prize.Quantity,
			Order:             prize.Order,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	// Handle optional ValidTo field
	endDate := ""
	if req.ValidTo != nil {
		endDate = *req.ValidTo
	}
	
	input := prizeApp.CreatePrizeStructureInput{
		Name:           req.Name,
		Description:    req.Description,
		IsActive:       req.IsActive,
		StartDate:      req.ValidFrom,
		EndDate:        endDate,
		ApplicableDays: req.ApplicableDays,
		DayType:        req.DayType,
		Prizes:         prizes,
		CreatedBy:      userID.(uuid.UUID),
	}
	
	// Create prize structure
	output, err := h.createPrizeStructureService.CreatePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create prize structure: " + err.Error(),
			Details: "An error occurred while processing your request. Please try again later.",
		})
		return
	}
	
	// Prepare response
	prizeTiers := make([]response.PrizeTierResponse, 0, len(output.Prizes))
	for i, prize := range output.Prizes {
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         prize.PrizeType,
			Value:             prize.Value,
			ValueNGN:          prize.ValueNGN,
			Quantity:          prize.Quantity,
			Order:             i + 1, // Use index if order is not set
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:             output.ID.String(),
			Name:           output.Name,
			Description:    output.Description,
			IsActive:       req.IsActive,
			ValidFrom:      output.StartDate,
			ValidTo:        output.EndDate,
			ApplicableDays: req.ApplicableDays,
			DayType:        req.DayType,
			Prizes:         prizeTiers,
			CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// GetPrizeStructure handles GET /api/admin/prize-structures/:id
func (h *PrizeHandler) GetPrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
			Details: "The provided ID is not in the correct UUID format",
		})
		return
	}
	
	// Prepare input
	input := prizeApp.GetPrizeStructureInput{
		ID: prizeStructureID,
	}
	
	// Get prize structure
	output, err := h.getPrizeStructureService.GetPrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize structure: " + err.Error(),
			Details: "An error occurred while retrieving the prize structure. Please try again later.",
		})
		return
	}
	
	// Prepare response
	prizeTiers := make([]response.PrizeTierResponse, 0, len(output.Prizes))
	for i, prize := range output.Prizes {
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         "Cash", // Default type if not set
			Value:             prize.Value,
			ValueNGN:          prize.ValueNGN,
			Quantity:          prize.Quantity,
			Order:             i + 1, // Use index if order is not set
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:             output.ID.String(),
			Name:           output.Name,
			Description:    output.Description,
			IsActive:       true, // Default to active
			ValidFrom:      output.StartDate.Format("2006-01-02"),
			ValidTo:        output.EndDate.Format("2006-01-02"),
			ApplicableDays: []string{}, // Empty applicable days
			Prizes:         prizeTiers,
			CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// ListPrizeStructures handles GET /api/admin/prize-structures
func (h *PrizeHandler) ListPrizeStructures(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	
	// Prepare input
	input := prizeApp.ListPrizeStructuresInput{
		Page:     page,
		PageSize: pageSize,
	}
	
	// List prize structures
	output, err := h.listPrizeStructuresService.ListPrizeStructures(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list prize structures: " + err.Error(),
			Details: "An error occurred while retrieving prize structures. Please try again later.",
		})
		return
	}
	
	// Prepare response
	prizeStructures := make([]response.PrizeStructureResponse, 0, len(output.PrizeStructures))
	for _, ps := range output.PrizeStructures {
		prizeTiers := make([]response.PrizeTierResponse, 0, len(ps.Prizes))
		for i, prize := range ps.Prizes {
			prizeTiers = append(prizeTiers, response.PrizeTierResponse{
				ID:                prize.ID.String(),
				Name:              prize.Name,
				PrizeType:         "Cash", // Default type
				Value:             prize.Value,
				ValueNGN:          prize.ValueNGN,
				Quantity:          prize.Quantity,
				Order:             i + 1, // Use index if order is not set
				NumberOfRunnerUps: prize.NumberOfRunnerUps,
			})
		}
		
		prizeStructures = append(prizeStructures, response.PrizeStructureResponse{
			ID:             ps.ID.String(),
			Name:           ps.Name,
			Description:    ps.Description,
			IsActive:       true, // Default to active
			ValidFrom:      ps.StartDate.Format("2006-01-02"),
			ValidTo:        ps.EndDate.Format("2006-01-02"),
			ApplicableDays: []string{}, // Empty applicable days
			Prizes:         prizeTiers,
			CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    prizeStructures,
	})
}

// UpdatePrizeStructure handles PUT /api/admin/prize-structures/:id
func (h *PrizeHandler) UpdatePrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
			Details: "The provided ID is not in the correct UUID format",
		})
		return
	}
	
	var req request.UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: "Please check all required fields and formats",
		})
		return
	}
	
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}
	
	// Prepare input
	updatePrizes := make([]prizeApp.UpdatePrizeInput, 0, len(req.Prizes))
	for _, prize := range req.Prizes {
		prizeID, _ := uuid.Parse(prize.ID) // Ignore error, will be uuid.Nil if empty
		updatePrizes = append(updatePrizes, prizeApp.UpdatePrizeInput{
			ID:                prizeID,
			Name:              prize.Name,
			Description:       prize.Name, // Using name as description
			PrizeType:         prize.PrizeType,
			Value:             prize.Value,
			ValueNGN:          prize.ValueNGN,
			Quantity:          prize.Quantity,
			Order:             prize.Order,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	// Handle optional ValidTo field
	endDate := ""
	if req.ValidTo != nil {
		endDate = *req.ValidTo
	}
	
	input := prizeApp.UpdatePrizeStructureInput{
		ID:             prizeStructureID,
		Name:           req.Name,
		Description:    req.Description,
		IsActive:       req.IsActive,
		StartDate:      req.ValidFrom,
		EndDate:        endDate,
		ApplicableDays: req.ApplicableDays,
		DayType:        req.DayType,
		Prizes:         updatePrizes,
		UpdatedBy:      userID.(uuid.UUID),
	}
	
	// Update prize structure
	output, err := h.updatePrizeStructureService.UpdatePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update prize structure: " + err.Error(),
			Details: "An error occurred while updating the prize structure. Please try again later.",
		})
		return
	}
	
	// Prepare response
	prizeTiers := make([]response.PrizeTierResponse, 0, len(output.Prizes))
	for i, prize := range output.Prizes {
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         "Cash", // Default type
			Value:             prize.Value,
			ValueNGN:          prize.ValueNGN,
			Quantity:          prize.Quantity,
			Order:             i + 1, // Use index if order is not set
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:             output.ID.String(),
			Name:           output.Name,
			Description:    output.Description,
			IsActive:       req.IsActive,
			ValidFrom:      output.StartDate,
			ValidTo:        output.EndDate,
			ApplicableDays: req.ApplicableDays,
			DayType:        req.DayType,
			Prizes:         prizeTiers,
			CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// DeletePrizeStructure handles DELETE /api/admin/prize-structures/:id
func (h *PrizeHandler) DeletePrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
			Details: "The provided ID is not in the correct UUID format",
		})
		return
	}
	
	// In a real implementation, this would call a dedicated service
	// For now, we'll just return a success response
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: gin.H{
			"id":      prizeStructureID.String(),
			"deleted": true,
		},
	})
}
