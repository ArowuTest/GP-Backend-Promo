package draw

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	drawDomain "github.com/ArowuTest/GP-Backend-Promo/internal/domain/draw"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/participant"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/prize"
)

// Repository defines the interface for draw repository
type Repository interface {
	Create(draw *drawDomain.Draw) error
	GetByID(id uuid.UUID) (*drawDomain.Draw, error)
	List(page, pageSize int) ([]drawDomain.Draw, int, error)
	GetByDate(date time.Time) (*drawDomain.Draw, error)
	Update(draw *drawDomain.Draw) error
	GetEligibilityStats(date time.Time) (int, int, error)
	CreateWinner(winner *drawDomain.Winner) error
	GetWinnerByID(id uuid.UUID) (*drawDomain.Winner, error)
	UpdateWinner(winner *drawDomain.Winner) error
	GetRunnerUps(drawID uuid.UUID, prizeTierID uuid.UUID, limit int) ([]drawDomain.Winner, error)
	ListWinners(ctx context.Context, page, pageSize int, startDate, endDate string) ([]*drawDomain.Winner, int, error)
	ExecuteDraw(drawDate time.Time, prizeStructureID uuid.UUID, executedByAdminID uuid.UUID, eligibleParticipants []participant.Participant, prizeTiers []prize.PrizeTier) (*drawDomain.Draw, error)
}
