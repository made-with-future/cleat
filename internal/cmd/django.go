package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/strategy"
	"github.com/spf13/cobra"
)

var djangoCmd = &cobra.Command{
	Use:   "django",
	Short: "Django related commands",
}

var djangoCreateUserDevCmd = &cobra.Command{
	Use:   "create-user-dev",
	Short: "Create a Django superuser (dev/dev)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		s := strategy.NewDjangoCreateUserDevStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

var djangoCollectStaticCmd = &cobra.Command{
	Use:   "collectstatic",
	Short: "Collect Django static files",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		s := strategy.NewDjangoCollectStaticStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

var djangoMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run Django migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		s := strategy.NewDjangoMigrateStrategy()
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	djangoCmd.AddCommand(djangoCreateUserDevCmd)
	djangoCmd.AddCommand(djangoCollectStaticCmd)
	djangoCmd.AddCommand(djangoMigrateCmd)
	rootCmd.AddCommand(djangoCmd)
}
