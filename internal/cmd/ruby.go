package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var rubyCmd = &cobra.Command{
	Use:   "ruby",
	Short: "Ruby and Rails project commands",
}

func newRubySubcommand(action string, short string, strategyAction string) *cobra.Command {
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

			// Ensure at least one Ruby module is detected
			foundRuby := false
			for i := range cfg.Services {
				for j := range cfg.Services[i].Modules {
					if cfg.Services[i].Modules[j].Ruby != nil {
						foundRuby = true
						break
					}
				}
				if foundRuby {
					break
				}
			}
			if !foundRuby {
				return fmt.Errorf("ruby project not detected or configured")
			}

			cmdStr := "ruby " + strategyAction
			if len(args) == 1 {
				cmdStr += ":" + args[0]
			}
			sess := createSessionAndMerge(cfg)
			s := strategy.GetStrategyForCommand(cmdStr, sess)
			if s == nil {
				return fmt.Errorf("no strategy found for %s", cmdStr)
			}
			if err := s.Execute(sess); err != nil {
				return fmt.Errorf("ruby %s failed: %w", action, err)
			}
			return nil
		},
	}
}

func init() {
	rubyCmd.AddCommand(newRubySubcommand("migrate", "Run database migrations", "migrate"))
	rubyCmd.AddCommand(newRubySubcommand("console", "Open Rails console", "console"))
	rubyCmd.AddCommand(newRubySubcommand("server", "Start Rails server", "server"))
	rubyCmd.AddCommand(newRubySubcommand("install", "Run bundle install", "install"))
	rootCmd.AddCommand(rubyCmd)
}
