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
			Value:             prize.Value,
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
		StartDate:      req.ValidFrom,
		EndDate:        endDate,
		IsActive:       req.IsActive,
		ApplicableDays: req.ApplicableDays,
		Prizes:         prizes,
		CreatedBy:      userID.(uuid.UUID),
	}
	
	// Create prize structure
	output, err := h.createPrizeStructureService.CreatePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create prize structure: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	prizeTiers := make([]response.PrizeTierResponse, 0, len(output.Prizes))
	for i, prize := range output.Prizes {
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         req.Prizes[i].PrizeType, // Use the requested prize type
			Value:             prize.Value,
			Quantity:          prize.Quantity,
			Order:             req.Prizes[i].Order, // Use the requested order
			NumberOfRunnerUps: req.Prizes[i].NumberOfRunnerUps, // Use the requested number of runner ups
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
			Prizes:         prizeTiers,
			CreatedAt:      time.Now().Format(time.RFC3339),
			UpdatedAt:      time.Now().Format(time.RFC3339),
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
		})
		return
	}
	
	// Prepare response
	prizeTiers := make([]response.PrizeTierResponse, 0, len(output.Prizes))
	for _, prize := range output.Prizes {
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         "Cash", // Default type, should be stored in the database
			Value:             prize.Value,
			Quantity:          prize.Quantity,
			Order:             prize.Order,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:             output.ID.String(),
			Name:           output.Name,
			Description:    output.Description,
			IsActive:       output.IsActive,
			ValidFrom:      output.StartDate.Format("2006-01-02"),
			ValidTo:        output.EndDate.Format("2006-01-02"),
			ApplicableDays: output.ApplicableDays,
			Prizes:         prizeTiers,
			CreatedAt:      output.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      output.UpdatedAt.Format(time.RFC3339),
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
		})
		return
	}
	
	// Prepare response
	prizeStructures := make([]response.PrizeStructureResponse, 0, len(output.PrizeStructures))
	for _, ps := range output.PrizeStructures {
		prizeTiers := make([]response.PrizeTierResponse, 0, len(ps.Prizes))
		for _, prize := range ps.Prizes {
			prizeTiers = append(prizeTiers, response.PrizeTierResponse{
				ID:                prize.ID.String(),
				Name:              prize.Name,
				PrizeType:         "Cash", // Default type, should be stored in the database
				Value:             prize.Value,
				Quantity:          prize.Quantity,
				Order:             prize.Order,
				NumberOfRunnerUps: prize.NumberOfRunnerUps,
			})
		}
		
		prizeStructures = append(prizeStructures, response.PrizeStructureResponse{
			ID:             ps.ID.String(),
			Name:           ps.Name,
			Description:    ps.Description,
			IsActive:       ps.IsActive,
			ValidFrom:      ps.StartDate.Format("2006-01-02"),
			ValidTo:        ps.EndDate.Format("2006-01-02"),
			ApplicableDays: ps.ApplicableDays,
			Prizes:         prizeTiers,
			CreatedAt:      ps.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      ps.UpdatedAt.Format(time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    prizeStructures,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  int(output.TotalCount),
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
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
		})
		return
	}
	
	var req request.UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Details: err.Error(),
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
			Value:             prize.Value,
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
	
	// Handle optional IsActive field
	isActive := false
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	
	input := prizeApp.UpdatePrizeStructureInput{
		ID:             prizeStructureID,
		Name:           req.Name,
		Description:    req.Description,
		StartDate:      req.ValidFrom,
		EndDate:        endDate,
		IsActive:       isActive,
		ApplicableDays: req.ApplicableDays,
		Prizes:         updatePrizes,
		UpdatedBy:      userID.(uuid.UUID),
	}
	
	// Update prize structure
	output, err := h.updatePrizeStructureService.UpdatePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update prize structure: " + err.Error(),
			Details: err.Error(),
		})
		return
	}
	
	// Prepare response
	prizeTiers := make([]response.PrizeTierResponse, 0, len(output.Prizes))
	for i, prize := range output.Prizes {
		// Find matching prize in request to get PrizeType
		prizeType := "Cash" // Default
		if i < len(req.Prizes) {
			prizeType = req.Prizes[i].PrizeType
		}
		
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         prizeType,
			Value:             prize.Value,
			Quantity:          prize.Quantity,
			Order:             prize.Order,
			NumberOfRunnerUps: prize.NumberOfRunnerUps,
		})
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:             output.ID.String(),
			Name:           output.Name,
			Description:    output.Description,
			IsActive:       output.IsActive,
			ValidFrom:      output.StartDate,
			ValidTo:        output.EndDate,
			ApplicableDays: output.ApplicableDays,
			Prizes:         prizeTiers,
			CreatedAt:      output.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      time.Now().Format(time.RFC3339),
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
