package cmd

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var runWorkflowCmd = &cobra.Command{
	Use:   "run-workflow [name]",
	Short: "Run a named workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wfName := args[0]
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sess := createSessionAndMerge(cfg)
		
		// Use the dispatcher to get the workflow strategy
		s := strategy.GetStrategyForCommand("workflow:"+wfName, sess)
		if s == nil {
			return fmt.Errorf("unknown workflow: %s", wfName)
		}

		if err := s.Execute(sess); err != nil {
			return fmt.Errorf("workflow execution failed: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runWorkflowCmd)
}
