package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Google Cloud Platform related commands",
}

var gcpActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate GCP project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		s := strategy.NewGCPActivateStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	gcpCmd.AddCommand(gcpActivateCmd)
	rootCmd.AddCommand(gcpCmd)
}
