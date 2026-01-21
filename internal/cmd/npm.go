package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var npmCmd = &cobra.Command{
	Use:   "npm [script] [service]",
	Short: "Run an npm script",
	Long:  `Runs the specified npm script, either locally or via Docker based on configuration. Optionally specify a service name in a mono-repo.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		var command string
		if len(args) == 2 {
			command = fmt.Sprintf("npm run %s:%s", args[1], args[0])
		} else {
			command = "npm run " + args[0]
		}

		s := strategy.GetStrategyForCommand(command, cfg)
		if s == nil {
			return fmt.Errorf("no strategy found for %s", command)
		}
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	rootCmd.AddCommand(npmCmd)
}
