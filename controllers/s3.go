package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/InspectorGadget/aws-compose-service/helpers"
	"github.com/InspectorGadget/aws-compose-service/structs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// deriveBucketName uses a sane default if none is provided.
func deriveBucketName(opt structs.Options, region string) string {
	if opt.BucketName != "" {
		return opt.BucketName
	}

	project := helpers.WithFallbackValue(opt.Project, "compose")
	name := helpers.WithFallbackValue(opt.Name, "s3")

	base := fmt.Sprintf("%s-%s-%s", project, name, region)
	return strings.ToLower(strings.ReplaceAll(base, "_", "-"))
}

// S3Up ensures a bucket exists and exports its details as environment variables.
func S3Up(ctx context.Context, opt structs.Options) error {
	region := helpers.WithFallbackValue(opt.Region, "ap-southeast-1")
	name := helpers.WithFallbackValue(opt.Name, "s3")

	bucket := deriveBucketName(opt, region)

	cfg, err := helpers.LoadAWSConfig(ctx, region)
	if err != nil {
		helpers.Error("unable to load AWS config: %v", err)
		return err
	}

	client := s3.NewFromConfig(cfg)

	// 1) Check if bucket exists
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err == nil {
		// Bucket exists → fail
		msg := fmt.Sprintf("S3 bucket %s already exists; aborting", bucket)
		helpers.Error("%s", msg)
		return fmt.Errorf("%s", msg)
	}

	// If HeadBucket returns an error, we *assume* bucket does not exist or is not accessible,
	// and proceed to create it. (If creation fails, that will be surfaced below.)
	helpers.Info("creating S3 bucket %s in region %s", bucket, region)

	createInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	// For us-east-1, LocationConstraint must be omitted
	if region != "us-east-1" {
		createInput.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}

	_, err = client.CreateBucket(ctx, createInput)
	if err != nil {
		helpers.Error("create S3 bucket failed: %v", err)
		return err
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)

	// Legacy-style env vars
	helpers.Setenv("BUCKET_NAME", bucket)
	helpers.Setenv("BUCKET_REGION", region)
	helpers.Setenv("BUCKET_URL", url)

	// S3-style env vars
	helpers.Setenv("S3_BUCKET_NAME", bucket)
	helpers.Setenv("S3_BUCKET_REGION", region)
	helpers.Setenv("S3_BUCKET_URL", url)

	helpers.Info("aws-compose-service (service=s3) ready for %s (bucket=%s)", name, bucket)

	return nil
}

// S3Down attempts to delete the bucket. It will fail if the bucket is not empty.
func S3Down(ctx context.Context, opt structs.Options) error {
	region := helpers.WithFallbackValue(opt.Region, "ap-southeast-1")
	bucket := deriveBucketName(opt, region)

	cfg, err := helpers.LoadAWSConfig(ctx, region)
	if err != nil {
		helpers.Error("unable to load AWS config: %v", err)
		return err
	}

	client := s3.NewFromConfig(cfg)

	helpers.Info("deleting S3 bucket %s in region %s (bucket must be empty)", bucket, region)

	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		switch err.(type) {
		case *s3types.NoSuchBucket:
			helpers.Info("S3 bucket %s does not exist, nothing to delete", bucket)
			return nil
		default:
			// real error, proceed
		}

		helpers.Error("delete S3 bucket failed: %v", err)
		return err
	}

	helpers.Info("S3 bucket %s delete requested", bucket)
	return nil
}
