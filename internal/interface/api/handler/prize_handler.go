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
	"github.com/ArowuTest/GP-Backend-Promo/internal/pkg/util"
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

// CreatePrize handles POST /api/admin/prizes
func (h *PrizeHandler) CreatePrize(c *gin.Context) {
	var req request.CreatePrizeTierRequest
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
	
	// Prepare input for CreatePrizeStructure service
	// Note: req.Description and req.ValueNGN don't exist in CreatePrizeTierRequest
	// Using req.Value instead and converting to int
	valueInt, err := strconv.Atoi(req.Value)
	if err != nil {
		valueInt = 0 // Default value if conversion fails
	}
	
	input := prizeApp.CreatePrizeStructureInput{
		Name:        "Prize Structure for " + req.Name,
		Description: "", // No Description field in CreatePrizeTierRequest
		StartDate:   time.Now().Format("2006-01-02"),
		EndDate:     time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		CreatedBy:   userID.(uuid.UUID),
		Prizes: []prizeApp.PrizeInput{
			{
				Name:        req.Name,
				Description: "", // No Description field in CreatePrizeTierRequest
				Value:       valueInt, // Convert string value to int
				Quantity:    req.Quantity,
			},
		},
	}
	
	// Create prize structure with prize
	output, err := h.createPrizeStructureService.CreatePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create prize: " + err.Error(),
		})
		return
	}
	
	// Use the first prize from the output
	var prize prizeApp.CreatePrizeOutput
	if len(output.Prizes) > 0 {
		prize = output.Prizes[0]
	} else {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "No prize created",
		})
		return
	}
	
	// Prepare response - PrizeTierResponse doesn't have Description or CreatedAt fields
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         "Cash", // Default value since it's not in the domain model
			Value:             strconv.Itoa(prize.Value), // Convert int to string
			Quantity:          prize.Quantity,
			Order:             0, // Default value since it's not in the domain model
			NumberOfRunnerUps: 1, // Default value
		},
	})
}

// GetPrizeByID handles GET /api/admin/prizes/:id
func (h *PrizeHandler) GetPrizeByID(c *gin.Context) {
	// Parse prize ID
	prizeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize ID format",
		})
		return
	}
	
	// Use GetPrizeStructure service to get the prize structure
	input := prizeApp.GetPrizeStructureInput{
		ID: prizeID,
	}
	
	output, err := h.getPrizeStructureService.GetPrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize: " + err.Error(),
		})
		return
	}
	
	// Find the prize in the prize structure
	var prize prizeApp.PrizeOutput
	if len(output.Prizes) > 0 {
		prize = output.Prizes[0]
	} else {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Success: false,
			Error:   "Prize not found",
		})
		return
	}
	
	// Prepare response - PrizeTierResponse doesn't have Description or CreatedAt fields
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeTierResponse{
			ID:                prize.ID.String(),
			Name:              prize.Name,
			PrizeType:         "Cash", // Default value since it's not in the domain model
			Value:             strconv.Itoa(prize.Value), // Convert int to string
			Quantity:          prize.Quantity,
			Order:             0, // Default value since it's not in the domain model
			NumberOfRunnerUps: 1, // Default value
		},
	})
}

// ListPrizes handles GET /api/admin/prizes
func (h *PrizeHandler) ListPrizes(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	
	// Use ListPrizeStructures service
	input := prizeApp.ListPrizeStructuresInput{
		Page:     page,
		PageSize: pageSize,
	}
	
	output, err := h.listPrizeStructuresService.ListPrizeStructures(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list prizes: " + err.Error(),
		})
		return
	}
	
	// Extract all prizes from all prize structures
	var allPrizes []response.PrizeTierResponse
	for _, ps := range output.PrizeStructures {
		for _, p := range ps.Prizes {
			allPrizes = append(allPrizes, response.PrizeTierResponse{
				ID:                p.ID.String(),
				Name:              p.Name,
				PrizeType:         "Cash", // Default value since it's not in the domain model
				Value:             strconv.Itoa(p.Value), // Convert int to string
				Quantity:          p.Quantity,
				Order:             0, // Default value since it's not in the domain model
				NumberOfRunnerUps: 1, // Default value
			})
		}
	}
	
	// Limit to pageSize
	start := 0
	end := len(allPrizes)
	if start < end {
		if start+pageSize < end {
			end = start + pageSize
		}
		allPrizes = allPrizes[start:end]
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    allPrizes,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  output.TotalCount,
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
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
	
	// Handle optional ValidTo field
	var endDate string
	if req.ValidTo != nil {
		endDate = *req.ValidTo
	}
	
	// Prepare input
	input := prizeApp.CreatePrizeStructureInput{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.ValidFrom,
		EndDate:     endDate,
		CreatedBy:   userID.(uuid.UUID),
		Prizes:      []prizeApp.PrizeInput{}, // Empty prizes, will be added later
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
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:             output.ID.String(),
			Name:           output.Name,
			Description:    output.Description,
			ValidFrom:      output.StartDate,
			ValidTo:        output.EndDate,
			IsActive:       true, // Default value since it's not in the domain model
			Prizes:         []response.PrizeTierResponse{}, // Empty slice since we just created it
			ApplicableDays: []string{}, // Empty slice since it's not provided
		},
	})
}

// GetPrizeStructureByID handles GET /api/admin/prize-structures/:id
func (h *PrizeHandler) GetPrizeStructureByID(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
		})
		return
	}
	
	// Get prize structure
	input := prizeApp.GetPrizeStructureInput{
		ID: prizeStructureID,
	}
	
	output, err := h.getPrizeStructureService.GetPrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize structure: " + err.Error(),
		})
		return
	}
	
	// Convert prizes to prize tiers
	prizeTiers := make([]response.PrizeTierResponse, 0, len(output.Prizes))
	for _, p := range output.Prizes {
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:                p.ID.String(),
			Name:              p.Name,
			PrizeType:         "Cash", // Default value since it's not in the domain model
			Value:             strconv.Itoa(p.Value), // Convert int to string
			Quantity:          p.Quantity,
			Order:             0, // Default value since it's not in the domain model
			NumberOfRunnerUps: 1, // Default value
		})
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:             output.ID.String(),
			Name:           output.Name,
			Description:    output.Description,
			ValidFrom:      util.FormatTimeOrEmpty(output.StartDate, "2006-01-02"),
			ValidTo:        util.FormatTimeOrEmpty(output.EndDate, "2006-01-02"),
			IsActive:       true, // Default value since it's not in the domain model
			Prizes:         prizeTiers,
			ApplicableDays: []string{}, // Empty slice since it's not provided
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
	
	// List prize structures
	input := prizeApp.ListPrizeStructuresInput{
		Page:     page,
		PageSize: pageSize,
	}
	
	output, err := h.listPrizeStructuresService.ListPrizeStructures(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list prize structures: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	structures := make([]response.PrizeStructureResponse, 0, len(output.PrizeStructures))
	for _, ps := range output.PrizeStructures {
		// Convert prizes to prize tiers
		prizeTiers := make([]response.PrizeTierResponse, 0, len(ps.Prizes))
		for _, p := range ps.Prizes {
			prizeTiers = append(prizeTiers, response.PrizeTierResponse{
				ID:                p.ID.String(),
				Name:              p.Name,
				PrizeType:         "Cash", // Default value since it's not in the domain model
				Value:             strconv.Itoa(p.Value), // Convert int to string
				Quantity:          p.Quantity,
				Order:             0, // Default value since it's not in the domain model
				NumberOfRunnerUps: 1, // Default value
			})
		}
		
		structures = append(structures, response.PrizeStructureResponse{
			ID:             ps.ID.String(),
			Name:           ps.Name,
			Description:    ps.Description,
			ValidFrom:      util.FormatTimeOrEmpty(ps.StartDate, "2006-01-02"),
			ValidTo:        util.FormatTimeOrEmpty(ps.EndDate, "2006-01-02"),
			IsActive:       true, // Default value since it's not in the domain model
			Prizes:         prizeTiers,
			ApplicableDays: []string{}, // Empty slice since it's not provided
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    structures,
		Pagination: response.Pagination{
			Page:       output.Page,
			PageSize:   output.PageSize,
			TotalRows:  output.TotalCount,
			TotalPages: output.TotalPages,
			TotalItems: int64(output.TotalCount),
		},
	})
}

// AddPrizeTier handles POST /api/admin/prize-structures/:id/prize-tiers
func (h *PrizeHandler) AddPrizeTier(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
		})
		return
	}
	
	var req request.CreatePrizeTierRequest
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
	
	// Get existing prize structure
	getInput := prizeApp.GetPrizeStructureInput{
		ID: prizeStructureID,
	}
	
	prizeStructure, err := h.getPrizeStructureService.GetPrizeStructure(c.Request.Context(), getInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize structure: " + err.Error(),
		})
		return
	}
	
	// Prepare update input
	updateInput := prizeApp.UpdatePrizeStructureInput{
		ID:          prizeStructureID,
		Name:        prizeStructure.Name,
		Description: prizeStructure.Description,
		StartDate:   util.FormatTimeOrEmpty(prizeStructure.StartDate, "2006-01-02"),
		EndDate:     util.FormatTimeOrEmpty(prizeStructure.EndDate, "2006-01-02"),
		UpdatedBy:   userID.(uuid.UUID),
		Prizes:      []prizeApp.UpdatePrizeInput{},
	}
	
	// Add existing prizes
	for _, p := range prizeStructure.Prizes {
		updateInput.Prizes = append(updateInput.Prizes, prizeApp.UpdatePrizeInput{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Value:       p.Value,
			Quantity:    p.Quantity,
		})
	}
	
	// Convert string value to int
	valueInt, err := strconv.Atoi(req.Value)
	if err != nil {
		valueInt = 0 // Default value if conversion fails
	}
	
	// Add new prize tier
	updateInput.Prizes = append(updateInput.Prizes, prizeApp.UpdatePrizeInput{
		ID:          uuid.New(), // Generate new ID
		Name:        req.Name,
		Description: "", // No Description field in CreatePrizeTierRequest
		Value:       valueInt, // Convert string value to int
		Quantity:    req.Quantity,
	})
	
	// Update prize structure
	_, err = h.updatePrizeStructureService.UpdatePrizeStructure(c.Request.Context(), updateInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to add prize tier to prize structure: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    "Prize tier added to prize structure successfully",
	})
}
