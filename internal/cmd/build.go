package cmd

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project based on cleat.yaml",
	Long:  `Executes build steps based on the project configuration in cleat.yaml. Supports Docker, Django, and NPM project types.`,
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
		s := strategy.NewBuildStrategy(cfg)
		return s.Execute(sess)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
