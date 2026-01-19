package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/task"
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
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		t := task.NewDockerDown()
		if !t.ShouldRun(cfg) {
			fmt.Println("Docker is not enabled in cleat.yaml")
			return nil
		}

		return t.Run(cfg, executor.Default)
	},
}

var dockerRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild Docker containers from scratch for all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		t := task.NewDockerRebuild()
		if !t.ShouldRun(cfg) {
			fmt.Println("Docker is not enabled in cleat.yaml")
			return nil
		}

		return t.Run(cfg, executor.Default)
	},
}

func init() {
	dockerCmd.AddCommand(dockerDownCmd)
	dockerCmd.AddCommand(dockerRebuildCmd)
	rootCmd.AddCommand(dockerCmd)
}
