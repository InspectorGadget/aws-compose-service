package commands

import (
	"context"

	"github.com/InspectorGadget/aws-compose-service/controllers"
	"github.com/InspectorGadget/aws-compose-service/helpers"
	"github.com/InspectorGadget/aws-compose-service/structs"
	"github.com/spf13/cobra"
)

// NewUpCommand wires "aws-compose-service up".
func NewUpCommand(ctx context.Context, opt *structs.Options) *cobra.Command {
	var subnetIDs string
	var sgIDs string

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Provision / configure AWS resources for this Compose service",
		Long:  `up is invoked by Docker Compose to bring provider-managed resources online. For example, creating or wiring RDS and S3 for a Compose service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Normalise slice fields just before execution.
			opt.SubnetIDs = helpers.SplitAndTrim(subnetIDs)
			opt.SecurityGroupIDs = helpers.SplitAndTrim(sgIDs)

			return controllers.ParseUpCommand(ctx, *opt)
		},
	}

	// Service selection (also available globally)
	cmd.Flags().StringVar(&opt.Service, "service", opt.Service, "AWS service to manage (rds|s3)")

	// RDS-related options
	cmd.Flags().StringVar(&opt.Region, "region", "ap-southeast-1", "AWS region")
	cmd.Flags().StringVar(&opt.Engine, "engine", "postgres", "Database engine")
	cmd.Flags().StringVar(&opt.EngineVersion, "engine_version", "", "Database engine version")

	cmd.Flags().StringVar(&opt.InstanceClass, "instance_class", "db.t3.micro", "RDS instance class")
	cmd.Flags().IntVar(&opt.AllocatedStorage, "allocated_storage", 20, "Allocated storage (GiB)")

	cmd.Flags().StringVar(&opt.DBName, "db_name", "app", "Database name")
	cmd.Flags().StringVar(&opt.Username, "username", "admin", "Master username")
	cmd.Flags().StringVar(&opt.Password, "password", "password", "Master password")

	cmd.Flags().StringVar(&subnetIDs, "subnet_ids", "", "Comma-separated subnet IDs")
	cmd.Flags().StringVar(&sgIDs, "security_group_ids", "", "Comma-separated security group IDs")

	cmd.Flags().BoolVar(&opt.PubliclyAccessible, "publicly_accessible", false, "Make RDS instance publicly accessible")
	cmd.Flags().BoolVar(&opt.MultiAZ, "multi_az", false, "Enable Multi-AZ deployment")

	// S3-related options
	cmd.Flags().StringVar(&opt.BucketName, "bucket_name", "", "S3 bucket name (optional; will be derived if empty)")

	return cmd
}
