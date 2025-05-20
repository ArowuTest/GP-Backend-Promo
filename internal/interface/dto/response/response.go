package response

// SuccessResponse represents a successful API response with data
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalRows  int   `json:"totalRows"`
	TotalPages int   `json:"totalPages"`
	TotalItems int64 `json:"totalItems"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FullName  string `json:"fullName"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// LoginResponse represents a login API response
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// PrizeStructureResponse represents a prize structure in API responses
type PrizeStructureResponse struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	PrizeType      string              `json:"prizeType"`
	IsActive       bool                `json:"isActive"`
	ValidFrom      string              `json:"validFrom"`
	ValidTo        string              `json:"validTo"`
	ApplicableDays []string            `json:"applicableDays"`
	PrizeTiers     []PrizeTierResponse `json:"prizeTiers"`
	CreatedAt      string              `json:"createdAt"`
	UpdatedAt      string              `json:"updatedAt"`
}

// PrizeTierResponse represents a prize tier in API responses
type PrizeTierResponse struct {
	ID                string `json:"id"`
	PrizeStructureID  string `json:"prizeStructureID"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Value             string `json:"value"`
	Order             int    `json:"order"`
	NumberOfWinners   int    `json:"numberOfWinners"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

// DrawResponse represents a draw in API responses
type DrawResponse struct {
	ID                   string           `json:"id"`
	DrawDate             string           `json:"drawDate"`
	PrizeStructureID     string           `json:"prizeStructureID"`
	Status               string           `json:"status"`
	TotalEligibleMSISDNs int              `json:"totalEligibleMSISDNs"`
	TotalEntries         int              `json:"totalEntries"`
	ExecutedByAdminID    string           `json:"executedByAdminID"`
	Winners              []WinnerResponse `json:"winners"`
	CreatedAt            string           `json:"createdAt"`
	UpdatedAt            string           `json:"updatedAt"`
}

// WinnerResponse represents a winner in API responses
type WinnerResponse struct {
	ID            string `json:"id"`
	DrawID        string `json:"drawID"`
	MSISDN        string `json:"msisdn"`
	PrizeTierID   string `json:"prizeTierID"`
	PrizeTierName string `json:"prizeTierName,omitempty"`
	PrizeValue    string `json:"prizeValue,omitempty"`
	Status        string `json:"status"`
	PaymentStatus string `json:"paymentStatus,omitempty"`
	PaymentNotes  string `json:"paymentNotes,omitempty"`
	PaidAt        string `json:"paidAt,omitempty"`
	IsRunnerUp    bool   `json:"isRunnerUp"`
	RunnerUpRank  int    `json:"runnerUpRank"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// RunnerUpResponse represents a runner-up invocation response
type RunnerUpResponse struct {
	Message        string         `json:"message"`
	OriginalWinner WinnerResponse `json:"originalWinner"`
	NewWinner      WinnerResponse `json:"newWinner"`
}

// ParticipantResponse represents a participant in API responses
type ParticipantResponse struct {
	ID        string `json:"id"`
	MSISDN    string `json:"msisdn"`
	Points    int    `json:"points"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// DataUploadAuditResponse represents a data upload audit in API responses
type DataUploadAuditResponse struct {
	ID                  string `json:"id"`
	UploadedBy          string `json:"uploadedBy"`
	UploadedAt          string `json:"uploadedAt"`
	FileName            string `json:"fileName"`
	TotalUploaded       int    `json:"totalUploaded"`
	SuccessfullyImported int    `json:"successfullyImported"`
	DuplicatesSkipped   int    `json:"duplicatesSkipped"`
	ErrorsEncountered   int    `json:"errorsEncountered"`
	Status              string `json:"status"`
	Details             string `json:"details,omitempty"`
	OperationType       string `json:"operationType"`
}

// AuditLogResponse represents an audit log in API responses
type AuditLogResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"userID"`
	Username   string `json:"username"`
	Action     string `json:"action"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityID"`
	Summary    string `json:"summary,omitempty"`
	Details    string `json:"details,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

// EligibilityStatsResponse represents eligibility statistics in API responses
type EligibilityStatsResponse struct {
	Date                 string `json:"date"`
	TotalEligibleMSISDNs int    `json:"totalEligibleMSISDNs"`
	TotalEntries         int    `json:"totalEntries"`
}
