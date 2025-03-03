package ec2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2API defines the interface for EC2 API operations
type EC2API interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// Client is the EC2 client
type Client struct {
	ec2Client EC2API
}

// NewClient creates a new EC2 client
func NewClient(ec2Client EC2API) *Client {
	return &Client{
		ec2Client: ec2Client,
	}
}

// InstanceSummary represents an EC2 instance summary
type InstanceSummary struct {
	InstanceID       string
	InstanceType     string
	State            string
	Name             string
	PrivateIP        string
	PublicIP         string
	LaunchTime       time.Time
	Platform         string
	VpcID            string
	SubnetID         string
	SecurityGroups   []string
	Tags             map[string]string
	AvailabilityZone string
}

// GetInstances returns a list of EC2 instances
func (c *Client) GetInstances(ctx context.Context) ([]InstanceSummary, error) {
	var instances []InstanceSummary
	var nextToken *string

	for {
		resp, err := c.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}

		for _, reservation := range resp.Reservations {
			for _, instance := range reservation.Instances {
				// Skip terminated instances
				if instance.State.Name == types.InstanceStateNameTerminated {
					continue
				}

				// Extract tags into a map
				tags := make(map[string]string)
				var name string
				for _, tag := range instance.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					if aws.ToString(tag.Key) == "Name" {
						name = aws.ToString(tag.Value)
					}
				}

				// Extract security group names
				var securityGroups []string
				for _, sg := range instance.SecurityGroups {
					securityGroups = append(securityGroups, aws.ToString(sg.GroupName))
				}

				// Create instance summary
				summary := InstanceSummary{
					InstanceID:       aws.ToString(instance.InstanceId),
					InstanceType:     string(instance.InstanceType),
					State:            string(instance.State.Name),
					Name:             name,
					PrivateIP:        aws.ToString(instance.PrivateIpAddress),
					PublicIP:         aws.ToString(instance.PublicIpAddress),
					LaunchTime:       aws.ToTime(instance.LaunchTime),
					Platform:         getPlatform(instance),
					VpcID:            aws.ToString(instance.VpcId),
					SubnetID:         aws.ToString(instance.SubnetId),
					SecurityGroups:   securityGroups,
					Tags:             tags,
					AvailabilityZone: getAvailabilityZone(instance),
				}

				instances = append(instances, summary)
			}
		}

		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}

	return instances, nil
}

// getPlatform returns the platform of the instance
func getPlatform(instance types.Instance) string {
	// Platform is a string value (types.PlatformValues), not a pointer
	if instance.Platform != "" {
		return string(instance.Platform)
	}
	
	// Check platform details if platform is empty
	if instance.PlatformDetails != nil {
		details := aws.ToString(instance.PlatformDetails)
		if strings.Contains(details, "Linux") {
			return "Linux"
		}
		if strings.Contains(details, "Windows") {
			return "Windows"
		}
	}
	
	return "Unknown"
}

// getAvailabilityZone safely returns the availability zone of the instance
func getAvailabilityZone(instance types.Instance) string {
	if instance.Placement == nil || instance.Placement.AvailabilityZone == nil {
		return ""
	}
	return aws.ToString(instance.Placement.AvailabilityZone)
}
