package alb

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// elbv2ClientAPI defines the interface for the ELBv2 client
type elbv2ClientAPI interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error)
	DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error)
}

// Client represents an ALB client
type Client struct {
	elbv2Client elbv2ClientAPI
}

// LoadBalancerSummary represents a summary of a load balancer and its target groups
type LoadBalancerSummary struct {
	Name         string
	DNSName      string
	TargetGroups []TargetGroupSummary
}

// TargetGroupSummary represents a summary of a target group and its targets
type TargetGroupSummary struct {
	Name    string
	ARN     string
	Targets []TargetSummary
}

// TargetSummary represents a summary of a target
type TargetSummary struct {
	ID     string
	Port   int32
	Status string
	Reason string
}

// NewClient returns a new ALB client
func NewClient(elbv2Client elbv2ClientAPI) *Client {
	return &Client{
		elbv2Client: elbv2Client,
	}
}

// GetLoadBalancers returns a list of load balancers with their target groups and health status
func (c *Client) GetLoadBalancers(ctx context.Context) ([]LoadBalancerSummary, error) {
	result, err := c.elbv2Client.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe load balancers: %w", err)
	}

	// Process load balancers in parallel
	var wg sync.WaitGroup
	summariesCh := make(chan LoadBalancerSummary, len(result.LoadBalancers))
	errorsCh := make(chan error, len(result.LoadBalancers))

	for _, lb := range result.LoadBalancers {
		wg.Add(1)
		go func(loadBalancer types.LoadBalancer) {
			defer wg.Done()

			// Create a summary for this load balancer
			lbSummary := LoadBalancerSummary{
				Name:    *loadBalancer.LoadBalancerName,
				DNSName: *loadBalancer.DNSName,
			}

			// Get target groups for this load balancer
			tgResult, err := c.elbv2Client.DescribeTargetGroups(ctx, &elasticloadbalancingv2.DescribeTargetGroupsInput{
				LoadBalancerArn: loadBalancer.LoadBalancerArn,
			})
			if err != nil {
				errorsCh <- fmt.Errorf("failed to describe target groups for LB %s: %w", *loadBalancer.LoadBalancerName, err)
				return
			}

			// Process target groups in parallel
			var tgWg sync.WaitGroup
			tgSummariesCh := make(chan TargetGroupSummary, len(tgResult.TargetGroups))
			tgErrorsCh := make(chan error, len(tgResult.TargetGroups))

			for _, tg := range tgResult.TargetGroups {
				tgWg.Add(1)
				go func(targetGroup types.TargetGroup) {
					defer tgWg.Done()
					tgSummary, err := c.getTargetGroupSummary(ctx, targetGroup)
					if err != nil {
						tgErrorsCh <- err
						return
					}
					tgSummariesCh <- tgSummary
				}(tg)
			}

			// Wait for all target group goroutines to complete
			tgWg.Wait()
			close(tgSummariesCh)
			close(tgErrorsCh)

			// Check for errors
			if len(tgErrorsCh) > 0 {
				errorsCh <- <-tgErrorsCh
				return
			}

			// Collect target group summaries
			for tgSummary := range tgSummariesCh {
				lbSummary.TargetGroups = append(lbSummary.TargetGroups, tgSummary)
			}

			// Send the load balancer summary
			summariesCh <- lbSummary
		}(lb)
	}

	// Wait for all load balancer goroutines to complete
	wg.Wait()
	close(summariesCh)
	close(errorsCh)

	// Check for errors
	if len(errorsCh) > 0 {
		return nil, <-errorsCh
	}

	// Collect all load balancer summaries
	var summaries []LoadBalancerSummary
	for summary := range summariesCh {
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// getTargetGroupSummary returns a summary of a target group with health status
func (c *Client) getTargetGroupSummary(ctx context.Context, tg types.TargetGroup) (TargetGroupSummary, error) {
	tgSummary := TargetGroupSummary{
		Name: *tg.TargetGroupName,
		ARN:  *tg.TargetGroupArn,
	}

	healthResult, err := c.elbv2Client.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: tg.TargetGroupArn,
	})
	if err != nil {
		return TargetGroupSummary{}, fmt.Errorf("failed to describe target health for TG %s: %w", *tg.TargetGroupName, err)
	}

	for _, target := range healthResult.TargetHealthDescriptions {
		var port int32
		if target.Target.Port != nil {
			port = *target.Target.Port
		}

		targetSummary := TargetSummary{
			ID:     *target.Target.Id,
			Port:   port,
			Status: string(target.TargetHealth.State),
		}

		if target.TargetHealth.Reason != "" {
			targetSummary.Reason = string(target.TargetHealth.Reason)
		}

		tgSummary.Targets = append(tgSummary.Targets, targetSummary)
	}

	return tgSummary, nil
}
