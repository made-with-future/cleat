package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/history"
	"github.com/madewithfuture/cleat/internal/logger"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/ui"
	"github.com/madewithfuture/cleat/internal/ui/theme"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type WaitAction int

const (
	WaitReturn WaitAction = iota // any other key: return to cleat
	WaitRerun                    // 'r' or 'R'
	WaitExit                     // 'q' or 'Q'
)

type WaitFunc func() WaitAction

var (
	ConfigPath string
	UIStart    = func(version string) (string, map[string]string, error) {
		configPath := ConfigPath
		if configPath == "" {
			root := config.FindProjectRoot()
			configPath = filepath.Join(root, "cleat.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				if _, err := os.Stat(filepath.Join(root, "cleat.yml")); err == nil {
					configPath = filepath.Join(root, "cleat.yml")
				}
			}
		}
		return ui.Start(version, configPath)
	}
	Exit          = os.Exit
	Wait WaitFunc = waitForAnyKey
)

var preCollectedInputs map[string]string

var rootCmd = &cobra.Command{
	Use:   "cleat",
	Short: "Cleat is a TUI-based CLI tool",
	Long:  `Cleat is a tool that provides both a terminal user interface and command line actions.`,
}

func createSessionAndMerge(cfg *config.Config) *session.Session {
	sess := session.NewSession(cfg, executor.Default)
	if preCollectedInputs != nil {
		for k, v := range preCollectedInputs {
			sess.Inputs[k] = v
		}
	}
	return sess
}

func Execute() {
	run(os.Args)
}

func run(args []string) {
	tuiMode := len(args) == 1
	var commandQueue []struct {
		selected      string
		inputs        map[string]string
		workflowRunID string
	}

	for {
		var selected string
		var inputs map[string]string
		var workflowRunID string
		if len(commandQueue) > 0 {
			item := commandQueue[0]
			commandQueue = commandQueue[1:]
			selected = item.selected
			inputs = item.inputs
			workflowRunID = item.workflowRunID
			logger.Info("running command from workflow queue", map[string]interface{}{"command": selected, "run_id": workflowRunID})
		} else if tuiMode {
			var err error
			selected, inputs, err = UIStart(Version)
			if err != nil {
				logger.Error("failed to start TUI", err, nil)
				fmt.Printf("Error starting TUI: %v\n", err)
				Exit(1)
				return
			}

			if selected == "" {
				logger.Debug("no command selected in TUI, exiting", nil)
				return
			}
		}

		if selected != "" {
			preCollectedInputs = inputs
			cmdArgs := mapSelectedToArgs(selected)

			if len(cmdArgs) > 0 {
				rootCmd.SetArgs(cmdArgs)
			} else {
				if tuiMode {
					logger.Warn("could not map selected command to CLI args", map[string]interface{}{"command": selected})
					continue
				}
				return
			}
		} else {
			rootCmd.SetArgs(args[1:])
		}

		for {
			err := rootCmd.Execute()
			if err != nil {
				logger.Error("command execution failed", err, map[string]interface{}{"selected": selected})
				fmt.Fprintln(os.Stderr, err)
				if !tuiMode {
					Exit(1)
					return
				}
				if len(commandQueue) > 0 {
					fmt.Println("Workflow failed. Stopping.")
					commandQueue = nil
				}
			}

			if tuiMode && selected != "" {
				history.Save(history.HistoryEntry{
					Timestamp:     time.Now(),
					Command:       selected,
					Inputs:        inputs,
					Success:       err == nil,
					WorkflowRunID: workflowRunID,
				})
				if workflowRunID == "" {
					history.UpdateStats(selected)
				}
			}

			if len(commandQueue) > 0 {
				fmt.Println("\nRunning next command in workflow...")
				time.Sleep(1 * time.Second)
				break
			}

			if tuiMode {
				switch Wait() {
				case WaitRerun:
					fmt.Println()
					if selected != "" {
						commandQueue = append(commandQueue, struct {
							selected      string
							inputs        map[string]string
							workflowRunID string
						}{selected, inputs, workflowRunID})
					}
					break
				case WaitExit:
					return
				default:
					// any other key: return to Cleat (go back to TUI)
				}
			}
			break
		}

		if !tuiMode {
			break
		}
	}
}

func mapSelectedToArgs(selected string) []string {
	var cmdArgs []string
	if strings.HasPrefix(selected, "workflow:") {
		// Let the dispatcher handle it
		cmdArgs = []string{"workflow", strings.TrimPrefix(selected, "workflow:")}
	} else if strings.HasPrefix(selected, "docker ") || strings.HasPrefix(selected, "gcp ") || strings.HasPrefix(selected, "terraform ") {
		cmdArgs = strings.Fields(selected)
		if strings.Contains(selected, ":") {
			parts := strings.Split(selected, ":")
			cmdArgs = strings.Fields(parts[0])
			if len(parts) == 2 {
				cmdArgs = append(cmdArgs, parts[1])
			}
		}
	} else if strings.HasPrefix(selected, "django ") {
		if colonIdx := strings.LastIndex(selected, ":"); colonIdx != -1 {
			cmdPart := selected[:colonIdx]
			svcName := selected[colonIdx+1:]
			cmdArgs = strings.Fields(cmdPart)
			cmdArgs = append(cmdArgs, svcName)
		} else {
			cmdArgs = strings.Fields(selected)
		}
	} else if strings.HasPrefix(selected, "npm run ") {
		scriptPart := strings.TrimPrefix(selected, "npm run ")
		parts := strings.SplitN(scriptPart, ":", 2)
		if len(parts) == 2 {
			cmdArgs = []string{"npm", parts[1], parts[0]}
		} else {
			cmdArgs = []string{"npm", scriptPart}
		}
	} else if strings.HasPrefix(selected, "npm install:") {
		svcName := strings.TrimPrefix(selected, "npm install:")
		cmdArgs = []string{"npm", "install", svcName}
	} else if strings.HasPrefix(selected, "go ") {
		if colonIdx := strings.LastIndex(selected, ":"); colonIdx != -1 {
			cmdPart := selected[:colonIdx]
			svcName := selected[colonIdx+1:]
			cmdArgs = strings.Fields(cmdPart)
			cmdArgs = append(cmdArgs, svcName)
		} else {
			cmdArgs = strings.Fields(selected)
		}
	}
	return cmdArgs
}

func init() {
}

func waitForAnyKey() WaitAction {
	fmt.Printf("\nPress %s to exit, %s to rerun, or %s to return to cleat",
		lipgloss.NewStyle().Foreground(theme.Red).Render("q"),
		lipgloss.NewStyle().Foreground(theme.Purple).Render("r"),
		lipgloss.NewStyle().Foreground(theme.Green).Render("any other key"),
	)

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Fallback for non-terminal (e.g. tests)
		reader := bufio.NewReader(os.Stdin)
		r, _, _ := reader.ReadRune()
		switch r {
		case 'r', 'R':
			return WaitRerun
		case 'q', 'Q':
			return WaitExit
		default:
			return WaitReturn
		}
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		var b [1]byte
		os.Stdin.Read(b[:])
		return WaitReturn
	}
	defer term.Restore(fd, oldState)

	var b [1]byte
	n, _ := os.Stdin.Read(b[:])
	if n > 0 {
		switch b[0] {
		case 'r', 'R':
			return WaitRerun
		case 'q', 'Q':
			return WaitExit
		default:
			return WaitReturn
		}
	}
	return WaitReturn
}
