package alb

import (
	"context"
	"testing"
	
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// Mock ELBV2 client
type mockELBV2Client struct {
	describeLoadBalancersFunc   func(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	describeTargetGroupsFunc    func(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error)
	describeTargetHealthFunc    func(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error)
}

func (m *mockELBV2Client) DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	return m.describeLoadBalancersFunc(ctx, params, optFns...)
}

func (m *mockELBV2Client) DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error) {
	return m.describeTargetGroupsFunc(ctx, params, optFns...)
}

func (m *mockELBV2Client) DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error) {
	return m.describeTargetHealthFunc(ctx, params, optFns...)
}

func TestGetLoadBalancers(t *testing.T) {
	// Create mock data
	lbName := "test-lb"
	lbARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test-lb/1234567890abcdef"
	lbDNSName := "test-lb-1234567890.us-east-1.elb.amazonaws.com"
	
	tgName := "test-tg"
	tgARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/test-tg/1234567890abcdef"
	
	targetID := "i-1234567890abcdef0"
	targetPort := int32(80)
	targetStatus := types.TargetHealthStateEnumHealthy
	
	// Create mock client
	mockClient := &mockELBV2Client{
		describeLoadBalancersFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
				LoadBalancers: []types.LoadBalancer{
					{
						LoadBalancerArn:  &lbARN,
						LoadBalancerName: &lbName,
						DNSName:          &lbDNSName,
					},
				},
			}, nil
		},
		describeTargetGroupsFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error) {
			return &elasticloadbalancingv2.DescribeTargetGroupsOutput{
				TargetGroups: []types.TargetGroup{
					{
						TargetGroupArn:  &tgARN,
						TargetGroupName: &tgName,
					},
				},
			}, nil
		},
		describeTargetHealthFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error) {
			return &elasticloadbalancingv2.DescribeTargetHealthOutput{
				TargetHealthDescriptions: []types.TargetHealthDescription{
					{
						Target: &types.TargetDescription{
							Id:   &targetID,
							Port: &targetPort,
						},
						TargetHealth: &types.TargetHealth{
							State: targetStatus,
						},
					},
				},
			}, nil
		},
	}
	
	// Create ALB client
	client := &Client{
		elbv2Client: mockClient,
	}
	
	// Call the method being tested
	lbs, err := client.GetLoadBalancers(context.Background())
	
	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(lbs) != 1 {
		t.Fatalf("Expected 1 load balancer, got %d", len(lbs))
	}
	
	lb := lbs[0]
	if lb.Name != lbName {
		t.Errorf("Expected load balancer name %s, got %s", lbName, lb.Name)
	}
	
	if lb.DNSName != lbDNSName {
		t.Errorf("Expected load balancer DNS name %s, got %s", lbDNSName, lb.DNSName)
	}
	
	if len(lb.TargetGroups) != 1 {
		t.Fatalf("Expected 1 target group, got %d", len(lb.TargetGroups))
	}
	
	tg := lb.TargetGroups[0]
	if tg.Name != tgName {
		t.Errorf("Expected target group name %s, got %s", tgName, tg.Name)
	}
	
	if len(tg.Targets) != 1 {
		t.Fatalf("Expected 1 target, got %d", len(tg.Targets))
	}
	
	target := tg.Targets[0]
	if target.ID != targetID {
		t.Errorf("Expected target ID %s, got %s", targetID, target.ID)
	}
	
	if target.Port != targetPort {
		t.Errorf("Expected target port %d, got %d", targetPort, target.Port)
	}
	
	if target.Status != string(targetStatus) {
		t.Errorf("Expected target status %s, got %s", targetStatus, target.Status)
	}
}