package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
)

// Repository defines the interface for draw repository
type Repository interface {
	GetDrawByID(ctx context.Context, id uuid.UUID) (*drawDomain.Draw, []*drawDomain.Winner, error)
	ListWinners(ctx context.Context, page, pageSize int, startDate, endDate string) ([]*drawDomain.Winner, int, error)
	UpdateWinnerPaymentStatus(ctx context.Context, winnerID uuid.UUID, paymentStatus, notes string, updatedBy uuid.UUID) (*drawDomain.Winner, error)
	ExecuteDraw(ctx context.Context, drawDate time.Time, prizeStructureID, executedBy uuid.UUID) (*drawDomain.Draw, []*drawDomain.Winner, error)
	InvokeRunnerUp(ctx context.Context, winnerID uuid.UUID, reason string, invokedBy uuid.UUID) (*drawDomain.Winner, *drawDomain.Winner, error)
}
