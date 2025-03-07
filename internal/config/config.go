package config

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Config holds the AWS configuration
type Config struct {
	Region string
}

// AWSConfig is an alias for aws.Config to make imports cleaner
type AWSConfig = aws.Config

// NewConfig returns a new Config
func NewConfig(region string) *Config {
	if region == "" {
		region = os.Getenv("AWS_REGION")
		if region == "" {
			region = os.Getenv("AWS_DEFAULT_REGION")
			if region == "" {
				// Assume a region is in the profile
				return &Config{}
			}
		}
	}
	return &Config{
		Region: region,
	}
}

// LoadAWSConfig loads the AWS SDK configuration
func LoadAWSConfig(ctx context.Context, cfg *Config) (aws.Config, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return awsConfig, err
	}

	// Update config region if empty (this happens when using AWS profile)
	if cfg.Region == "" {
		cfg.Region = awsConfig.Region
	}

	return awsConfig, nil
}
