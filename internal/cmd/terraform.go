package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var terraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Terraform related commands",
}

func newTerraformSubcommand(action string, short string, tfAction string, tfArgs []string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s [env]", action),
		Short: short,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg *config.Config
			var err error
			if ConfigPath != "" {
				cfg, err = config.LoadConfig(ConfigPath)
			} else {
				cfg, err = config.LoadDefaultConfig()
			}
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.Terraform == nil {
				return fmt.Errorf("terraform is not configured in cleat.yaml")
			}

			var env string
			if len(args) > 0 {
				env = args[0]
			}

			validEnvs := cfg.Terraform.Envs
			if !cfg.Terraform.UseFolders {
				// Also allow general environments when not using folders
				for _, e := range cfg.Envs {
					found := false
					for _, existing := range validEnvs {
						if existing == e {
							found = true
							break
						}
					}
					if !found {
						validEnvs = append(validEnvs, e)
					}
				}
			}

			if env == "" {
				if len(validEnvs) == 1 {
					env = validEnvs[0]
				} else if len(validEnvs) > 1 {
					return fmt.Errorf("environment is required, must be one of: %v", validEnvs)
				} else if cfg.Terraform.UseFolders {
					return fmt.Errorf("environment is required when using terraform folders")
				}
			}

			if env != "" {
				validEnv := false
				for _, e := range validEnvs {
					if e == env {
						validEnv = true
						break
					}
				}
				if !validEnv {
					return fmt.Errorf("invalid environment '%s', must be one of: %v", env, validEnvs)
				}
			}

			sess := createSessionAndMerge(cfg)
			s := strategy.NewTerraformStrategy(env, tfAction, tfArgs)
			if err := s.Execute(sess); err != nil {
				return fmt.Errorf("terraform %s failed: %w", action, err)
			}
			return nil
		},
	}
}

func init() {
	terraformCmd.AddCommand(newTerraformSubcommand("init", "Initialize Terraform", "init", nil))
	terraformCmd.AddCommand(newTerraformSubcommand("init-upgrade", "Initialize Terraform and upgrade modules/plugins", "init", []string{"-upgrade"}))
	terraformCmd.AddCommand(newTerraformSubcommand("plan", "Plan Terraform changes", "plan", nil))
	terraformCmd.AddCommand(newTerraformSubcommand("apply", "Apply Terraform changes", "apply", nil))
	terraformCmd.AddCommand(newTerraformSubcommand("apply-refresh", "Apply Terraform changes with refresh-only", "apply", []string{"-refresh-only"}))
	rootCmd.AddCommand(terraformCmd)
}
