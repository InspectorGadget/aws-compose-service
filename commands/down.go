package commands

import (
	"context"

	"github.com/InspectorGadget/aws-compose-service/controllers"
	"github.com/InspectorGadget/aws-compose-service/helpers"
	"github.com/InspectorGadget/aws-compose-service/structs"
	"github.com/spf13/cobra"
)

// NewDownCommand wires "aws-compose-service down".
func NewDownCommand(ctx context.Context, opt *structs.Options) *cobra.Command {
	var subnetIDs string
	var sgIDs string

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Tear down AWS resources for this Compose service",
		Long:  `down is invoked by Docker Compose to bring provider-managed resources offline. For example, deleting or disconnecting RDS and S3 for a service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Normalise slice fields so we can use them in matching logic if needed.
			opt.SubnetIDs = helpers.SplitAndTrim(subnetIDs)
			opt.SecurityGroupIDs = helpers.SplitAndTrim(sgIDs)

			return controllers.ParseDownCommand(ctx, *opt)
		},
	}

	// Service selection (also available globally)
	cmd.Flags().StringVar(&opt.Service, "service", opt.Service, "AWS service to manage (rds|s3)")

	// RDS-related options (usually used to identify the instance)
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

	cmd.Flags().BoolVar(&opt.PubliclyAccessible, "publicly_accessible", false, "Instance is/was publicly accessible")
	cmd.Flags().BoolVar(&opt.MultiAZ, "multi_az", false, "Instance is/was Multi-AZ")

	// S3-related options
	cmd.Flags().StringVar(&opt.BucketName, "bucket_name", "", "S3 bucket name to tear down")

	return cmd
}
