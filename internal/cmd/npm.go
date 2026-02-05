package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
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
			return fmt.Errorf("failed to load config: %w", err)
		}

		sess := createSessionAndMerge(cfg)
		var command string
		if len(args) == 2 {
			command = fmt.Sprintf("npm run %s:%s", args[1], args[0])
		} else {
			command = "npm run " + args[0]
		}

		s := strategy.GetStrategyForCommand(command, sess)
		if s == nil {
			return fmt.Errorf("no strategy found for %s", command)
		}
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("npm script failed: %w", err)
		}
		return nil
	},
}

var npmInstallCmd = &cobra.Command{
	Use:   "install [service]",
	Short: "Run npm install",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sess := createSessionAndMerge(cfg)
		var command string
		if len(args) == 1 {
			command = "npm install:" + args[0]
		} else {
			command = "npm install"
		}

		s := strategy.GetStrategyForCommand(command, sess)
		if s == nil {
			return fmt.Errorf("no strategy found for %s", command)
		}
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("npm install failed: %w", err)
		}
		return nil
	},
}

func init() {
	npmCmd.AddCommand(npmInstallCmd)
	rootCmd.AddCommand(npmCmd)
}
