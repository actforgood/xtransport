package broker

import (
	"strconv"
	"time"
)

// Props represents a set of properties that can be associated with a consumer or publisher.
type Props map[string]any

// Get retrieves the value of the specified property name.
// If the property is not found, it returns nil.
func (p Props) Get(propName string) any {
	if value, found := p[propName]; found {
		return value
	}

	return nil
}

// GetString retrieves the string value of the specified property name.
// If the property is not found or cannot be converted to a string, it returns an empty string.
func (p Props) GetString(propName string) string {
	if value, found := p[propName]; found {
		if valStr, ok := value.(string); ok {
			return valStr
		}
	}

	return ""
}

// GetBool retrieves the boolean value of the specified property name.
// If the property is not found or cannot be converted to a boolean, it returns false.
func (p Props) GetBool(propName string) bool {
	if value, found := p[propName]; found {
		if valBool, ok := value.(bool); ok {
			return valBool
		}
	}

	return false
}

// GetInt retrieves the integer value of the specified property name.
// If the property is not found or cannot be converted to an integer, it returns 0.
func (p Props) GetInt(propName string) int {
	if value, found := p[propName]; found {
		if valInt, ok := value.(int); ok {
			return valInt
		}
		if valInt64, ok := value.(int64); ok {
			return int(valInt64)
		}
		if valUint8, ok := value.(uint8); ok {
			return int(valUint8)
		}
		if valStr, ok := value.(string); ok {
			if valInt64, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				return int(valInt64)
			}
		}
	}

	return 0
}

// GetTime retrieves the time value of the specified property name.
// If the property is not found or cannot be converted to a time, it returns the zero value of time.Time.
func (p Props) GetTime(propName string) time.Time {
	if value, found := p[propName]; found {
		if value, ok := value.(time.Time); ok {
			return value
		}
	}

	return time.Time{}
}

// GetDuration retrieves the duration value of the specified property name.
// If the property is not found or cannot be converted to a duration, it returns 0.
func (p Props) GetDuration(propName string) time.Duration {
	if value, found := p[propName]; found {
		if value, ok := value.(time.Duration); ok {
			return value
		}
		if valueStr, ok := value.(string); ok {
			if value, err := time.ParseDuration(valueStr); err == nil {
				return value
			}
		}
	}

	return 0
}
