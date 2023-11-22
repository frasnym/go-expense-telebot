package common

import (
	"fmt"
	"time"
)

// Helper function to parse the date string into a time.Time object
func ParseSpendeeDate(dateString string) time.Time {
	date, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		fmt.Println("Error parsing date:", err)
	}
	return date
}
