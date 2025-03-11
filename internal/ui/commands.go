package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbletea"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	rdssvc "github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/correctedcloud/aws-overview/internal/config"
	"github.com/correctedcloud/aws-overview/pkg/alb"
	ec2pkg "github.com/correctedcloud/aws-overview/pkg/ec2"
	ecspkg "github.com/correctedcloud/aws-overview/pkg/ecs"
	"github.com/correctedcloud/aws-overview/pkg/rds"
)

// Message types for bubbletea
type albDataLoadedMsg struct {
	loadBalancers []alb.LoadBalancerSummary
	err           error
	region        string
}

type rdsDataLoadedMsg struct {
	dbInstances []rds.DBInstanceSummary
	err         error
	region      string
}

type ec2DataLoadedMsg struct {
	instances []ec2pkg.InstanceSummary
	err       error
	region    string
}

type ecsDataLoadedMsg struct {
	services []ecspkg.ServiceSummary
	err      error
	region   string
}

// refreshTimerMsg is sent when it's time to refresh data
type refreshTimerMsg struct{}

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
			region:        cfg.Region, // Pass the potentially updated region
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
			region:      cfg.Region, // Pass the potentially updated region
		}
	}
}

// loadEC2Data is a command that loads EC2 data and returns a message
func (m Model) loadEC2Data() tea.Cmd {
	return func() tea.Msg {
		// Create context
		ctx := context.Background()

		// Load AWS config
		cfg := config.NewConfig(m.region)
		awsConfig, err := config.LoadAWSConfig(ctx, cfg)
		if err != nil {
			return ec2DataLoadedMsg{err: err}
		}

		// Create EC2 client
		ec2Client := ec2pkg.NewClient(ec2.NewFromConfig(awsConfig))

		// Get instance data
		instances, err := ec2Client.GetInstances(ctx)
		return ec2DataLoadedMsg{
			instances: instances,
			err:       err,
			region:    cfg.Region, // Pass the potentially updated region
		}
	}
}

// loadECSData is a command that loads ECS data and returns a message
func (m Model) loadECSData() tea.Cmd {
	return func() tea.Msg {
		// Create context
		ctx := context.Background()

		// Load AWS config
		cfg := config.NewConfig(m.region)
		awsConfig, err := config.LoadAWSConfig(ctx, cfg)
		if err != nil {
			return ecsDataLoadedMsg{err: err}
		}

		// Create ECS client
		ecsClient := ecspkg.NewClient(ecs.NewFromConfig(awsConfig))

		// Get service data
		services, err := ecsClient.GetServices(ctx)
		return ecsDataLoadedMsg{
			services: services,
			err:      err,
			region:   cfg.Region, // Pass the potentially updated region
		}
	}
}

// refreshTimer is a command that triggers data refresh every minute
func refreshTimer() tea.Cmd {
	return tea.Tick(time.Minute, func(time.Time) tea.Msg {
		return refreshTimerMsg{}
	})
}

// refreshData triggers a refresh of all enabled data sources
func (m Model) refreshData() tea.Cmd {
	var cmds []tea.Cmd

	if m.showALB {
		cmds = append(cmds, m.loadALBData())
	}

	if m.showRDS {
		cmds = append(cmds, m.loadRDSData())
	}

	if m.showEC2 {
		cmds = append(cmds, m.loadEC2Data())
	}

	if m.showECS {
		cmds = append(cmds, m.loadECSData())
	}

	return tea.Batch(cmds...)
}
