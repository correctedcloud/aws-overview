package rds

import (
	"strings"
	"testing"
)

func TestFormatDBInstances(t *testing.T) {
	// Test with empty summaries
	emptyResult := FormatDBInstances([]DBInstanceSummary{})
	if emptyResult != "No DB instances found" {
		t.Errorf("Expected 'No DB instances found', got '%s'", emptyResult)
	}

	// Test with actual summaries
	summaries := []DBInstanceSummary{
		{
			Identifier:   "test-db",
			Engine:       "postgres",
			Status:       "available",
			Endpoint:     "test-db.xyz123.us-east-1.rds.amazonaws.com:5432",
			CPUData:      []float64{10.0, 15.0, 12.0, 8.0},
			MemoryData:   []float64{45.0, 48.0, 50.0, 47.0},
			RecentErrors: []string{},
		},
		{
			Identifier:   "test-db-2",
			Engine:       "mysql",
			Status:       "stopped",
			Endpoint:     "test-db-2.xyz123.us-east-1.rds.amazonaws.com:3306",
			CPUData:      []float64{},
			MemoryData:   []float64{},
			RecentErrors: []string{"Error detected at 2023-01-01 12:00:00: Out of memory"},
		},
	}

	result := FormatDBInstances(summaries)

	// Validate the output contains expected elements
	expectedElements := []string{
		"RDS INSTANCES",
		"✅ test-db (postgres)",
		"Endpoint: test-db.xyz123.us-east-1.rds.amazonaws.com:5432",
		"CPU Utilization (1 hour):",
		"Memory Utilization (1 hour):",
		"No recent errors",
		"⏹️ test-db-2 (mysql)",
		"Endpoint: test-db-2.xyz123.us-east-1.rds.amazonaws.com:3306",
		"No CPU data available",
		"No memory data available",
		"Error detected at 2023-01-01 12:00:00: Out of memory",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't", expected)
		}
	}
}

func TestGetStatusSymbol(t *testing.T) {
	testCases := []struct {
		status   string
		expected string
	}{
		{"available", "✅"},
		{"creating", "🔄"},
		{"deleting", "🗑️"},
		{"failed", "❌"},
		{"maintenance", "🔧"},
		{"modifying", "🔄"},
		{"stopped", "⏹️"},
		{"storage-full", "💾"},
		{"unknown", "❓"},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			symbol := getStatusSymbol(tc.status)
			if symbol != tc.expected {
				t.Errorf("Expected symbol '%s' for status '%s', got '%s'", tc.expected, tc.status, symbol)
			}
		})
	}
}
