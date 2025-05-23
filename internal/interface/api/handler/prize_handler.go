package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// PrizeHandler handles HTTP requests related to prize structures
type PrizeHandler struct {
	createPrizeStructureService *prize.CreatePrizeStructureService
	getPrizeStructureService    *prize.GetPrizeStructureService
	listPrizeStructuresService  *prize.ListPrizeStructuresService
	updatePrizeStructureService *prize.UpdatePrizeStructureService
}

// NewPrizeHandler creates a new PrizeHandler
func NewPrizeHandler(
	createPrizeStructureService *prize.CreatePrizeStructureService,
	getPrizeStructureService *prize.GetPrizeStructureService,
	listPrizeStructuresService *prize.ListPrizeStructuresService,
	updatePrizeStructureService *prize.UpdatePrizeStructureService,
) *PrizeHandler {
	return &PrizeHandler{
		createPrizeStructureService: createPrizeStructureService,
		getPrizeStructureService:    getPrizeStructureService,
		listPrizeStructuresService:  listPrizeStructuresService,
		updatePrizeStructureService: updatePrizeStructureService,
	}
}

// CreatePrizeStructure handles the creation of a new prize structure
func (h *PrizeHandler) CreatePrizeStructure(c *gin.Context) {
	var req request.CreatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Details: "User ID not found in context",
		})
		return
	}

	// Parse dates
	validFrom, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid valid_from date",
			Details: err.Error(),
		})
		return
	}

	var validTo time.Time
	if req.ValidTo != "" {
		validTo, err = time.Parse("2006-01-02", req.ValidTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid valid_to date",
				Details: err.Error(),
			})
			return
		}
	}

	// Convert prizes
	prizes := make([]prize.CreatePrizeInput, len(req.Prizes))
	for i, p := range req.Prizes {
		prizes[i] = prize.CreatePrizeInput{
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		}
	}

	// Create input for service
	input := prize.CreatePrizeStructureInput{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   validFrom,
		EndDate:     validTo,
		Prizes:      prizes,
		CreatedBy:   userID.(uuid.UUID),
	}

	// Call service
	result, err := h.createPrizeStructureService.CreatePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create prize structure",
			Details: err.Error(),
		})
		return
	}

	// Convert prizes for response
	prizesResponse := make([]response.PrizeResponse, len(result.Prizes))
	for i, p := range result.Prizes {
		prizesResponse[i] = response.PrizeResponse{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		}
	}

	// Create response
	resp := response.PrizeStructureResponse{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		ValidFrom:   result.StartDate.Format("2006-01-02"),
		ValidTo:     result.EndDate.Format("2006-01-02"),
		Prizes:      prizesResponse,
		IsActive:    true,
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Message: "Prize structure created successfully",
		Data:    resp,
	})
}

// GetPrizeStructure handles the retrieval of a prize structure by ID
func (h *PrizeHandler) GetPrizeStructure(c *gin.Context) {
	// Get prize structure ID from URL
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: "Prize structure ID is required",
		})
		return
	}

	// Parse UUID
	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: err.Error(),
		})
		return
	}

	// Call service
	result, err := h.getPrizeStructureService.GetPrizeStructure(c.Request.Context(), prizeStructureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize structure",
			Details: err.Error(),
		})
		return
	}

	// Convert prizes for response
	prizesResponse := make([]response.PrizeResponse, len(result.Prizes))
	for i, p := range result.Prizes {
		prizesResponse[i] = response.PrizeResponse{
			ID:                p.ID,
			Name:              p.Name,
			Description:       p.Description,
			Value:             p.Value,
			Quantity:          p.Quantity,
			NumberOfRunnerUps: p.NumberOfRunnerUps,
		}
	}

	// Create response
	resp := response.PrizeStructureResponse{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		ValidFrom:   result.StartDate.Format("2006-01-02"),
		ValidTo:     result.EndDate.Format("2006-01-02"),
		Prizes:      prizesResponse,
		IsActive:    true,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Prize structure retrieved successfully",
		Data:    resp,
	})
}

// ListPrizeStructures handles the retrieval of all prize structures
func (h *PrizeHandler) ListPrizeStructures(c *gin.Context) {
	// Call service
	results, err := h.listPrizeStructuresService.ListPrizeStructures(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list prize structures",
			Details: err.Error(),
		})
		return
	}

	// Convert results for response
	prizeStructuresResponse := make([]response.PrizeStructureResponse, len(results))
	for i, ps := range results {
		// Convert prizes for response
		prizesResponse := make([]response.PrizeResponse, len(ps.Prizes))
		for j, p := range ps.Prizes {
			prizesResponse[j] = response.PrizeResponse{
				ID:                p.ID,
				Name:              p.Name,
				Description:       p.Description,
				Value:             p.Value,
				Quantity:          p.Quantity,
				NumberOfRunnerUps: p.NumberOfRunnerUps,
			}
		}

		prizeStructuresResponse[i] = response.PrizeStructureResponse{
			ID:          ps.ID,
			Name:        ps.Name,
			Description: ps.Description,
			ValidFrom:   ps.StartDate.Format("2006-01-02"),
			ValidTo:     ps.EndDate.Format("2006-01-02"),
			Prizes:      prizesResponse,
			IsActive:    true,
		}
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Prize structures retrieved successfully",
		Data:    prizeStructuresResponse,
	})
}

// UpdatePrizeStructure handles the update of a prize structure
func (h *PrizeHandler) UpdatePrizeStructure(c *gin.Context) {
	// Get prize structure ID from URL
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: "Prize structure ID is required",
		})
		return
	}

	// Parse UUID
	prizeStructureID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: err.Error(),
		})
		return
	}

	// Parse request body
	var req request.UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Details: "User ID not found in context",
		})
		return
	}

	// Convert prizes
	prizes := make([]prize.UpdatePrizeInput, len(req.Prizes))
	for i, p := range req.Prizes {
		var prizeID uuid.UUID
		if p.ID != "" {
			prizeID, err = uuid.Parse(p.ID)
			if err != nil {
				c.JSON(http.StatusBadRequest, response.ErrorResponse{
					Success: false,
					Error:   "Invalid prize ID",
					Details: err.Error(),
				})
				return
			}
		}

		prizes[i] = prize.UpdatePrizeInput{
			ID:          prizeID,
			Name:        p.Name,
			Description: p.Description,
			Value:       p.Value,
			Quantity:    p.Quantity,
		}
	}

	// Create input for service
	input := prize.UpdatePrizeStructureInput{
		ID:          prizeStructureID,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.ValidFrom,
		EndDate:     req.ValidTo,
		Prizes:      prizes,
		UpdatedBy:   userID.(uuid.UUID),
	}

	// Call service
	result, err := h.updatePrizeStructureService.UpdatePrizeStructure(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to update prize structure",
			Details: err.Error(),
		})
		return
	}

	// Convert prizes for response
	prizesResponse := make([]response.PrizeResponse, len(result.Prizes))
	for i, p := range result.Prizes {
		prizesResponse[i] = response.PrizeResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Value:       p.Value,
			Quantity:    p.Quantity,
		}
	}

	// Create response
	resp := response.PrizeStructureResponse{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		ValidFrom:   result.StartDate,
		ValidTo:     result.EndDate,
		Prizes:      prizesResponse,
		IsActive:    true,
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Prize structure updated successfully",
		Data:    resp,
	})
}
