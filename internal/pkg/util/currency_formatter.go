package util

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatCurrency formats a float64 value as a currency string with the given currency code
// Example: FormatCurrency(5000.00, "N") returns "N5,000.00"
func FormatCurrency(value float64, currencyCode string) string {
	// Convert the float to a string with 2 decimal places
	valueStr := strconv.FormatFloat(value, 'f', 2, 64)
	
	// Split the string into integer and decimal parts
	parts := strings.Split(valueStr, ".")
	intPart := parts[0]
	decPart := parts[1]
	
	// Format the integer part with commas as thousand separators
	var formattedInt string
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			formattedInt += ","
		}
		formattedInt += string(c)
	}
	
	// Combine the parts with the currency code
	return fmt.Sprintf("%s%s.%s", currencyCode, formattedInt, decPart)
}

// ParseCurrency parses a currency string into a float64 value
// Example: ParseCurrency("N5,000.00") returns 5000.00
func ParseCurrency(currencyStr string) (float64, error) {
	// Remove any currency code (assuming it's at the beginning and non-numeric)
	cleanStr := currencyStr
	for i, c := range currencyStr {
		if c >= '0' && c <= '9' || c == '.' || c == '-' {
			cleanStr = currencyStr[i:]
			break
		}
	}
	
	// Remove commas
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")
	
	// Parse the string to float64
	return strconv.ParseFloat(cleanStr, 64)
}
