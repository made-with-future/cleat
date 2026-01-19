package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/madewithfuture/cleat/internal/ui"
	"github.com/spf13/cobra"
)

var (
	UIStart = ui.Start
	Exit    = os.Exit
)

var rootCmd = &cobra.Command{
	Use:   "cleat",
	Short: "Cleat is a TUI-based CLI tool",
	Long:  `Cleat is a tool that provides both a terminal user interface and command line actions.`,
}

func Execute() {
	run(os.Args)
}

// ... existing code ...
func run(args []string) {
	// If no subcommand is provided, run the TUI
	if len(args) == 1 {
		selected, err := UIStart()
		if err != nil {
			fmt.Printf("Error starting TUI: %v\n", err)
			Exit(1)
			return
		}

		if selected == "" {
			return
		}

		var cmdArgs []string
		if selected == "build" {
			cmdArgs = []string{"build"}
		} else if selected == "run" {
			cmdArgs = []string{"run"}
		} else if selected == "docker down" {
			cmdArgs = []string{"docker", "down"}
		} else if selected == "docker rebuild" {
			cmdArgs = []string{"docker", "rebuild"}
		} else if strings.HasPrefix(selected, "npm run ") {
			script := strings.TrimPrefix(selected, "npm run ")
			cmdArgs = []string{"npm", script}
		}

		if len(cmdArgs) > 0 {
			rootCmd.SetArgs(cmdArgs)
		} else {
			return
		}
	} else {
		rootCmd.SetArgs(args[1:])
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		Exit(1)
	}
}

// ... existing code ...

func init() {
	// Add flags or subcommands here
}
