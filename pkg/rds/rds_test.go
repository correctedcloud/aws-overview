package rds

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// Mock RDS client
type mockRDSClient struct {
	describeDBInstancesFunc func(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}

func (m *mockRDSClient) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return m.describeDBInstancesFunc(ctx, params, optFns...)
}

// Mock CloudWatch client
type mockCloudWatchClient struct {
	getMetricDataFunc func(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error)
}

func (m *mockCloudWatchClient) GetMetricData(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error) {
	return m.getMetricDataFunc(ctx, params, optFns...)
}

func TestGetDBInstances(t *testing.T) {
	// Create mock data
	dbIdentifier := "test-db"
	dbEngine := "postgres"
	dbStatus := "available"
	dbClass := "db.t3.medium"
	dbEndpointAddress := "test-db.xyz123.us-east-1.rds.amazonaws.com"
	dbEndpointPort := int32(5432)

	// Create mock clients
	mockRDSClient := &mockRDSClient{
		describeDBInstancesFunc: func(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
			return &rds.DescribeDBInstancesOutput{
				DBInstances: []types.DBInstance{
					{
						DBInstanceIdentifier: &dbIdentifier,
						Engine:               &dbEngine,
						DBInstanceStatus:     &dbStatus,
						DBInstanceClass:      &dbClass,
						Endpoint: &types.Endpoint{
							Address: &dbEndpointAddress,
							Port:    &dbEndpointPort,
						},
					},
				},
			}, nil
		},
	}

	mockCloudWatchClient := &mockCloudWatchClient{
		getMetricDataFunc: func(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error) {
			// Check which metric is being requested
			id := *params.MetricDataQueries[0].Id

			var values []float64
			if id == "CPUUtilization" {
				values = []float64{10.0, 15.0, 12.0, 8.0}
			} else if id == "FreeableMemory" {
				// Return 50% free memory (2GB free out of 4GB total for a medium instance)
				values = []float64{2 * 1024 * 1024 * 1024, 2.1 * 1024 * 1024 * 1024}
			}

			return &cloudwatch.GetMetricDataOutput{
				MetricDataResults: []cwtypes.MetricDataResult{
					{
						Id:     &id,
						Values: values,
					},
				},
			}, nil
		},
	}

	// Create RDS client
	client := &Client{
		rdsClient:        mockRDSClient,
		cloudwatchClient: mockCloudWatchClient,
	}

	// Call the method being tested
	instances, err := client.GetDBInstances(context.Background())

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(instances) != 1 {
		t.Fatalf("Expected 1 instance, got %d", len(instances))
	}

	instance := instances[0]
	if instance.Identifier != dbIdentifier {
		t.Errorf("Expected instance identifier %s, got %s", dbIdentifier, instance.Identifier)
	}

	if instance.Engine != dbEngine {
		t.Errorf("Expected instance engine %s, got %s", dbEngine, instance.Engine)
	}

	if instance.Status != dbStatus {
		t.Errorf("Expected instance status %s, got %s", dbStatus, instance.Status)
	}

	expectedEndpoint := dbEndpointAddress + ":" + "5432"
	if instance.Endpoint != expectedEndpoint {
		t.Errorf("Expected instance endpoint %s, got %s", expectedEndpoint, instance.Endpoint)
	}

	if len(instance.CPUData) != 4 {
		t.Errorf("Expected 4 CPU data points, got %d", len(instance.CPUData))
	}

	if len(instance.MemoryData) != 2 {
		t.Errorf("Expected 2 memory data points, got %d", len(instance.MemoryData))
	}

	// Check if memory utilization is around 50%
	if instance.MemoryData[0] < 45 || instance.MemoryData[0] > 55 {
		t.Errorf("Expected memory utilization around 50%%, got %f%%", instance.MemoryData[0])
	}
}
