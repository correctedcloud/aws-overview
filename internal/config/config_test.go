package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {

	// Test with specified region
	customRegion := "eu-west-1"
	cfg := NewConfig(customRegion)
	if cfg.Region != customRegion {
		t.Errorf("Expected region '%s', got '%s'", customRegion, cfg.Region)
	}
}

func TestLoadAWSConfig(t *testing.T) {
	// This test would typically load the AWS SDK config, but since we don't want to
	// actually make AWS API calls in unit tests, we'll just check that the function exists

	// Ensure LoadAWSConfig function exists and can be called
	t.Run("LoadAWSConfig exists", func(t *testing.T) {
		// We're just checking that the function exists and can be referenced
		// We don't actually call it since we don't want to rely on AWS credentials in tests
		_ = LoadAWSConfig
	})
}
