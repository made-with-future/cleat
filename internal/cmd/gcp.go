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

var gcpADCImpersonateLoginCmd = &cobra.Command{
	Use:   "adc-impersonate-login",
	Short: "Login to GCP with service account impersonation (ADC)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		s := strategy.NewGCPADCImpersonateLoginStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

var gcpAppEngineCmd = &cobra.Command{
	Use:   "app-engine",
	Short: "Google App Engine related commands",
}

var gcpAppEngineDeployCmd = &cobra.Command{
	Use:   "deploy [service]",
	Short: "Deploy to Google App Engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		var appYaml string
		if len(args) > 0 {
			svcName := args[0]
			for _, svc := range cfg.Services {
				if svc.Name == svcName {
					appYaml = svc.AppYaml
					break
				}
			}
			if appYaml == "" {
				return fmt.Errorf("service %s not found or has no app.yaml", svcName)
			}
		} else {
			appYaml = cfg.AppYaml
			if appYaml == "" {
				return fmt.Errorf("no root app.yaml found")
			}
		}

		s := strategy.NewGCPAppEngineDeployStrategy(appYaml)
		return s.Execute(cfg, executor.Default)
	},
}

var gcpAppEnginePromoteCmd = &cobra.Command{
	Use:   "promote [service]",
	Short: "Promote a version to receive all traffic",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		var service string
		if len(args) > 0 {
			service = args[0]
		}

		s := strategy.NewGCPAppEnginePromoteStrategy(service)
		return s.Execute(cfg, executor.Default)
	},
}

var gcpConsoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Open Google Cloud Console in browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
			return fmt.Errorf("google_cloud_platform.project_name is not configured")
		}

		s := strategy.NewGCPConsoleStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	gcpCmd.AddCommand(gcpActivateCmd)
	gcpCmd.AddCommand(gcpInitCmd)
	gcpCmd.AddCommand(gcpSetConfigCmd)
	gcpCmd.AddCommand(gcpADCLoginCmd)
	gcpCmd.AddCommand(gcpADCImpersonateLoginCmd)
	gcpAppEngineCmd.AddCommand(gcpAppEngineDeployCmd)
	gcpAppEngineCmd.AddCommand(gcpAppEnginePromoteCmd)
	gcpCmd.AddCommand(gcpAppEngineCmd)
	gcpCmd.AddCommand(gcpConsoleCmd)
	rootCmd.AddCommand(gcpCmd)
}
