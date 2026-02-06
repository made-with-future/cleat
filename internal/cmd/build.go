package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
	Long:  `Executes build steps based on the project configuration. Supports Docker, Django, and NPM project types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sess := session.NewSession(cfg, executor.Default)
		if preCollectedInputs != nil {
			for k, v := range preCollectedInputs {
				sess.Inputs[k] = v
			}
		}
		s := strategy.NewBuildStrategy(cfg)
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("build execution failed: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
