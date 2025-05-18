package interface

import (
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize/entity"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/request"
	"github.com/ArowuTest/GP-Backend-Promo/internal/interface/dto/response"
)

// PrizeHandler handles HTTP requests related to prizes
type PrizeHandler struct {
	createPrizeStructure *prize.CreatePrizeStructureUseCase
	getPrizeStructure    *prize.GetPrizeStructureUseCase
	listPrizeStructures  *prize.ListPrizeStructuresUseCase
	updatePrizeStructure *prize.UpdatePrizeStructureUseCase
	deletePrizeStructure *prize.DeletePrizeStructureUseCase
	getActivePrizeStructure *prize.GetActivePrizeStructureUseCase
}

// NewPrizeHandler creates a new PrizeHandler
func NewPrizeHandler(
	createPrizeStructure *prize.CreatePrizeStructureUseCase,
	getPrizeStructure *prize.GetPrizeStructureUseCase,
	listPrizeStructures *prize.ListPrizeStructuresUseCase,
	updatePrizeStructure *prize.UpdatePrizeStructureUseCase,
	deletePrizeStructure *prize.DeletePrizeStructureUseCase,
	getActivePrizeStructure *prize.GetActivePrizeStructureUseCase,
) *PrizeHandler {
	return &PrizeHandler{
		createPrizeStructure: createPrizeStructure,
		getPrizeStructure:    getPrizeStructure,
		listPrizeStructures:  listPrizeStructures,
		updatePrizeStructure: updatePrizeStructure,
		deletePrizeStructure: deletePrizeStructure,
		getActivePrizeStructure: getActivePrizeStructure,
	}
}

// CreatePrizeStructure handles the request to create a prize structure
func (h *PrizeHandler) CreatePrizeStructure(c *gin.Context) {
	var req request.CreatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Parse dates
	validFrom, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid valid from date format",
			Details: "Date must be in YYYY-MM-DD format",
		})
		return
	}

	var validTo *time.Time
	if req.ValidTo != "" {
		parsedValidTo, err := time.Parse("2006-01-02", req.ValidTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid valid to date format",
				Details: "Date must be in YYYY-MM-DD format",
			})
			return
		}
		validTo = &parsedValidTo
	}

	// Convert prize tiers
	prizeTiers := make([]prize.PrizeTierInput, 0, len(req.Prizes))
	for _, pt := range req.Prizes {
		prizeTier := prize.PrizeTierInput{
			Rank:        pt.Rank,
			Name:        pt.Name,
			Description: pt.Description,
			Value:       pt.Value,
			ValueNGN:    pt.ValueNGN,
			Quantity:    pt.Quantity,
		}
		prizeTiers = append(prizeTiers, prizeTier)
	}

	// Create prize structure
	input := prize.CreatePrizeStructureInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		Prizes:      prizeTiers,
	}

	output, err := h.createPrizeStructure.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.PrizeError:
			prizeErr := err.(*entity.PrizeError)
			switch prizeErr.Code() {
			case entity.ErrInvalidPrizeStructure:
				statusCode = http.StatusBadRequest
				errorMessage = "Invalid prize structure"
			case entity.ErrInvalidPrizeTier:
				statusCode = http.StatusBadRequest
				errorMessage = "Invalid prize tier"
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to create prize structure"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to create prize structure"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Convert prize tiers to response format
	prizes := make([]response.PrizeTierResponse, 0, len(output.PrizeStructure.Prizes))
	for _, pt := range output.PrizeStructure.Prizes {
		prizeTier := response.PrizeTierResponse{
			ID:          pt.ID.String(),
			Rank:        pt.Rank,
			Name:        pt.Name,
			Description: pt.Description,
			Value:       pt.Value,
			ValueNGN:    pt.ValueNGN,
			Quantity:    pt.Quantity,
		}
		prizes = append(prizes, prizeTier)
	}

	// Format dates for response
	validToStr := ""
	if output.PrizeStructure.ValidTo != nil {
		validToStr = output.PrizeStructure.ValidTo.Format("2006-01-02")
	}

	// Prepare response
	resp := response.PrizeStructureResponse{
		ID:          output.PrizeStructure.ID.String(),
		Name:        output.PrizeStructure.Name,
		Description: output.PrizeStructure.Description,
		IsActive:    output.PrizeStructure.IsActive,
		ValidFrom:   output.PrizeStructure.ValidFrom.Format("2006-01-02"),
		ValidTo:     validToStr,
		Prizes:      prizes,
		CreatedAt:   output.PrizeStructure.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// GetPrizeStructure handles the request to get a prize structure by ID
func (h *PrizeHandler) GetPrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureIDStr := c.Param("id")
	prizeStructureID, err := uuid.Parse(prizeStructureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: err.Error(),
		})
		return
	}

	// Get prize structure
	input := prize.GetPrizeStructureInput{
		PrizeStructureID: prizeStructureID,
	}

	output, err := h.getPrizeStructure.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.PrizeError:
			prizeErr := err.(*entity.PrizeError)
			if prizeErr.Code() == entity.ErrPrizeStructureNotFound {
				statusCode = http.StatusNotFound
				errorMessage = "Prize structure not found"
			} else {
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to get prize structure"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to get prize structure"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Convert prize tiers to response format
	prizes := make([]response.PrizeTierResponse, 0, len(output.PrizeStructure.Prizes))
	for _, pt := range output.PrizeStructure.Prizes {
		prizeTier := response.PrizeTierResponse{
			ID:          pt.ID.String(),
			Rank:        pt.Rank,
			Name:        pt.Name,
			Description: pt.Description,
			Value:       pt.Value,
			ValueNGN:    pt.ValueNGN,
			Quantity:    pt.Quantity,
		}
		prizes = append(prizes, prizeTier)
	}

	// Format dates for response
	validToStr := ""
	if output.PrizeStructure.ValidTo != nil {
		validToStr = output.PrizeStructure.ValidTo.Format("2006-01-02")
	}

	// Prepare response
	resp := response.PrizeStructureResponse{
		ID:          output.PrizeStructure.ID.String(),
		Name:        output.PrizeStructure.Name,
		Description: output.PrizeStructure.Description,
		IsActive:    output.PrizeStructure.IsActive,
		ValidFrom:   output.PrizeStructure.ValidFrom.Format("2006-01-02"),
		ValidTo:     validToStr,
		Prizes:      prizes,
		CreatedAt:   output.PrizeStructure.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// ListPrizeStructures handles the request to list prize structures
func (h *PrizeHandler) ListPrizeStructures(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// List prize structures
	input := prize.ListPrizeStructuresInput{
		Page:     page,
		PageSize: pageSize,
	}

	output, err := h.listPrizeStructures.Execute(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list prize structures",
			Details: err.Error(),
		})
		return
	}

	// Convert prize structures to response format
	prizeStructures := make([]response.PrizeStructureResponse, 0, len(output.PrizeStructures))
	for _, ps := range output.PrizeStructures {
		// Convert prize tiers to response format
		prizes := make([]response.PrizeTierResponse, 0, len(ps.Prizes))
		for _, pt := range ps.Prizes {
			prizeTier := response.PrizeTierResponse{
				ID:          pt.ID.String(),
				Rank:        pt.Rank,
				Name:        pt.Name,
				Description: pt.Description,
				Value:       pt.Value,
				ValueNGN:    pt.ValueNGN,
				Quantity:    pt.Quantity,
			}
			prizes = append(prizes, prizeTier)
		}

		// Format dates for response
		validToStr := ""
		if ps.ValidTo != nil {
			validToStr = ps.ValidTo.Format("2006-01-02")
		}

		prizeStructure := response.PrizeStructureResponse{
			ID:          ps.ID.String(),
			Name:        ps.Name,
			Description: ps.Description,
			IsActive:    ps.IsActive,
			ValidFrom:   ps.ValidFrom.Format("2006-01-02"),
			ValidTo:     validToStr,
			Prizes:      prizes,
			CreatedAt:   ps.CreatedAt.Format(time.RFC3339),
		}
		prizeStructures = append(prizeStructures, prizeStructure)
	}

	// Prepare response
	resp := response.PaginatedResponse{
		Success: true,
		Data:    prizeStructures,
		Pagination: response.Pagination{
			Page:      page,
			PageSize:  pageSize,
			TotalRows: output.Total,
			TotalPages: (output.Total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// UpdatePrizeStructure handles the request to update a prize structure
func (h *PrizeHandler) UpdatePrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureIDStr := c.Param("id")
	prizeStructureID, err := uuid.Parse(prizeStructureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: err.Error(),
		})
		return
	}

	var req request.UpdatePrizeStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Parse dates
	validFrom, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid valid from date format",
			Details: "Date must be in YYYY-MM-DD format",
		})
		return
	}

	var validTo *time.Time
	if req.ValidTo != "" {
		parsedValidTo, err := time.Parse("2006-01-02", req.ValidTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid valid to date format",
				Details: "Date must be in YYYY-MM-DD format",
			})
			return
		}
		validTo = &parsedValidTo
	}

	// Convert prize tiers
	prizeTiers := make([]prize.PrizeTierInput, 0, len(req.Prizes))
	for _, pt := range req.Prizes {
		prizeTier := prize.PrizeTierInput{
			Rank:        pt.Rank,
			Name:        pt.Name,
			Description: pt.Description,
			Value:       pt.Value,
			ValueNGN:    pt.ValueNGN,
			Quantity:    pt.Quantity,
		}
		prizeTiers = append(prizeTiers, prizeTier)
	}

	// Update prize structure
	input := prize.UpdatePrizeStructureInput{
		PrizeStructureID: prizeStructureID,
		Name:             req.Name,
		Description:      req.Description,
		IsActive:         req.IsActive,
		ValidFrom:        validFrom,
		ValidTo:          validTo,
		Prizes:           prizeTiers,
	}

	output, err := h.updatePrizeStructure.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.PrizeError:
			prizeErr := err.(*entity.PrizeError)
			switch prizeErr.Code() {
			case entity.ErrPrizeStructureNotFound:
				statusCode = http.StatusNotFound
				errorMessage = "Prize structure not found"
			case entity.ErrInvalidPrizeStructure:
				statusCode = http.StatusBadRequest
				errorMessage = "Invalid prize structure"
			case entity.ErrInvalidPrizeTier:
				statusCode = http.StatusBadRequest
				errorMessage = "Invalid prize tier"
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to update prize structure"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to update prize structure"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Convert prize tiers to response format
	prizes := make([]response.PrizeTierResponse, 0, len(output.PrizeStructure.Prizes))
	for _, pt := range output.PrizeStructure.Prizes {
		prizeTier := response.PrizeTierResponse{
			ID:          pt.ID.String(),
			Rank:        pt.Rank,
			Name:        pt.Name,
			Description: pt.Description,
			Value:       pt.Value,
			ValueNGN:    pt.ValueNGN,
			Quantity:    pt.Quantity,
		}
		prizes = append(prizes, prizeTier)
	}

	// Format dates for response
	validToStr := ""
	if output.PrizeStructure.ValidTo != nil {
		validToStr = output.PrizeStructure.ValidTo.Format("2006-01-02")
	}

	// Prepare response
	resp := response.PrizeStructureResponse{
		ID:          output.PrizeStructure.ID.String(),
		Name:        output.PrizeStructure.Name,
		Description: output.PrizeStructure.Description,
		IsActive:    output.PrizeStructure.IsActive,
		ValidFrom:   output.PrizeStructure.ValidFrom.Format("2006-01-02"),
		ValidTo:     validToStr,
		Prizes:      prizes,
		CreatedAt:   output.PrizeStructure.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   output.PrizeStructure.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// DeletePrizeStructure handles the request to delete a prize structure
func (h *PrizeHandler) DeletePrizeStructure(c *gin.Context) {
	// Parse prize structure ID
	prizeStructureIDStr := c.Param("id")
	prizeStructureID, err := uuid.Parse(prizeStructureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize structure ID",
			Details: err.Error(),
		})
		return
	}

	// Delete prize structure
	input := prize.DeletePrizeStructureInput{
		PrizeStructureID: prizeStructureID,
	}

	err = h.deletePrizeStructure.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.PrizeError:
			prizeErr := err.(*entity.PrizeError)
			if prizeErr.Code() == entity.ErrPrizeStructureNotFound {
				statusCode = http.StatusNotFound
				errorMessage = "Prize structure not found"
			} else {
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to delete prize structure"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to delete prize structure"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    "Prize structure deleted successfully",
	})
}

// GetActivePrizeStructure handles the request to get the active prize structure for a date
func (h *PrizeHandler) GetActivePrizeStructure(c *gin.Context) {
	// Parse date
	dateStr := c.Query("date")
	var date time.Time
	var err error
	
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Success: false,
				Error:   "Invalid date format",
				Details: "Date must be in YYYY-MM-DD format",
			})
			return
		}
	}

	// Get active prize structure
	input := prize.GetActivePrizeStructureInput{
		Date: date,
	}

	output, err := h.getActivePrizeStructure.Execute(input)
	if err != nil {
		var statusCode int
		var errorMessage string

		// Handle domain-specific errors
		switch err.(type) {
		case *entity.PrizeError:
			prizeErr := err.(*entity.PrizeError)
			if prizeErr.Code() == entity.ErrNoPrizeStructureActive {
				statusCode = http.StatusNotFound
				errorMessage = "No active prize structure found for the given date"
			} else {
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to get active prize structure"
			}
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to get active prize structure"
		}

		c.JSON(statusCode, response.ErrorResponse{
			Success: false,
			Error:   errorMessage,
			Details: err.Error(),
		})
		return
	}

	// Convert prize tiers to response format
	prizes := make([]response.PrizeTierResponse, 0, len(output.PrizeStructure.Prizes))
	for _, pt := range output.PrizeStructure.Prizes {
		prizeTier := response.PrizeTierResponse{
			ID:          pt.ID.String(),
			Rank:        pt.Rank,
			Name:        pt.Name,
			Description: pt.Description,
			Value:       pt.Value,
			ValueNGN:    pt.ValueNGN,
			Quantity:    pt.Quantity,
		}
		prizes = append(prizes, prizeTier)
	}

	// Format dates for response
	validToStr := ""
	if output.PrizeStructure.ValidTo != nil {
		validToStr = output.PrizeStructure.ValidTo.Format("2006-01-02")
	}

	// Prepare response
	resp := response.PrizeStructureResponse{
		ID:          output.PrizeStructure.ID.String(),
		Name:        output.PrizeStructure.Name,
		Description: output.PrizeStructure.Description,
		IsActive:    output.PrizeStructure.IsActive,
		ValidFrom:   output.PrizeStructure.ValidFrom.Format("2006-01-02"),
		ValidTo:     validToStr,
		Prizes:      prizes,
		CreatedAt:   output.PrizeStructure.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}
