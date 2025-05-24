package adapter

import (
	"context"

	userApp "github.com/ArowuTest/GP-Backend-Promo/internal/application/user"
	"github.com/ArowuTest/GP-Backend-Promo/internal/domain/entity"
)

// UserHandlerAdapter adapts the user service adapter to match the handler's expected interface
type UserHandlerAdapter struct {
	userServiceAdapter *UserServiceAdapter
}

// NewUserHandlerAdapter creates a new UserHandlerAdapter
func NewUserHandlerAdapter(
	userServiceAdapter *UserServiceAdapter,
) *UserHandlerAdapter {
	return &UserHandlerAdapter{
		userServiceAdapter: userServiceAdapter,
	}
}

// GetUserByIDService provides the missing service required by the handler
type GetUserByIDService struct {
	userServiceAdapter *UserServiceAdapter
}

// NewGetUserByIDService creates a new GetUserByIDService
func NewGetUserByIDService(
	userServiceAdapter *UserServiceAdapter,
) *GetUserByIDService {
	return &GetUserByIDService{
		userServiceAdapter: userServiceAdapter,
	}
}

// GetUser implements the service method required by the handler
func (s *GetUserByIDService) GetUser(
	ctx context.Context,
	input userApp.GetUserInput,
) (*entity.User, error) {
	// Use GetUser instead of GetUserByID to match the adapter's method name
	output, err := s.userServiceAdapter.GetUser(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &output.User, nil
}
