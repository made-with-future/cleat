package cmd

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker related commands",
}

var dockerDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop Docker containers and remove orphans for all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		s := strategy.NewDockerDownStrategy(cfg)
		return s.Execute(cfg, executor.Default)
	},
}

var dockerRemoveOrphansCmd = &cobra.Command{
	Use:   "remove-orphans",
	Short: "Remove orphan Docker containers for all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		s := strategy.NewDockerRemoveOrphansStrategy(cfg)
		return s.Execute(cfg, executor.Default)
	},
}

var dockerRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild Docker containers from scratch for all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		s := strategy.NewDockerRebuildStrategy(cfg)
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	dockerCmd.AddCommand(dockerDownCmd)
	dockerCmd.AddCommand(dockerRemoveOrphansCmd)
	dockerCmd.AddCommand(dockerRebuildCmd)
	rootCmd.AddCommand(dockerCmd)
}
