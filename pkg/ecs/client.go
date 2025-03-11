package ecs

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// ECSAPI defines the interface for ECS API operations
type ECSAPI interface {
	ListClusters(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error)
	DescribeClusters(ctx context.Context, params *ecs.DescribeClustersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error)
	ListServices(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
	DescribeServices(ctx context.Context, params *ecs.DescribeServicesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error)
}

// Client is the ECS client
type Client struct {
	ecsClient ECSAPI
}

// NewClient creates a new ECS client
func NewClient(ecsClient ECSAPI) *Client {
	return &Client{
		ecsClient: ecsClient,
	}
}

// ServiceSummary represents an ECS service summary
type ServiceSummary struct {
	ServiceName        string
	ClusterName        string
	Status             string
	DesiredCount       int32
	RunningCount       int32
	PendingCount       int32
	TaskDefinition     string
	LaunchType         string
	CreatedAt          time.Time
	LastDeploymentTime time.Time
	Tags               map[string]string
	LoadBalancers      []string
	HealthStatus       string
	DeploymentStatus   string
	NetworkMode        string
}

// ClusterInfo represents basic cluster information
type ClusterInfo struct {
	Name                string
	Status              string
	RegisteredInstances int32
}

// GetServices returns a list of ECS services from all clusters
func (c *Client) GetServices(ctx context.Context) ([]ServiceSummary, error) {
	// Step 1: List all clusters
	clusters, err := c.getClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Step 2: Process clusters in parallel using goroutines
	var wg sync.WaitGroup
	servicesCh := make(chan []ServiceSummary, len(clusters))
	errorsCh := make(chan error, len(clusters))

	for _, cluster := range clusters {
		wg.Add(1)
		go func(clusterName string) {
			defer wg.Done()

			clusterServices, err := c.getClusterServices(ctx, clusterName)
			if err != nil {
				// Log error but don't fail the entire operation
				fmt.Printf("Error getting services for cluster %s: %v\n", clusterName, err)
				errorsCh <- fmt.Errorf("failed to get services for cluster %s: %w", clusterName, err)
				return
			}

			// Send the cluster services to the channel
			servicesCh <- clusterServices
		}(cluster.Name)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(servicesCh)
	close(errorsCh)

	// Collect all services from the channel
	var services []ServiceSummary
	for clusterServices := range servicesCh {
		services = append(services, clusterServices...)
	}

	return services, nil
}

// getClusters retrieves all ECS clusters
func (c *Client) getClusters(ctx context.Context) ([]ClusterInfo, error) {
	var clusters []ClusterInfo
	var nextToken *string

	// List all cluster ARNs
	for {
		listResp, err := c.ecsClient.ListClusters(ctx, &ecs.ListClustersInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list clusters: %w", err)
		}

		if len(listResp.ClusterArns) == 0 {
			break
		}

		// Describe clusters to get details
		descResp, err := c.ecsClient.DescribeClusters(ctx, &ecs.DescribeClustersInput{
			Clusters: listResp.ClusterArns,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe clusters: %w", err)
		}

		for _, cluster := range descResp.Clusters {
			clusters = append(clusters, ClusterInfo{
				Name:                aws.ToString(cluster.ClusterName),
				Status:              aws.ToString(cluster.Status),
				RegisteredInstances: cluster.RegisteredContainerInstancesCount,
			})
		}

		nextToken = listResp.NextToken
		if nextToken == nil {
			break
		}
	}

	return clusters, nil
}

// getClusterServices retrieves all services in a cluster
func (c *Client) getClusterServices(ctx context.Context, clusterName string) ([]ServiceSummary, error) {
	var services []ServiceSummary
	var nextToken *string

	// List all service ARNs for the cluster
	for {
		listResp, err := c.ecsClient.ListServices(ctx, &ecs.ListServicesInput{
			Cluster:   aws.String(clusterName),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list services: %w", err)
		}

		if len(listResp.ServiceArns) == 0 {
			break
		}

		// Describe services to get details
		descResp, err := c.ecsClient.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  aws.String(clusterName),
			Services: listResp.ServiceArns,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe services: %w", err)
		}

		for _, service := range descResp.Services {
			// Extract tags into a map
			tags := make(map[string]string)
			for _, tag := range service.Tags {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}

			// Extract load balancers
			var loadBalancers []string
			for _, lb := range service.LoadBalancers {
				if lb.TargetGroupArn != nil {
					// Extract the target group name from ARN
					parts := strings.Split(aws.ToString(lb.TargetGroupArn), "/")
					if len(parts) > 1 {
						loadBalancers = append(loadBalancers, parts[len(parts)-1])
					} else {
						loadBalancers = append(loadBalancers, aws.ToString(lb.TargetGroupArn))
					}
				} else if lb.LoadBalancerName != nil {
					loadBalancers = append(loadBalancers, aws.ToString(lb.LoadBalancerName))
				}
			}

			// Get deployment status and time
			deploymentStatus := "stable"
			var lastDeploymentTime time.Time

			if len(service.Deployments) > 0 {
				// Use the most recent deployment's updated time
				if service.Deployments[0].UpdatedAt != nil {
					lastDeploymentTime = aws.ToTime(service.Deployments[0].UpdatedAt)
				} else if service.Deployments[0].CreatedAt != nil {
					lastDeploymentTime = aws.ToTime(service.Deployments[0].CreatedAt)
				} else {
					// Default to service creation time if no deployment timestamps
					lastDeploymentTime = aws.ToTime(service.CreatedAt)
				}

				if len(service.Deployments) > 1 {
					deploymentStatus = "in-progress"
				} else if service.Deployments[0].RolloutState != types.DeploymentRolloutStateCompleted {
					deploymentStatus = string(service.Deployments[0].RolloutState)
				}
			} else {
				// No deployments, use service creation time
				lastDeploymentTime = aws.ToTime(service.CreatedAt)
			}

			// Get network mode from task definition ARN (just the name)
			taskDefParts := strings.Split(aws.ToString(service.TaskDefinition), "/")
			taskDefName := taskDefParts[len(taskDefParts)-1]

			// Health status (not directly available in API)
			healthStatus := "UNKNOWN"
			if service.RunningCount == service.DesiredCount && service.DesiredCount > 0 {
				healthStatus = "HEALTHY"
			} else if service.RunningCount > 0 {
				healthStatus = "PARTIAL"
			} else {
				healthStatus = "UNHEALTHY"
			}

			services = append(services, ServiceSummary{
				ServiceName:        aws.ToString(service.ServiceName),
				ClusterName:        clusterName,
				Status:             aws.ToString(service.Status),
				DesiredCount:       service.DesiredCount,
				RunningCount:       service.RunningCount,
				PendingCount:       service.PendingCount,
				TaskDefinition:     taskDefName,
				LaunchType:         string(service.LaunchType),
				CreatedAt:          aws.ToTime(service.CreatedAt),
				LastDeploymentTime: lastDeploymentTime,
				Tags:               tags,
				LoadBalancers:      loadBalancers,
				HealthStatus:       healthStatus,
				DeploymentStatus:   deploymentStatus,
				NetworkMode:        getNetworkMode(service),
			})
		}

		nextToken = listResp.NextToken
		if nextToken == nil {
			break
		}
	}

	return services, nil
}

// getNetworkMode safely returns the network mode of the service
func getNetworkMode(service types.Service) string {
	// NetworkMode is not directly accessible in the current SDK version
	// We'll just return bridge or awsvpc based on whether NetworkConfiguration is present
	if service.NetworkConfiguration != nil && service.NetworkConfiguration.AwsvpcConfiguration != nil {
		return "awsvpc"
	}

	return "bridge" // Default for most ECS services
}
