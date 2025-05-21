package util

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

// ParseTimeOrZero converts a string to time.Time using the specified format
// Returns zero time if the string is empty or parsing fails
func ParseTimeOrZero(dateStr string, format string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse(format, dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// FormatTimeOrEmpty converts a time.Time to string using the specified format
// Returns empty string if the time is zero
func FormatTimeOrEmpty(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(format)
}

// FormatFloat converts a float64 to string
func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ParseFloatOrZero converts a string to float64
// Returns 0 if the string is empty or parsing fails
func ParseFloatOrZero(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ParseUUIDOrNil converts a string to uuid.UUID
// Returns nil UUID if the string is empty or parsing fails
func ParseUUIDOrNil(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// FormatUUID converts a uuid.UUID to string
// Returns empty string if the UUID is nil
func FormatUUID(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}
