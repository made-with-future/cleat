package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/madewithfuture/cleat/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cleat",
	Short: "Cleat is a TUI-based CLI tool",
	Long:  `Cleat is a tool that provides both a terminal user interface and command line actions.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This will be handled in Execute()
	},
}

func Execute() {
	// If no subcommand is provided, run the TUI
	if len(os.Args) == 1 {
		selected, err := ui.Start()
		if err != nil {
			fmt.Printf("Error starting TUI: %v\n", err)
			os.Exit(1)
		}

		if selected == "" {
			return
		}

		var cmdArgs []string
		if selected == "build" {
			cmdArgs = []string{"build"}
		} else if selected == "run" {
			cmdArgs = []string{"run"}
		} else if strings.HasPrefix(selected, "npm run ") {
			script := strings.TrimPrefix(selected, "npm run ")
			cmdArgs = []string{"npm-run", script}
		}

		if len(cmdArgs) > 0 {
			os.Args = append([]string{os.Args[0]}, cmdArgs...)
		} else {
			return
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add flags or subcommands here
}
