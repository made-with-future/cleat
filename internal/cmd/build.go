package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project based on cleat.yaml",
	Long:  `Executes build steps based on the project configuration in cleat.yaml. Supports Docker, Django, and NPM project types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		s := strategy.NewBuildStrategy(cfg)
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
