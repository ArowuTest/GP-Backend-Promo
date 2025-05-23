package draw

import (
	"errors"
	"github.com/google/uuid"
)

// Helper function to parse UUID from string
func parseUUID(id string) (uuid.UUID, error) {
	if id == "" {
		return uuid.Nil, errors.New("ID is required")
	}
	
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.New("invalid ID format")
	}
	
	return parsedID, nil
}
