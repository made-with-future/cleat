package cmd

import (
	"fmt"

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
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		s := strategy.NewGCPActivateStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

var gcpInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize GCP project configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		s := strategy.NewGCPInitStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

var gcpSetConfigCmd = &cobra.Command{
	Use:   "set-config",
	Short: "Set GCP project configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		s := strategy.NewGCPSetConfigStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

var gcpADCLoginCmd = &cobra.Command{
	Use:   "adc-login",
	Short: "Login to GCP and set application default credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		s := strategy.NewGCPADCLoginStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	gcpCmd.AddCommand(gcpActivateCmd)
	gcpCmd.AddCommand(gcpInitCmd)
	gcpCmd.AddCommand(gcpSetConfigCmd)
	gcpCmd.AddCommand(gcpADCLoginCmd)
	rootCmd.AddCommand(gcpCmd)
}
