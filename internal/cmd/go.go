package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var golangCmd = &cobra.Command{
	Use:   "go",
	Short: "Go project commands",
}

func newGoSubcommand(action string, short string, strategyAction string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s [service]", action),
		Short: short,
		Args:  cobra.MaximumNArgs(1),
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

			// Ensure at least one Go module is detected
			foundGo := false
			for i := range cfg.Services {
				for j := range cfg.Services[i].Modules {
					if cfg.Services[i].Modules[j].Go != nil {
						foundGo = true
						break
					}
				}
				if foundGo {
					break
				}
			}
			if !foundGo {
				return fmt.Errorf("go project not detected or configured")
			}

			cmdStr := "go " + strategyAction
			if len(args) == 1 {
				cmdStr += ":" + args[0]
			}
			sess := createSessionAndMerge(cfg)
			s := strategy.GetStrategyForCommand(cmdStr, sess)
			if s == nil {
				return fmt.Errorf("no strategy found for %s", cmdStr)
			}
			if err := s.Execute(sess); err != nil {
				return fmt.Errorf("go %s failed: %w", action, err)
			}
			return nil
		},
	}
}

func init() {
	golangCmd.AddCommand(newGoSubcommand("build", "Build all packages", "build"))
	golangCmd.AddCommand(newGoSubcommand("test", "Run tests", "test"))
	golangCmd.AddCommand(newGoSubcommand("fmt", "Format code", "fmt"))
	golangCmd.AddCommand(newGoSubcommand("vet", "Vet code", "vet"))

	modCmd := &cobra.Command{
		Use:   "mod",
		Short: "Module maintenance",
	}
	modCmd.AddCommand(newGoSubcommand("tidy", "Tidy go.mod", "mod tidy"))
	golangCmd.AddCommand(modCmd)

	golangCmd.AddCommand(newGoSubcommand("generate", "Run go generate", "generate"))
	golangCmd.AddCommand(newGoSubcommand("run", "Run the main package (go run .)", "run"))
	golangCmd.AddCommand(newGoSubcommand("coverage", "Run tests with coverage and display summary", "coverage"))
	golangCmd.AddCommand(newGoSubcommand("install", "Build and install the binary", "install"))
	rootCmd.AddCommand(golangCmd)
}
