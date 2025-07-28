package utils

import (
	"time"
)

// GetString extracts a string value from a map[string]interface{}
func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetFloat64 extracts a float64 value from a map[string]interface{}
func GetFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

// GetInt64 extracts an int64 value from a map[string]interface{}
func GetInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		case float32:
			return int64(val)
		}
	}
	return 0
}

// GetTime extracts a time.Time value from a map[string]interface{}
func GetTime(m map[string]interface{}, key string) time.Time {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t
			}
			// Try alternative formats
			formats := []string{
				"2006-01-02T15:04:05",
				"2006-01-02 15:04:05",
				"2006-01-02",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, s); err == nil {
					return t
				}
			}
		}
	}
	return time.Time{}
}

// ParseTradeTime converts trade hour string to time.Time
func ParseTradeTime(tradeHour string) time.Time {
	if tradeHour == "" {
		return time.Now()
	}

	// Try to parse the trade hour with different formats
	formats := []string{
		"15:04:05",
		"15:04",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, tradeHour); err == nil {
			// If only time was parsed, set it to today
			if format == "15:04:05" || format == "15:04" {
				now := time.Now()
				return time.Date(now.Year(), now.Month(), now.Day(),
					t.Hour(), t.Minute(), t.Second(), 0, now.Location())
			}
			return t
		}
	}

	// If parsing fails, return current time
	return time.Now()
}
