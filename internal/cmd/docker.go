package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker related commands",
}

var dockerDownCmd = &cobra.Command{
	Use:   "down [service]",
	Short: "Stop Docker containers and remove orphans for all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sess := createSessionAndMerge(cfg)
		var s strategy.Strategy
		if len(args) > 0 {
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == args[0] {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc == nil {
				return fmt.Errorf("service '%s' not found", args[0])
			}
			s = strategy.NewDockerDownStrategyForService(targetSvc)
		} else {
			s = strategy.NewDockerDownStrategy(cfg)
		}
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("docker down failed: %w", err)
		}
		return nil
	},
}

var dockerRemoveOrphansCmd = &cobra.Command{
	Use:   "remove-orphans [service]",
	Short: "Remove orphan Docker containers for all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sess := createSessionAndMerge(cfg)
		var s strategy.Strategy
		if len(args) > 0 {
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == args[0] {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc == nil {
				return fmt.Errorf("service '%s' not found", args[0])
			}
			s = strategy.NewDockerRemoveOrphansStrategyForService(targetSvc)
		} else {
			s = strategy.NewDockerRemoveOrphansStrategy(cfg)
		}
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("docker remove-orphans failed: %w", err)
		}
		return nil
	},
}

var dockerRebuildCmd = &cobra.Command{
	Use:   "rebuild [service]",
	Short: "Rebuild Docker containers from scratch for all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sess := createSessionAndMerge(cfg)
		var s strategy.Strategy
		if len(args) > 0 {
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == args[0] {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc == nil {
				return fmt.Errorf("service '%s' not found", args[0])
			}
			s = strategy.NewDockerRebuildStrategyForService(targetSvc)
		} else {
			s = strategy.NewDockerRebuildStrategy(cfg)
		}
		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("docker rebuild failed: %w", err)
		}
		return nil
	},
}

func init() {
	dockerCmd.AddCommand(dockerDownCmd)
	dockerCmd.AddCommand(dockerRemoveOrphansCmd)
	dockerCmd.AddCommand(dockerRebuildCmd)
	rootCmd.AddCommand(dockerCmd)
}
