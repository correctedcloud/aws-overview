package ecs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type mockECSAPI struct {
	ListClustersFunc     func(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error)
	DescribeClustersFunc func(ctx context.Context, params *ecs.DescribeClustersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error)
	ListServicesFunc     func(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
	DescribeServicesFunc func(ctx context.Context, params *ecs.DescribeServicesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error)
}

func (m *mockECSAPI) ListClusters(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
	return m.ListClustersFunc(ctx, params, optFns...)
}

func (m *mockECSAPI) DescribeClusters(ctx context.Context, params *ecs.DescribeClustersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error) {
	return m.DescribeClustersFunc(ctx, params, optFns...)
}

func (m *mockECSAPI) ListServices(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
	return m.ListServicesFunc(ctx, params, optFns...)
}

func (m *mockECSAPI) DescribeServices(ctx context.Context, params *ecs.DescribeServicesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error) {
	return m.DescribeServicesFunc(ctx, params, optFns...)
}

func TestGetClusters(t *testing.T) {
	tests := []struct {
		name          string
		listResponse  *ecs.ListClustersOutput
		descResponse  *ecs.DescribeClustersOutput
		expectedCount int
		wantErr       bool
	}{
		{
			name: "Single cluster",
			listResponse: &ecs.ListClustersOutput{
				ClusterArns: []string{"arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster"},
			},
			descResponse: &ecs.DescribeClustersOutput{
				Clusters: []types.Cluster{
					{
						ClusterName:                       aws.String("test-cluster"),
						Status:                            aws.String("ACTIVE"),
						RegisteredContainerInstancesCount: 5,
					},
				},
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name: "Multiple clusters",
			listResponse: &ecs.ListClustersOutput{
				ClusterArns: []string{
					"arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster-1",
					"arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster-2",
				},
			},
			descResponse: &ecs.DescribeClustersOutput{
				Clusters: []types.Cluster{
					{
						ClusterName:                       aws.String("test-cluster-1"),
						Status:                            aws.String("ACTIVE"),
						RegisteredContainerInstancesCount: 3,
					},
					{
						ClusterName:                       aws.String("test-cluster-2"),
						Status:                            aws.String("ACTIVE"),
						RegisteredContainerInstancesCount: 2,
					},
				},
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "No clusters",
			listResponse: &ecs.ListClustersOutput{
				ClusterArns: []string{},
			},
			descResponse:  &ecs.DescribeClustersOutput{},
			expectedCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&mockECSAPI{
				ListClustersFunc: func(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
					return tt.listResponse, nil
				},
				DescribeClustersFunc: func(ctx context.Context, params *ecs.DescribeClustersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error) {
					return tt.descResponse, nil
				},
			})

			clusters, err := client.getClusters(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("getClusters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(clusters) != tt.expectedCount {
				t.Errorf("getClusters() count = %d, want %d", len(clusters), tt.expectedCount)
			}
		})
	}
}

func TestGetClusterServices(t *testing.T) {
	refTime := time.Now()

	tests := []struct {
		name             string
		clusterName      string
		listServicesResp *ecs.ListServicesOutput
		descServicesResp *ecs.DescribeServicesOutput
		expectedCount    int
		wantErr          bool
	}{
		{
			name:        "Single service",
			clusterName: "test-cluster",
			listServicesResp: &ecs.ListServicesOutput{
				ServiceArns: []string{"arn:aws:ecs:us-west-2:123456789012:service/test-cluster/test-service"},
			},
			descServicesResp: &ecs.DescribeServicesOutput{
				Services: []types.Service{
					{
						ServiceName:    aws.String("test-service"),
						Status:         aws.String("ACTIVE"),
						DesiredCount:   2,
						RunningCount:   2,
						PendingCount:   0,
						TaskDefinition: aws.String("arn:aws:ecs:us-west-2:123456789012:task-definition/task-def:1"),
						LaunchType:     types.LaunchTypeFargate,
						CreatedAt:      aws.Time(refTime.Add(-24 * time.Hour)),
						Deployments: []types.Deployment{
							{
								RolloutState: types.DeploymentRolloutStateCompleted,
							},
						},
						Tags: []types.Tag{
							{
								Key:   aws.String("Environment"),
								Value: aws.String("production"),
							},
						},
					},
				},
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:        "Multiple services",
			clusterName: "test-cluster",
			listServicesResp: &ecs.ListServicesOutput{
				ServiceArns: []string{
					"arn:aws:ecs:us-west-2:123456789012:service/test-cluster/service-1",
					"arn:aws:ecs:us-west-2:123456789012:service/test-cluster/service-2",
				},
			},
			descServicesResp: &ecs.DescribeServicesOutput{
				Services: []types.Service{
					{
						ServiceName:    aws.String("service-1"),
						Status:         aws.String("ACTIVE"),
						DesiredCount:   2,
						RunningCount:   2,
						PendingCount:   0,
						TaskDefinition: aws.String("arn:aws:ecs:us-west-2:123456789012:task-definition/task-def-1:1"),
						LaunchType:     types.LaunchTypeFargate,
						CreatedAt:      aws.Time(refTime.Add(-24 * time.Hour)),
						Deployments: []types.Deployment{
							{
								RolloutState: types.DeploymentRolloutStateCompleted,
							},
						},
					},
					{
						ServiceName:    aws.String("service-2"),
						Status:         aws.String("ACTIVE"),
						DesiredCount:   3,
						RunningCount:   2,
						PendingCount:   1,
						TaskDefinition: aws.String("arn:aws:ecs:us-west-2:123456789012:task-definition/task-def-2:1"),
						LaunchType:     types.LaunchTypeEc2,
						CreatedAt:      aws.Time(refTime.Add(-48 * time.Hour)),
						Deployments: []types.Deployment{
							{
								RolloutState: types.DeploymentRolloutStateInProgress,
							},
						},
					},
				},
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:        "No services",
			clusterName: "test-cluster",
			listServicesResp: &ecs.ListServicesOutput{
				ServiceArns: []string{},
			},
			descServicesResp: &ecs.DescribeServicesOutput{},
			expectedCount:    0,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&mockECSAPI{
				ListServicesFunc: func(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
					if params.Cluster == nil || *params.Cluster != tt.clusterName {
						t.Errorf("ListServices() called with wrong cluster name: got %v, want %v",
							aws.ToString(params.Cluster), tt.clusterName)
					}
					return tt.listServicesResp, nil
				},
				DescribeServicesFunc: func(ctx context.Context, params *ecs.DescribeServicesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error) {
					if params.Cluster == nil || *params.Cluster != tt.clusterName {
						t.Errorf("DescribeServices() called with wrong cluster name: got %v, want %v",
							aws.ToString(params.Cluster), tt.clusterName)
					}
					return tt.descServicesResp, nil
				},
			})

			services, err := client.getClusterServices(context.Background(), tt.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getClusterServices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(services) != tt.expectedCount {
				t.Errorf("getClusterServices() count = %d, want %d", len(services), tt.expectedCount)
			}
		})
	}
}

func TestGetServices(t *testing.T) {
	refTime := time.Now()

	tests := []struct {
		name                 string
		listClustersResp     *ecs.ListClustersOutput
		describeClustersResp *ecs.DescribeClustersOutput
		listServicesResp     map[string]*ecs.ListServicesOutput
		descServicesResp     map[string]*ecs.DescribeServicesOutput
		expectedCount        int
		wantErr              bool
	}{
		{
			name: "Multiple clusters with services",
			listClustersResp: &ecs.ListClustersOutput{
				ClusterArns: []string{
					"arn:aws:ecs:us-west-2:123456789012:cluster/cluster-1",
					"arn:aws:ecs:us-west-2:123456789012:cluster/cluster-2",
				},
			},
			describeClustersResp: &ecs.DescribeClustersOutput{
				Clusters: []types.Cluster{
					{
						ClusterName: aws.String("cluster-1"),
						Status:      aws.String("ACTIVE"),
					},
					{
						ClusterName: aws.String("cluster-2"),
						Status:      aws.String("ACTIVE"),
					},
				},
			},
			listServicesResp: map[string]*ecs.ListServicesOutput{
				"cluster-1": {
					ServiceArns: []string{
						"arn:aws:ecs:us-west-2:123456789012:service/cluster-1/service-1",
					},
				},
				"cluster-2": {
					ServiceArns: []string{
						"arn:aws:ecs:us-west-2:123456789012:service/cluster-2/service-2",
						"arn:aws:ecs:us-west-2:123456789012:service/cluster-2/service-3",
					},
				},
			},
			descServicesResp: map[string]*ecs.DescribeServicesOutput{
				"cluster-1": {
					Services: []types.Service{
						{
							ServiceName:    aws.String("service-1"),
							Status:         aws.String("ACTIVE"),
							DesiredCount:   2,
							RunningCount:   2,
							TaskDefinition: aws.String("arn:aws:ecs:us-west-2:123456789012:task-definition/task-def-1:1"),
							CreatedAt:      aws.Time(refTime),
							Deployments:    []types.Deployment{{RolloutState: types.DeploymentRolloutStateCompleted}},
						},
					},
				},
				"cluster-2": {
					Services: []types.Service{
						{
							ServiceName:    aws.String("service-2"),
							Status:         aws.String("ACTIVE"),
							DesiredCount:   1,
							RunningCount:   1,
							TaskDefinition: aws.String("arn:aws:ecs:us-west-2:123456789012:task-definition/task-def-2:1"),
							CreatedAt:      aws.Time(refTime),
							Deployments:    []types.Deployment{{RolloutState: types.DeploymentRolloutStateCompleted}},
						},
						{
							ServiceName:    aws.String("service-3"),
							Status:         aws.String("ACTIVE"),
							DesiredCount:   1,
							RunningCount:   1,
							TaskDefinition: aws.String("arn:aws:ecs:us-west-2:123456789012:task-definition/task-def-3:1"),
							CreatedAt:      aws.Time(refTime),
							Deployments:    []types.Deployment{{RolloutState: types.DeploymentRolloutStateCompleted}},
						},
					},
				},
			},
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name: "No clusters",
			listClustersResp: &ecs.ListClustersOutput{
				ClusterArns: []string{},
			},
			describeClustersResp: &ecs.DescribeClustersOutput{},
			listServicesResp:     map[string]*ecs.ListServicesOutput{},
			descServicesResp:     map[string]*ecs.DescribeServicesOutput{},
			expectedCount:        0,
			wantErr:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&mockECSAPI{
				ListClustersFunc: func(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
					return tt.listClustersResp, nil
				},
				DescribeClustersFunc: func(ctx context.Context, params *ecs.DescribeClustersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error) {
					return tt.describeClustersResp, nil
				},
				ListServicesFunc: func(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
					clusterName := aws.ToString(params.Cluster)
					if resp, ok := tt.listServicesResp[clusterName]; ok {
						return resp, nil
					}
					t.Fatalf("Unexpected ListServices call for cluster: %s", clusterName)
					return nil, nil
				},
				DescribeServicesFunc: func(ctx context.Context, params *ecs.DescribeServicesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error) {
					clusterName := aws.ToString(params.Cluster)
					if resp, ok := tt.descServicesResp[clusterName]; ok {
						return resp, nil
					}
					t.Fatalf("Unexpected DescribeServices call for cluster: %s", clusterName)
					return nil, nil
				},
			})

			services, err := client.GetServices(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(services) != tt.expectedCount {
				t.Errorf("GetServices() count = %d, want %d", len(services), tt.expectedCount)
			}
		})
	}
}

func TestGetNetworkMode(t *testing.T) {
	tests := []struct {
		name    string
		service types.Service
		want    string
	}{
		{
			name: "awsvpc network mode",
			service: types.Service{
				NetworkConfiguration: &types.NetworkConfiguration{
					AwsvpcConfiguration: &types.AwsVpcConfiguration{},
				},
			},
			want: "awsvpc",
		},
		{
			name:    "default bridge network mode",
			service: types.Service{},
			want:    "bridge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNetworkMode(tt.service); got != tt.want {
				t.Errorf("getNetworkMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
