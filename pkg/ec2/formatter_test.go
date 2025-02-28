package ec2

import (
	"strings"
	"testing"
	"time"
)

func TestFormatInstances(t *testing.T) {
	refTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		instances []InstanceSummary
		contains  []string
	}{
		{
			name:      "Empty list",
			instances: []InstanceSummary{},
			contains:  []string{"No EC2 instances found"},
		},
		{
			name: "Multiple instances sorted",
			instances: []InstanceSummary{
				{
					Name:         "z-instance",
					InstanceID:   "i-2222",
					InstanceType: "t3.medium",
					LaunchTime:   refTime.Add(-2 * time.Hour),
				},
				{
					Name:         "a-instance",
					InstanceID:   "i-1111",
					InstanceType: "t3.micro",
					LaunchTime:   refTime.Add(-1 * time.Hour),
				},
			},
			contains: []string{
				"EC2 Instances (2):",
				"a-instance (i-1111)",
				"z-instance (i-2222)",
				"t3.micro",
				"Launched: 2024-01-01 11:00:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatInstances(tt.instances)
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("Output missing expected string: %q", s)
				}
			}
		})
	}
}

func TestGetInstancesSummary(t *testing.T) {
	tests := []struct {
		name      string
		instances []InstanceSummary
		want      string
	}{
		{
			name: "Mixed states",
			instances: []InstanceSummary{
				{State: "running"},
				{State: "running"},
				{State: "stopped"},
				{State: "pending"},
			},
			want: "4 total (2 running, 1 stopped, 1 other)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInstancesSummary(tt.instances); got != tt.want {
				t.Errorf("GetInstancesSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatUptime(t *testing.T) {
	refTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	
	// Save the original time.Now function and restore it after the test
	originalNow := time.Now
	defer func() { time.Now = originalNow }()
	
	// Mock time.Now to return a fixed time
	time.Now = func() time.Time { return refTime }
	
	tests := []struct {
		name   string
		launch time.Time
		want   string
	}{
		{
			name:   "1 hour",
			launch: refTime.Add(-65 * time.Minute),
			want:   "1h 5m",
		},
		{
			name:   "Exact days",
			launch: refTime.Add(-48 * time.Hour),
			want:   "2d 0h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatUptime(tt.launch); got != tt.want {
				t.Errorf("formatUptime() = %v, want %v", got, tt.want)
			}
		})
	}
}
