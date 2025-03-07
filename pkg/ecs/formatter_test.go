package ecs

import (
	"strings"
	"testing"
	"time"
)

func TestGetServicesSummary(t *testing.T) {
	tests := []struct {
		name     string
		services []ServiceSummary
		want     string
	}{
		{
			name:     "Empty services",
			services: []ServiceSummary{},
			want:     "0 services in 0 clusters (0 active, 0 draining, 0 other, 0/0 healthy)",
		},
		{
			name: "Mixed service states",
			services: []ServiceSummary{
				{
					ServiceName:  "service-1",
					ClusterName:  "cluster-1",
					Status:       "ACTIVE",
					DesiredCount: 2,
					RunningCount: 2,
				},
				{
					ServiceName:  "service-2",
					ClusterName:  "cluster-1",
					Status:       "ACTIVE",
					DesiredCount: 3,
					RunningCount: 1,
				},
				{
					ServiceName:  "service-3",
					ClusterName:  "cluster-2",
					Status:       "DRAINING",
					DesiredCount: 1,
					RunningCount: 0,
				},
				{
					ServiceName:  "service-4",
					ClusterName:  "cluster-3",
					Status:       "INACTIVE",
					DesiredCount: 0,
					RunningCount: 0,
				},
			},
			want: "4 services in 3 clusters (2 active, 1 draining, 1 other, 1/4 healthy)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetServicesSummary(tt.services)
			if got != tt.want {
				t.Errorf("GetServicesSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatServices(t *testing.T) {
	refTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Backup and restore the original timeNow
	oldTimeNow := timeNow
	defer func() { timeNow = oldTimeNow }()

	// Set fixed time for test
	timeNow = func() time.Time { return refTime.Add(24 * time.Hour) } // 1 day after refTime

	tests := []struct {
		name        string
		services    []ServiceSummary
		contains    []string
		notContains []string
	}{
		{
			name:     "Empty services",
			services: []ServiceSummary{},
			contains: []string{"No ECS services found."},
		},
		{
			name: "Multiple services across clusters",
			services: []ServiceSummary{
				{
					ServiceName:      "api-service",
					ClusterName:      "production",
					Status:           "ACTIVE",
					DesiredCount:     2,
					RunningCount:     2,
					TaskDefinition:   "api-task:1",
					LaunchType:       "FARGATE",
					CreatedAt:        refTime,
					DeploymentStatus: "stable",
					NetworkMode:      "awsvpc",
					Tags: map[string]string{
						"Environment": "production",
						"Project":     "demo",
					},
					LoadBalancers: []string{"api-tg"},
				},
				{
					ServiceName:      "worker-service",
					ClusterName:      "production",
					Status:           "ACTIVE",
					DesiredCount:     3,
					RunningCount:     2,
					PendingCount:     1,
					TaskDefinition:   "worker-task:1",
					LaunchType:       "EC2",
					CreatedAt:        refTime.Add(-48 * time.Hour),
					DeploymentStatus: "in-progress",
					NetworkMode:      "bridge",
				},
				{
					ServiceName:      "staging-api",
					ClusterName:      "staging",
					Status:           "ACTIVE",
					DesiredCount:     1,
					RunningCount:     0,
					TaskDefinition:   "api-task:1",
					LaunchType:       "FARGATE",
					CreatedAt:        refTime,
					DeploymentStatus: "stable",
					NetworkMode:      "awsvpc",
				},
			},
			contains: []string{
				"ECS Services (3):",
				"🚀 Cluster: production",
				"🚀 Cluster: staging",
				"🟢 api-service",
				"🟠 worker-service",
				"🔴 staging-api",
				"Status: ACTIVE (deployment: in-progress)",
				"Tasks: 2/3 running (1 pending)",
				"Load Balancers: api-tg",
				"Tags: Environment: production | Project: demo",
				"Created: 2024-01-01 12:00:00 (1d 0h ago)",
			},
			notContains: []string{
				"Status: ACTIVE (deployment: stable)", // We don't show stable deployments with status
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatServices(tt.services)

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("FormatServices() missing expected string: %q", s)
				}
			}

			for _, s := range tt.notContains {
				if strings.Contains(got, s) {
					t.Errorf("FormatServices() contains unexpected string: %q", s)
				}
			}
		})
	}
}

func TestFormatUptime(t *testing.T) {
	refTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Backup and restore the original timeNow
	oldTimeNow := timeNow
	defer func() { timeNow = oldTimeNow }()

	// Set fixed time for test
	timeNow = func() time.Time { return refTime }

	tests := []struct {
		name    string
		created time.Time
		want    string
	}{
		{
			name:    "Minutes",
			created: refTime.Add(-30 * time.Minute),
			want:    "30m",
		},
		{
			name:    "Hours and minutes",
			created: refTime.Add(-90 * time.Minute),
			want:    "1h 30m",
		},
		{
			name:    "Days and hours",
			created: refTime.Add(-25 * time.Hour),
			want:    "1d 1h",
		},
		{
			name:    "Months and days",
			created: refTime.Add(-40 * 24 * time.Hour),
			want:    "1m 10d",
		},
		{
			name:    "Years and months",
			created: refTime.Add(-380 * 24 * time.Hour),
			want:    "1y 0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUptime(tt.created)
			if got != tt.want {
				t.Errorf("formatUptime() = %v, want %v", got, tt.want)
			}
		})
	}
}
