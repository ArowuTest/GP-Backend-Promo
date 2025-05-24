package adapter

import (
	"context"

	"github.com/ArowuTest/GP-Backend-Promo/internal/application/audit"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/prize"
	"github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
)

// SetupAdapter initializes all adapter services
type SetupAdapter struct {
	DrawAdapter        *DrawServiceAdapter
	AuditAdapter       *AuditServiceAdapter
	PrizeAdapter       *PrizeServiceAdapter
	ParticipantAdapter *ParticipantServiceAdapter
	UserAdapter        *UserServiceAdapter
}

// NewSetupAdapter creates a new SetupAdapter
func NewSetupAdapter(
	// Draw services
	drawService *draw.ExecuteDrawService,
	getDrawByIDService *draw.GetDrawByIDService,
	listDrawsService *draw.ListDrawsService,
	eligibilityService *draw.GetEligibilityStatsService,
	invokeRunnerUpService *draw.InvokeRunnerUpService,
	updateWinnerService *draw.UpdateWinnerPaymentStatusService,
	listWinnersService *draw.ListWinnersService,

	// Audit services
	auditService *audit.AuditService,
	getAuditLogsService *audit.GetAuditLogsService,

	// Prize services
	createPrizeStructureService *prize.CreatePrizeStructureService,
	getPrizeStructureService *prize.GetPrizeStructureService,
	listPrizeStructuresService *prize.ListPrizeStructuresService,
	updatePrizeStructureService *prize.UpdatePrizeStructureService,

	// Participant services
	uploadParticipantsService *participant.UploadParticipantsService,
	listParticipantsService *participant.ListParticipantsService,
	deleteUploadService *participant.DeleteUploadService,
	getParticipantStatsService *participant.GetParticipantStatsService,
	listUploadAuditsService *participant.ListUploadAuditsService,

	// User services
	authenticateUserService *user.AuthenticateUserService,
	createUserService *user.CreateUserService,
	getUserService *user.GetUserService,
	listUsersService *user.ListUsersService,
	updateUserService *user.UpdateUserService,
	resetPasswordService *user.ResetPasswordService,
) *SetupAdapter {
	// Create draw adapter
	drawAdapter := NewDrawServiceAdapter(
		drawService,
		getDrawByIDService,
		listDrawsService,
		eligibilityService,
		invokeRunnerUpService,
		updateWinnerService,
		listWinnersService,
	)

	// Create audit adapter
	auditAdapter := NewAuditServiceAdapter(
		auditService,
		*getAuditLogsService, // Dereference to get the interface value
	)

	// Create prize adapter
	prizeAdapter := NewPrizeServiceAdapter(
		createPrizeStructureService,
		getPrizeStructureService,
		listPrizeStructuresService,
		updatePrizeStructureService,
		nil, // Add nil for deletePrizeStructureService to match constructor signature
	)

	// Create participant adapter with interface{} types to resolve type mismatches
	participantAdapter := NewParticipantServiceAdapter(
		uploadParticipantsService,
		getParticipantStatsService,
		listUploadAuditsService,
		listParticipantsService,
		deleteUploadService,
	)

	// Create user adapter
	userAdapter := NewUserServiceAdapter(
		authenticateUserService,
		createUserService,
		getUserService,
		listUsersService,
		updateUserService,
	)
	
	return &SetupAdapter{
		DrawAdapter:        drawAdapter,
		AuditAdapter:       auditAdapter,
		PrizeAdapter:       prizeAdapter,
		ParticipantAdapter: participantAdapter,
		UserAdapter:        userAdapter,
	}
}

// SetupRoutes initializes all routes with the adapter services
func (s *SetupAdapter) SetupRoutes(ctx context.Context) {
	// This function would be implemented to set up all routes
	// using the adapter services. For now, it's a placeholder.
}
