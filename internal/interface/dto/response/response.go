package response

// SuccessResponse is a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse is a generic error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// LoginResponse is the response for login
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	Role      string `json:"role"`
}

// UserResponse is the response for user
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// PrizeStructureResponse is the response for prize structure
type PrizeStructureResponse struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	IsActive       bool             `json:"isActive"`
	ValidFrom      string           `json:"validFrom"`
	ValidTo        string           `json:"validTo,omitempty"`
	ApplicableDays []string         `json:"applicableDays"`
	Prizes         []PrizeTierResponse `json:"prizes"`
	CreatedAt      string           `json:"createdAt"`
	UpdatedAt      string           `json:"updatedAt"`
}

// PrizeTierResponse is the response for prize tier
type PrizeTierResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	PrizeType         string `json:"prizeType"`
	Value             string `json:"value"`
	Quantity          int    `json:"quantity"`
	Order             int    `json:"order"`
	NumberOfRunnerUps int    `json:"numberOfRunnerUps"`
}

// DrawResponse is the response for draw
type DrawResponse struct {
	ID               string    `json:"id"`
	DrawDate         string    `json:"drawDate"`
	PrizeStructureID string    `json:"prizeStructureId"`
	Status           string    `json:"status"`
	CreatedAt        string    `json:"createdAt"`
	UpdatedAt        string    `json:"updatedAt"`
	Winners          []WinnerResponse `json:"winners,omitempty"`
}

// WinnerResponse is the response for winner
type WinnerResponse struct {
	ID               string `json:"id"`
	DrawID           string `json:"drawId"`
	PrizeTierID      string `json:"prizeTierId"`
	MSISDN           string `json:"msisdn"`
	PrizeName        string `json:"prizeName"`
	PrizeValue       string `json:"prizeValue"`
	IsRunnerUp       bool   `json:"isRunnerUp"`
	RunnerUpPosition int    `json:"runnerUpPosition"`
	Status           string `json:"status"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

// ParticipantResponse is the response for participant
type ParticipantResponse struct {
	ID             string  `json:"id"`
	MSISDN         string  `json:"msisdn"`
	RechargeAmount float64 `json:"rechargeAmount"`
	RechargeDate   string  `json:"rechargeDate"`
	Points         int     `json:"points"`
	CreatedAt      string  `json:"createdAt"`
	UploadID       string  `json:"uploadId"`
	UploadedAt     string  `json:"uploadedAt"`
}

// ParticipantStatsResponse is the response for participant stats
type ParticipantStatsResponse struct {
	Date              string `json:"date"`
	TotalParticipants int    `json:"totalParticipants"`
	TotalPoints       int    `json:"totalPoints"`
}

// UploadParticipantsResponse is the response for upload participants
type UploadParticipantsResponse struct {
	TotalUploaded        int    `json:"totalUploaded"`
	UploadID             string `json:"uploadId"`
	UploadedAt           string `json:"uploadedAt"`
	FileName             string `json:"fileName"`
	SuccessfullyImported int    `json:"successfullyImported"`
	DuplicatesSkipped    int    `json:"duplicatesSkipped"`
	ErrorsEncountered    int    `json:"errorsEncountered"`
	Status               string `json:"status"`
	Notes                string `json:"notes"`
	OperationType        string `json:"operationType"`
}

// UploadAuditResponse is the response for upload audit
type UploadAuditResponse struct {
	ID             string `json:"id"`
	UploadedBy     string `json:"uploadedBy"`
	UploadDate     string `json:"uploadDate"`
	FileName       string `json:"fileName"`
	Status         string `json:"status"`
	TotalRows      int    `json:"totalRows"`
	SuccessfulRows int    `json:"successfulRows"`
	ErrorCount     int    `json:"errorCount"`
}

// AuditLogResponse is the response for audit log
type AuditLogResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Action    string `json:"action"`
	EntityID  string `json:"entityId"`
	EntityType string `json:"entityType"`
	CreatedAt string `json:"timestamp"`
	Details   string `json:"details"`
}

// DataUploadAuditResponse is the response for data upload audit
type DataUploadAuditResponse struct {
	ID                 string `json:"id"`
	UploadedBy         string `json:"uploadedByUserId"`
	UploadedAt         string `json:"uploadTimestamp"`
	FileName           string `json:"fileName"`
	TotalUploaded      int    `json:"recordCount"`
	SuccessfullyImported int  `json:"successfullyImported"`
	DuplicatesSkipped  int    `json:"duplicatesSkipped"`
	ErrorsEncountered  int    `json:"errorsEncountered"`
	Status             string `json:"status"`
	Details            string `json:"notes"`
	OperationType      string `json:"operationType"`
}

// Pagination is the response for pagination
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalRows  int   `json:"totalRows"`
	TotalPages int   `json:"totalPages"`
	TotalItems int64 `json:"totalItems"`
}

// PaginatedResponse is a generic paginated response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}
