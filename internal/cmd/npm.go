package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
	"github.com/spf13/cobra"
)

var npmRunCmd = &cobra.Command{
	Use:    "npm-run [script]",
	Short:  "Run an NPM script defined in cleat.yaml",
	Hidden: true, // Hidden because it's mainly used by the TUI
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		return task.RunNpmScript(cfg, args[0])
	},
}

func init() {
	rootCmd.AddCommand(npmRunCmd)
}
