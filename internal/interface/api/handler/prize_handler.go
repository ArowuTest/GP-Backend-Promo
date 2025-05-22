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
	createPrizeService          *prizeApp.CreatePrizeService
	getPrizeByIDService         *prizeApp.GetPrizeByIDService
	listPrizesService           *prizeApp.ListPrizesService
	createPrizeStructureService *prizeApp.CreatePrizeStructureService
	getPrizeStructureByIDService *prizeApp.GetPrizeStructureByIDService
	listPrizeStructuresService  *prizeApp.ListPrizeStructuresService
	addPrizeTierService         *prizeApp.AddPrizeTierService
}

// NewPrizeHandler creates a new PrizeHandler
func NewPrizeHandler(
	createPrizeService *prizeApp.CreatePrizeService,
	getPrizeByIDService *prizeApp.GetPrizeByIDService,
	listPrizesService *prizeApp.ListPrizesService,
	createPrizeStructureService *prizeApp.CreatePrizeStructureService,
	getPrizeStructureByIDService *prizeApp.GetPrizeStructureByIDService,
	listPrizeStructuresService *prizeApp.ListPrizeStructuresService,
	addPrizeTierService *prizeApp.AddPrizeTierService,
) *PrizeHandler {
	return &PrizeHandler{
		createPrizeService:          createPrizeService,
		getPrizeByIDService:         getPrizeByIDService,
		listPrizesService:           listPrizesService,
		createPrizeStructureService: createPrizeStructureService,
		getPrizeStructureByIDService: getPrizeStructureByIDService,
		listPrizeStructuresService:  listPrizeStructuresService,
		addPrizeTierService:         addPrizeTierService,
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
	
	// Prepare input - using only fields that exist in PrizeInput
	input := prizeApp.PrizeInput{
		Name:        req.Name,
		Description: req.Description,
		Value:       req.ValueNGN, // Map ValueNGN to Value
		Quantity:    req.Quantity,
		// PrizeType and Order fields don't exist in PrizeInput, so they're omitted
	}
	
	// Create prize
	prize, err := h.createPrizeService.CreatePrize(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to create prize: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeTierResponse{
			ID:          prize.ID.String(),
			Name:        prize.Name,
			Description: prize.Description,
			PrizeType:   "Cash", // Default value since it's not in the domain model
			ValueNGN:    prize.Value,
			Quantity:    prize.Quantity,
			Order:       0, // Default value since it's not in the domain model
			CreatedAt:   util.FormatTimeOrEmpty(prize.CreatedAt, time.RFC3339),
			UpdatedAt:   util.FormatTimeOrEmpty(prize.UpdatedAt, time.RFC3339),
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
	
	// Get prize
	prize, err := h.getPrizeByIDService.GetPrizeByID(c.Request.Context(), prizeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeTierResponse{
			ID:          prize.ID.String(),
			Name:        prize.Name,
			Description: prize.Description,
			PrizeType:   "Cash", // Default value since it's not in the domain model
			ValueNGN:    prize.Value,
			Quantity:    prize.Quantity,
			Order:       0, // Default value since it's not in the domain model
			CreatedAt:   util.FormatTimeOrEmpty(prize.CreatedAt, time.RFC3339),
			UpdatedAt:   util.FormatTimeOrEmpty(prize.UpdatedAt, time.RFC3339),
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
	
	// List prizes
	prizes, totalCount, err := h.listPrizesService.ListPrizes(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list prizes: " + err.Error(),
		})
		return
	}
	
	// Calculate total pages
	totalPages := (totalCount + pageSize - 1) / pageSize
	
	// Prepare response
	prizeTiers := make([]response.PrizeTierResponse, 0, len(prizes))
	for _, p := range prizes {
		prizeTiers = append(prizeTiers, response.PrizeTierResponse{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description,
			PrizeType:   "Cash", // Default value since it's not in the domain model
			ValueNGN:    p.Value,
			Quantity:    p.Quantity,
			Order:       0, // Default value since it's not in the domain model
			CreatedAt:   util.FormatTimeOrEmpty(p.CreatedAt, time.RFC3339),
			UpdatedAt:   util.FormatTimeOrEmpty(p.UpdatedAt, time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    prizeTiers,
		Pagination: response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalRows:  totalCount,
			TotalPages: totalPages,
			TotalItems: int64(totalCount),
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
	
	// Parse date strings to time.Time
	validFrom := util.ParseTimeOrZero(req.ValidFrom, "2006-01-02")
	if validFrom.IsZero() {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid valid from date format. Expected YYYY-MM-DD.",
		})
		return
	}
	
	validTo := util.ParseTimeOrZero(req.ValidTo, "2006-01-02")
	if validTo.IsZero() {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid valid to date format. Expected YYYY-MM-DD.",
		})
		return
	}
	
	// Prepare input
	input := prizeApp.PrizeStructureInput{
		Name:        req.Name,
		Description: req.Description,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		IsActive:    req.IsActive,
	}
	
	// Create prize structure
	prizeStructure, err := h.createPrizeStructureService.CreatePrizeStructure(c.Request.Context(), input)
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
			ID:          prizeStructure.ID.String(),
			Name:        prizeStructure.Name,
			Description: prizeStructure.Description,
			ValidFrom:   util.FormatTimeOrEmpty(prizeStructure.ValidFrom, "2006-01-02"),
			ValidTo:     util.FormatTimeOrEmpty(prizeStructure.ValidTo, "2006-01-02"),
			IsActive:    prizeStructure.IsActive,
			PrizeTiers:  []response.PrizeTierResponse{}, // Empty slice since we just created it
			CreatedAt:   util.FormatTimeOrEmpty(prizeStructure.CreatedAt, time.RFC3339),
			UpdatedAt:   util.FormatTimeOrEmpty(prizeStructure.UpdatedAt, time.RFC3339),
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
	prizeStructure, err := h.getPrizeStructureByIDService.GetPrizeStructureByID(c.Request.Context(), prizeStructureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to get prize structure: " + err.Error(),
		})
		return
	}
	
	// Prepare response
	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: response.PrizeStructureResponse{
			ID:          prizeStructure.ID.String(),
			Name:        prizeStructure.Name,
			Description: prizeStructure.Description,
			ValidFrom:   util.FormatTimeOrEmpty(prizeStructure.ValidFrom, "2006-01-02"),
			ValidTo:     util.FormatTimeOrEmpty(prizeStructure.ValidTo, "2006-01-02"),
			IsActive:    prizeStructure.IsActive,
			PrizeTiers:  []response.PrizeTierResponse{}, // Mock empty slice since we don't have prize tiers in the domain model
			CreatedAt:   util.FormatTimeOrEmpty(prizeStructure.CreatedAt, time.RFC3339),
			UpdatedAt:   util.FormatTimeOrEmpty(prizeStructure.UpdatedAt, time.RFC3339),
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
	prizeStructures, totalCount, err := h.listPrizeStructuresService.ListPrizeStructures(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Success: false,
			Error:   "Failed to list prize structures: " + err.Error(),
		})
		return
	}
	
	// Calculate total pages
	totalPages := (totalCount + pageSize - 1) / pageSize
	
	// Prepare response
	structures := make([]response.PrizeStructureResponse, 0, len(prizeStructures))
	for _, ps := range prizeStructures {
		structures = append(structures, response.PrizeStructureResponse{
			ID:          ps.ID.String(),
			Name:        ps.Name,
			Description: ps.Description,
			ValidFrom:   util.FormatTimeOrEmpty(ps.ValidFrom, "2006-01-02"),
			ValidTo:     util.FormatTimeOrEmpty(ps.ValidTo, "2006-01-02"),
			IsActive:    ps.IsActive,
			PrizeTiers:  []response.PrizeTierResponse{}, // Mock empty slice since we don't have prize tiers in the domain model
			CreatedAt:   util.FormatTimeOrEmpty(ps.CreatedAt, time.RFC3339),
			UpdatedAt:   util.FormatTimeOrEmpty(ps.UpdatedAt, time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success: true,
		Data:    structures,
		Pagination: response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalRows:  totalCount,
			TotalPages: totalPages,
			TotalItems: int64(totalCount),
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
	
	var req request.AddPrizeTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Parse prize tier ID
	prizeTierID, err := uuid.Parse(req.PrizeTierID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Success: false,
			Error:   "Invalid prize tier ID format",
		})
		return
	}
	
	// Prepare input
	input := prizeApp.AddPrizeTierInput{
		PrizeStructureID: prizeStructureID,
		PrizeTierID:      prizeTierID,
		Quantity:         req.Quantity,
	}
	
	// Add prize tier to prize structure
	err = h.addPrizeTierService.AddPrizeTier(c.Request.Context(), input)
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
