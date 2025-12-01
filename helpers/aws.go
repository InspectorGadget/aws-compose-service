package helpers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAWSConfig loads an AWS config for the given region.
func LoadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
	if region == "" {
		region = "ap-southeast-1"
	}

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("load AWS config: %w", err)
	}

	return cfg, nil
}
