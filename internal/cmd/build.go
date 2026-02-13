package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
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

		sess := createSessionAndMerge(cfg)
		s := strategy.GetStrategyForCommand("build", sess)
		if s == nil {
			return fmt.Errorf("no strategy found for build")
		}
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
