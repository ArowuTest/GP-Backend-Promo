package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	prizeApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
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
	deletePrizeStructureService *prizeApp.DeletePrizeStructureService
}

// NewPrizeHandler creates a new PrizeHandler
func NewPrizeHandler(
	createPrizeStructureService *prizeApp.CreatePrizeStructureService,
	getPrizeStructureService *prizeApp.GetPrizeStructureService,
	listPrizeStructuresService *prizeApp.ListPrizeStructuresService,
	updatePrizeStructureService *prizeApp.UpdatePrizeStructureService,
	deletePrizeStructureService *prizeApp.DeletePrizeStructureService,
) *PrizeHandler {
	return &PrizeHandler{
		createPrizeStructureService: createPrizeStructureService,
		getPrizeStructureService:    getPrizeStructureService,
		listPrizeStructuresService:  listPrizeStructuresService,
		updatePrizeStructureService: updatePrizeStructureService,
		deletePrizeStructureService: deletePrizeStructureService,
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
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Type assertion with safety check
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		// Try to parse as string if not UUID
		if userIDStr, ok := userIDValue.(string); ok {
			var err error
			userID, err = uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Success: false,
					Error:   "Invalid user ID format in token",
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID type in token",
			})
			return
		}
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid start date format. Expected YYYY-MM-DD",
		})
		return
	}

	var endDate time.Time
	if req.ValidTo != "" {
		endDate, err = time.Parse("2006-01-02", req.ValidTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid end date format. Expected YYYY-MM-DD",
			})
			return
		}
	}

	// Convert prizes
	prizes := make([]prize.CreatePrizeInput, 0, len(req.Prizes))
	for _, p := range req.Prizes {
		// Parse the currency string to float64 using the utility
		value, err := util.ParseCurrency(p.Value)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid prize value format: " + err.Error(),
			})
			return
		}
		
		prizes = append(prizes, prize.CreatePrizeInput{
			Name:              p.Name,
			Description:       p.Description,
			Value:             value, // Using converted float64 value
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create input for application layer
	appInput := prizeApp.CreatePrizeStructureInput{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Prizes:      make([]prizeApp.PrizeInput, 0, len(prizes)),
		CreatedBy:   userID,
		IsActive:    req.IsActive,
	}
	
	// Convert domain prizes to application prizes
	for _, p := range prizes {
		appInput.Prizes = append(appInput.Prizes, prizeApp.PrizeInput{
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create prize structure
	result, err := h.createPrizeStructureService.CreatePrizeStructure(c.Request.Context(), appInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create prize structure: " + err.Error(),
		})
		return
	}

	// Convert prizes for response
	prizesResponse := make([]response.PrizeResponse, 0, len(result.Prizes))
	for _, p := range result.Prizes {
		prizesResponse = append(prizesResponse, response.PrizeResponse{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             util.FormatCurrency(p.Value, "N"), // Format as currency string
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create response
	resp := response.PrizeStructureResponse{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		ValidFrom:   result.StartDate.Format("2006-01-02"),
		ValidTo:     result.EndDate.Format("2006-01-02"),
		Prizes:      prizesResponse,
		IsActive:    result.IsActive,
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Message: "Prize structure created successfully",
		Data:    resp,
	})
}

// GetPrizeStructure handles GET /api/admin/prize-structures/:id
func (h *PrizeHandler) GetPrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Prize structure ID is required",
		})
		return
	}

	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
		})
		return
	}

	// Create input for application layer
	appInput := prizeApp.GetPrizeStructureInput{
		ID: prizeStructureID,
	}

	// Get prize structure
	result, err := h.getPrizeStructureService.GetPrizeStructure(c.Request.Context(), appInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize structure: " + err.Error(),
		})
		return
	}

	// Convert prizes for response
	prizesResponse := make([]response.PrizeResponse, 0, len(result.Prizes))
	for _, p := range result.Prizes {
		prizesResponse = append(prizesResponse, response.PrizeResponse{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             util.FormatCurrency(p.Value, "N"), // Format as currency string
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create response
	resp := response.PrizeStructureResponse{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		ValidFrom:   result.StartDate.Format("2006-01-02"),
		ValidTo:     result.EndDate.Format("2006-01-02"),
		Prizes:      prizesResponse,
		IsActive:    result.IsActive,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Prize structure retrieved successfully",
		Data:    resp,
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

	// Convert results for response
	prizeStructuresResponse := make([]response.PrizeStructureResponse, 0, len(output.PrizeStructures))
	for _, ps := range output.PrizeStructures {
		// Convert prizes for response
		prizesResponse := make([]response.PrizeResponse, 0, len(ps.Prizes))
		for _, p := range ps.Prizes {
			prizesResponse = append(prizesResponse, response.PrizeResponse{
				ID:                p.ID,
				Name:              p.Name,
				Description:       p.Description,
				Value:             util.FormatCurrency(p.Value, "N"), // Format as currency string
				Quantity:          p.Quantity,
				NumberOfRunnerUps: p.NumberOfRunnerUps,
			})
		}

		prizeStructuresResponse = append(prizeStructuresResponse, response.PrizeStructureResponse{
			ID:          ps.ID,
			Name:        ps.Name,
			Description: ps.Description,
			ValidFrom:   ps.StartDate.Format("2006-01-02"),
			ValidTo:     ps.EndDate.Format("2006-01-02"),
			Prizes:      prizesResponse,
			IsActive:    ps.IsActive,
		})
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    prizeStructuresResponse,
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
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Prize structure ID is required",
		})
		return
	}

	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
		})
		return
	}

	// Parse request
	var req request.UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Type assertion with safety check
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		// Try to parse as string if not UUID
		if userIDStr, ok := userIDValue.(string); ok {
			var err error
			userID, err = uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Success: false,
					Error:   "Invalid user ID format in token",
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID type in token",
			})
			return
		}
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid start date format. Expected YYYY-MM-DD",
		})
		return
	}

	var endDate time.Time
	if req.ValidTo != "" {
		endDate, err = time.Parse("2006-01-02", req.ValidTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid end date format. Expected YYYY-MM-DD",
			})
			return
		}
	}

	// Convert prizes
	prizes := make([]prize.UpdatePrizeInput, 0, len(req.Prizes))
	for _, p := range req.Prizes {
		var prizeID uuid.UUID
		if p.ID != "" {
			prizeID, err = uuid.Parse(p.ID)
			if err != nil {
				c.JSON(http.StatusBadRequest, response.ErrorResponse{
					Success: false,
					Error:   "Invalid prize ID format",
				})
				return
			}
		}
		
		// Parse the currency string to float64 using the utility
		value, err := util.ParseCurrency(p.Value)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid prize value format: " + err.Error(),
			})
			return
		}

		prizes = append(prizes, prize.UpdatePrizeInput{
			ID:                prizeID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             value, // Using converted float64 value
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create input for application layer
	appInput := prizeApp.UpdatePrizeStructureInput{
		ID:          prizeStructureID,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Prizes:      make([]prizeApp.UpdatePrizeInput, 0, len(prizes)),
		UpdatedBy:   userID,
		IsActive:    req.IsActive,
	}
	
	// Convert domain prizes to application prizes
	for _, p := range prizes {
		appInput.Prizes = append(appInput.Prizes, prizeApp.UpdatePrizeInput{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Update prize structure
	result, err := h.updatePrizeStructureService.UpdatePrizeStructure(c.Request.Context(), appInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update prize structure: " + err.Error(),
		})
		return
	}

	// Convert prizes for response
	prizesResponse := make([]response.PrizeResponse, 0, len(result.Prizes))
	for _, p := range result.Prizes {
		prizesResponse = append(prizesResponse, response.PrizeResponse{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             util.FormatCurrency(p.Value, "N"), // Format as currency string
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		})
	}

	// Create response
	resp := response.PrizeStructureResponse{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		ValidFrom:   startDate.Format("2006-01-02"),
		ValidTo:     endDate.Format("2006-01-02"),
		Prizes:      prizesResponse,
		IsActive:    result.IsActive,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Prize structure updated successfully",
		Data:    resp,
	})
}

// DeletePrizeStructure handles DELETE /api/admin/prize-structures/:id
func (h *PrizeHandler) DeletePrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Prize structure ID is required",
		})
		return
	}

	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID format",
		})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Type assertion with safety check
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		// Try to parse as string if not UUID
		if userIDStr, ok := userIDValue.(string); ok {
			var err error
			userID, err = uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Success: false,
					Error:   "Invalid user ID format in token",
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Success: false,
				Error:   "Invalid user ID type in token",
			})
			return
		}
	}

	// Create input for application layer
	input := prizeApp.DeletePrizeStructureInput{
            ID:        prizeStructureID,
            DeletedBy: userID,
	}

	// Delete prize structure
	err = h.deletePrizeStructureService.DeletePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete prize structure: " + err.Error(),
		})
		return
	}

	// Create response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Prize structure deleted successfully",
	})
}
