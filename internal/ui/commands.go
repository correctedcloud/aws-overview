package ui

import (
	"context"

	"github.com/charmbracelet/bubbletea"
	
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	rdssvc "github.com/aws/aws-sdk-go-v2/service/rds"
	
	"github.com/correctedcloud/aws-overview/internal/config"
	"github.com/correctedcloud/aws-overview/pkg/alb"
	"github.com/correctedcloud/aws-overview/pkg/rds"
)

// Message types for bubbletea
type albDataLoadedMsg struct {
	loadBalancers []alb.LoadBalancerSummary
	err           error
}

type rdsDataLoadedMsg struct {
	dbInstances []rds.DBInstanceSummary
	err         error
}

// loadALBData is a command that loads ALB data and returns a message
func (m Model) loadALBData() tea.Cmd {
	return func() tea.Msg {
		// Create context
		ctx := context.Background()
		
		// Load AWS config
		cfg := config.NewConfig(m.region)
		awsConfig, err := config.LoadAWSConfig(ctx, cfg)
		if err != nil {
			return albDataLoadedMsg{err: err}
		}
		
		// Create ALB client
		albClient := alb.NewClient(elasticloadbalancingv2.NewFromConfig(awsConfig))
		
		// Get load balancer data
		lbs, err := albClient.GetLoadBalancers(ctx)
		return albDataLoadedMsg{
			loadBalancers: lbs,
			err:           err,
		}
	}
}

// loadRDSData is a command that loads RDS data and returns a message
func (m Model) loadRDSData() tea.Cmd {
	return func() tea.Msg {
		// Create context
		ctx := context.Background()
		
		// Load AWS config
		cfg := config.NewConfig(m.region)
		awsConfig, err := config.LoadAWSConfig(ctx, cfg)
		if err != nil {
			return rdsDataLoadedMsg{err: err}
		}
		
		// Create RDS client
		rdsClient := rds.NewClient(
			rdssvc.NewFromConfig(awsConfig),
			cloudwatch.NewFromConfig(awsConfig),
		)
		
		// Get DB instance data
		instances, err := rdsClient.GetDBInstances(ctx)
		return rdsDataLoadedMsg{
			dbInstances: instances,
			err:         err,
		}
	}
}