package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var npmCmd = &cobra.Command{
	Use:   "npm [script]",
	Short: "Run an npm script",
	Long:  `Runs the specified npm script, either locally or via Docker based on configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		s := strategy.NewNpmScriptStrategy(args[0])
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	rootCmd.AddCommand(npmCmd)
}
