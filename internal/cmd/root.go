package cmd

import (
	"fmt"
	"os"

	"strings"
	"time"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/history"
	"github.com/madewithfuture/cleat/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	UIStart = ui.Start
	Exit    = os.Exit
	Wait    = waitForAnyKey
)

var rootCmd = &cobra.Command{
	Use:   "cleat",
	Short: "Cleat is a TUI-based CLI tool",
	Long:  `Cleat is a tool that provides both a terminal user interface and command line actions.`,
}

func Execute() {
	run(os.Args)
}

func run(args []string) {
	tuiMode := len(args) == 1
	for {
		var selected string
		if tuiMode {
			var inputs map[string]string
			var err error
			selected, inputs, err = UIStart()
			if err != nil {
				fmt.Printf("Error starting TUI: %v\n", err)
				Exit(1)
				return
			}

			if selected == "" {
				return
			}

			if len(inputs) > 0 {
				config.SetTransientInputs(inputs)
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
			} else if selected == "docker remove-orphans" {
				cmdArgs = []string{"docker", "remove-orphans"}
			} else if strings.HasPrefix(selected, "gcp ") || strings.HasPrefix(selected, "terraform ") {
				cmdArgs = strings.Fields(selected)
				if strings.Contains(selected, ":") {
					parts := strings.Split(selected, ":")
					cmdArgs = strings.Fields(parts[0])
					if len(parts) == 2 {
						cmdArgs = append(cmdArgs, parts[1])
					}
				}
			} else if strings.HasPrefix(selected, "django ") {
				parts := strings.Split(selected, ":")
				cmdPart := parts[0]
				cmdArgs = strings.Fields(cmdPart)
				if len(parts) == 2 {
					cmdArgs = append(cmdArgs, parts[1])
				}
			} else if strings.HasPrefix(selected, "npm run ") {
				scriptPart := strings.TrimPrefix(selected, "npm run ")
				parts := strings.Split(scriptPart, ":")
				if len(parts) == 2 {
					// npm run svc:script -> cleat npm script svc
					cmdArgs = []string{"npm", parts[1], parts[0]}
				} else {
					cmdArgs = []string{"npm", scriptPart}
				}
			}

			if len(cmdArgs) > 0 {
				rootCmd.SetArgs(cmdArgs)
				// Save to history
				history.Save(history.HistoryEntry{
					Timestamp: time.Now(),
					Command:   selected,
					Inputs:    inputs,
				})
			} else {
				return
			}
		} else {
			rootCmd.SetArgs(args[1:])
		}

		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if !tuiMode {
				Exit(1)
				return
			}
		}

		if tuiMode {
			Wait()
			continue
		}
		break
	}
}

func init() {
	// Add flags or subcommands here
}

func waitForAnyKey() {
	fmt.Print("\nPress any key to return to Cleat...")

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		var b [1]byte
		os.Stdin.Read(b[:])
		return
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		var b [1]byte
		os.Stdin.Read(b[:])
		return
	}
	defer term.Restore(fd, oldState)

	var b [1]byte
	os.Stdin.Read(b[:])
}
