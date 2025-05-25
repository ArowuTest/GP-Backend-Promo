package response

// This file contains additional prize-related response types
// that complement the main response.go definitions

// PrizeTierDetailResponse represents a detailed prize tier response
// with additional fields for admin views
type PrizeTierDetailResponse struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name"`
	PrizeType         string `json:"prizeType"`
	Value             string `json:"value"`
	ValueNGN          int    `json:"valueNGN,omitempty"`
	CurrencyCode      string `json:"currencyCode,omitempty"`
	Quantity          int    `json:"quantity"`
	Order             int    `json:"order"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps"`
	Description       string `json:"description,omitempty"`
	CreatedAt         string `json:"createdAt,omitempty"`
	UpdatedAt         string `json:"updatedAt,omitempty"`
}

// PrizeStructureDetailResponse represents a detailed prize structure response
// with additional fields for admin views
type PrizeStructureDetailResponse struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	IsActive       bool                    `json:"isActive"`
	ValidFrom      string                  `json:"validFrom"`
	ValidTo        string                  `json:"validTo,omitempty"`
	ApplicableDays []string                `json:"applicableDays,omitempty"`
	DayType        string                  `json:"dayType,omitempty"`
	Prizes         []PrizeTierDetailResponse `json:"prizes"`
	CreatedAt      string                  `json:"createdAt,omitempty"`
	UpdatedAt      string                  `json:"updatedAt,omitempty"`
	CreatedBy      string                  `json:"createdBy,omitempty"`
	UpdatedBy      string                  `json:"updatedBy,omitempty"`
}

// PrizeStructureSummaryResponse represents a summarized prize structure response
// for list views
type PrizeStructureSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
	ValidFrom   string `json:"validFrom"`
	ValidTo     string `json:"validTo,omitempty"`
	DayType     string `json:"dayType,omitempty"`
	PrizeCount  int    `json:"prizeCount"`
	CreatedAt   string `json:"createdAt,omitempty"`
}
