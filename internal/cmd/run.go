package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the project",
	Long:  `Runs the project based on detected configuration. If Docker is enabled, it runs 'docker compose up'.`,
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
		s := strategy.NewRunStrategy(cfg)
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("run execution failed: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
