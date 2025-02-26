package common

import (
	"strings"
	"testing"
)

func TestGenerateSparkline(t *testing.T) {
	// Test with empty data
	emptyResult := GenerateSparkline([]float64{}, "Test", 5)
	if emptyResult != "No data available" {
		t.Errorf("Expected 'No data available', got '%s'", emptyResult)
	}
	
	// Test with actual data
	data := []float64{1, 2, 3, 4, 5}
	result := GenerateSparkline(data, "Test Label", 3)
	
	// Check that the result contains the label
	if !strings.Contains(result, "Test Label") {
		t.Errorf("Expected sparkline to contain label 'Test Label', but it didn't")
	}
	
	// Check that the result is not empty
	if len(result) < 10 {
		t.Errorf("Expected non-empty sparkline, got '%s'", result)
	}
	
	// Test with zero height (should use default)
	result = GenerateSparkline(data, "Test", 0)
	if len(result) < 10 {
		t.Errorf("Expected non-empty sparkline with default height, got '%s'", result)
	}
}

func TestFormatPercentage(t *testing.T) {
	testCases := []struct {
		value    float64
		expected string
	}{
		{50.123, "50.12%"},
		{0, "0.00%"},
		{100, "100.00%"},
		{-10.5, "-10.50%"},
	}
	
	for _, tc := range testCases {
		t.Run(FormatFloat(tc.value), func(t *testing.T) {
			result := FormatPercentage(tc.value)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestFormatFloat(t *testing.T) {
	testCases := []struct {
		value    float64
		expected string
	}{
		{50.123, "50.12"},
		{0, "0.00"},
		{100, "100.00"},
		{-10.5, "-10.50"},
	}
	
	for _, tc := range testCases {
		t.Run(FormatFloat(tc.value), func(t *testing.T) {
			result := FormatFloat(tc.value)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestFormatFloatWithPrecision(t *testing.T) {
	testCases := []struct {
		value     float64
		precision int
		expected  string
	}{
		{50.123, 1, "50.1"},
		{50.123, 2, "50.12"},
		{50.123, 3, "50.123"},
		{50.123, 4, "50.1230"},
	}
	
	for _, tc := range testCases {
		t.Run(FormatFloat(tc.value), func(t *testing.T) {
			result := FormatFloatWithPrecision(tc.value, tc.precision)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}