package db

import (
	"fmt"
	"strings"
)

// Format is a method on Currency that formats a float64 value as a string
func (c *Currency) Format(value float64) string {
	format := fmt.Sprintf("%%.%df %%s", c.DecimalPlaces)
	return fmt.Sprintf(format, value, c.Code)
}

func (c *Currency) GetSupportedPaymentSchemes() []string {

	if c.SupportedPaymentSchemes != "" {
		parts := strings.Split(c.SupportedPaymentSchemes, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}

	return []string{}
}
