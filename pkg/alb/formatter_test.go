package alb

import (
	"strings"
	"testing"
)

func TestFormatLoadBalancers(t *testing.T) {
	// Test with empty summaries
	emptyResult := FormatLoadBalancers([]LoadBalancerSummary{})
	if emptyResult != "No load balancers found" {
		t.Errorf("Expected 'No load balancers found', got '%s'", emptyResult)
	}
	
	// Test with actual summaries
	summaries := []LoadBalancerSummary{
		{
			Name:    "test-lb",
			DNSName: "test-lb.example.com",
			TargetGroups: []TargetGroupSummary{
				{
					Name: "test-tg",
					ARN:  "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/test-tg/1234567890abcdef",
					Targets: []TargetSummary{
						{
							ID:     "i-1234567890abcdef0",
							Port:   80,
							Status: "healthy",
						},
						{
							ID:     "i-0987654321fedcba0",
							Port:   80,
							Status: "unhealthy",
							Reason: "Connection refused",
						},
					},
				},
			},
		},
	}
	
	result := FormatLoadBalancers(summaries)
	
	// Validate the output contains expected elements
	expectedElements := []string{
		"LOAD BALANCERS",
		"test-lb (test-lb.example.com)",
		"test-tg",
		"✅ i-1234567890abcdef0:80 - healthy",
		"❌ i-0987654321fedcba0:80 - unhealthy (Connection refused)",
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
		{"healthy", "✅"},
		{"unhealthy", "❌"},
		{"draining", "🔄"},
		{"unavailable", "⚠️"},
		{"initial", "🔍"},
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