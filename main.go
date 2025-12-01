package main

import (
	"context"
	"os"

	"github.com/InspectorGadget/aws-compose-service/commands"
	"github.com/InspectorGadget/aws-compose-service/helpers"
	"github.com/InspectorGadget/aws-compose-service/structs"
	"github.com/spf13/cobra"
)

func newRootCommand(ctx context.Context) *cobra.Command {
	opt := &structs.Options{}

	root := &cobra.Command{
		Use:   "aws-compose-service",
		Short: "Docker Compose AWS provider for RDS and S3",
		Long: `aws-compose-service is a Docker Compose provider that wires AWS
				services (RDS and S3 for now) into your Compose applications by emitting
				JSONL messages and environment variables that Docker Compose understands.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags that Compose typically passes in
	root.PersistentFlags().StringVar(&opt.Project, "project-name", "", "Compose project name (alias)")
	root.PersistentFlags().StringVar(&opt.Name, "name", "", "Compose service logical name")

	// The entrypoint Docker Compose uses:
	composeCmd := &cobra.Command{
		Use:   "compose",
		Short: "Entry point when invoked by docker compose",
		Long:  "This command is used by Docker Compose when calling the provider.",
	}

	composeCmd.AddCommand(
		commands.NewUpCommand(ctx, opt),
		commands.NewDownCommand(ctx, opt),
	)

	// Also allow direct usage:
	//   aws-compose-service up ...
	//   aws-compose-service down ...
	root.AddCommand(
		composeCmd,
		commands.NewUpCommand(ctx, opt),
		commands.NewDownCommand(ctx, opt),
	)

	return root
}

func main() {
	ctx := context.Background()
	root := newRootCommand(ctx)

	if err := root.Execute(); err != nil {
		helpers.Error("command failed: %v", err)
		os.Exit(1)
	}
}
