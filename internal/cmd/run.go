package cmd

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the project",
	Long:  `Runs the project based on cleat.yaml. If Docker is enabled, it runs 'docker compose up'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		sess := session.NewSession(cfg, executor.Default)
		if preCollectedInputs != nil {
			for k, v := range preCollectedInputs {
				sess.Inputs[k] = v
			}
		}
		s := strategy.NewRunStrategy(cfg)
		return s.Execute(sess)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
