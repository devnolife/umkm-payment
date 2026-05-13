package utils

import "fmt"

// GenerateOrderNumber returns a human-readable, time-sortable order number.
// Format: ORD-YYYYMMDD-XXXXXX where X is the last 6 of a CUID.
func GenerateOrderNumber(date string) string {
	id := NewID()
	suffix := id
	if len(id) > 6 {
		suffix = id[len(id)-6:]
	}
	return fmt.Sprintf("ORD-%s-%s", date, suffix)
}
