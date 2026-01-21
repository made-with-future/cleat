package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("cleat.yaml")
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("no cleat.yaml found in current directory")
				}
				return fmt.Errorf("error loading config: %w", err)
			}

			if cfg.Terraform == nil {
				return fmt.Errorf("terraform is not configured in cleat.yaml")
			}

			env := args[0]
			validEnv := false
			for _, e := range cfg.Envs {
				if e == env {
					validEnv = true
					break
				}
			}
			if !validEnv {
				return fmt.Errorf("invalid environment '%s', must be one of: %v", env, cfg.Envs)
			}

			s := strategy.NewTerraformStrategy(env, tfAction, tfArgs)
			return s.Execute(cfg, executor.Default)
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
