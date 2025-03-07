package common

import (
	"fmt"
	"github.com/guptarohit/asciigraph"
)

// GenerateSparkline creates a simple ASCII sparkline from data points
func GenerateSparkline(data []float64, label string, height int) string {
	if len(data) == 0 {
		return "No data available"
	}

	if height <= 0 {
		height = 5 // Default height
	}

	return asciigraph.Plot(
		data,
		asciigraph.Height(height),
		asciigraph.Caption(label),
	)
}

// FormatPercentage formats a value as a percentage string
func FormatPercentage(value float64) string {
	return FormatFloat(value) + "%"
}

// FormatFloat formats a float value to 2 decimal places
func FormatFloat(value float64) string {
	return FormatFloatWithPrecision(value, 2)
}

// FormatFloatWithPrecision formats a float value with the specified precision
func FormatFloatWithPrecision(value float64, precision int) string {
	format := "%." + fmt.Sprintf("%d", precision) + "f"
	return fmt.Sprintf(format, value)
}
