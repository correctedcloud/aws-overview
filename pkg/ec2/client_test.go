package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type mockEC2API struct {
	DescribeInstancesFunc func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func (m *mockEC2API) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.DescribeInstancesFunc(ctx, params, optFns...)
}

func TestGetInstances(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		mockResponses []*ec2.DescribeInstancesOutput
		mockError     error
		wantCount     int
		wantErr       bool
	}{
		{
			name: "Success with multiple pages",
			mockResponses: []*ec2.DescribeInstancesOutput{
				{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{InstanceId: ptrString("i-12345"), State: &types.InstanceState{Name: types.InstanceStateNameRunning}},
							},
						},
					},
					NextToken: ptrString("token"),
				},
				{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{InstanceId: ptrString("i-67890"), State: &types.InstanceState{Name: types.InstanceStateNameStopped}},
							},
						},
					},
				},
			},
			wantCount: 2,
		},
		{
			name: "Filter terminated instances",
			mockResponses: []*ec2.DescribeInstancesOutput{
				{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{InstanceId: ptrString("i-12345"), State: &types.InstanceState{Name: types.InstanceStateNameTerminated}},
								{InstanceId: ptrString("i-67890"), State: &types.InstanceState{Name: types.InstanceStateNameRunning}},
							},
						},
					},
				},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			client := NewClient(&mockEC2API{
				DescribeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					resp := tt.mockResponses[callCount]
					callCount++
					return resp, nil
				},
			})

			got, err := client.GetInstances(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetInstances() count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestGetPlatform(t *testing.T) {
	tests := []struct {
		name     string
		instance types.Instance
		want     string
	}{
		{
			name:     "Explicit Windows platform",
			instance: types.Instance{Platform: types.PlatformValuesWindows},
			want:     "Windows",
		},
		{
			name:     "Linux from details",
			instance: types.Instance{PlatformDetails: ptrString("Linux/UNIX")},
			want:     "Linux",
		},
		{
			name:     "Unknown platform",
			instance: types.Instance{PlatformDetails: ptrString("Other")},
			want:     "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPlatform(tt.instance); got != tt.want {
				t.Errorf("getPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}
