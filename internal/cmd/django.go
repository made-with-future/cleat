package cmd

import (
	"fmt"

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
	Use:   "create-user-dev [service]",
	Short: "Create a Django superuser (dev/dev)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		var s strategy.Strategy
		if len(args) > 0 {
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == args[0] {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc == nil {
				return fmt.Errorf("service '%s' not found", args[0])
			}
			s = strategy.NewDjangoCreateUserDevStrategy(targetSvc)
		} else {
			s = strategy.NewDjangoCreateUserDevStrategyGlobal(cfg)
		}
		return s.Execute(cfg, executor.Default)
	},
}

var djangoCollectStaticCmd = &cobra.Command{
	Use:   "collectstatic [service]",
	Short: "Collect Django static files",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		var s strategy.Strategy
		if len(args) > 0 {
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == args[0] {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc == nil {
				return fmt.Errorf("service '%s' not found", args[0])
			}
			s = strategy.NewDjangoCollectStaticStrategy(targetSvc)
		} else {
			s = strategy.NewDjangoCollectStaticStrategyGlobal(cfg)
		}
		return s.Execute(cfg, executor.Default)
	},
}

var djangoMigrateCmd = &cobra.Command{
	Use:   "migrate [service]",
	Short: "Run Django migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		var s strategy.Strategy
		if len(args) > 0 {
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == args[0] {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc == nil {
				return fmt.Errorf("service '%s' not found", args[0])
			}
			s = strategy.NewDjangoMigrateStrategy(targetSvc)
		} else {
			s = strategy.NewDjangoMigrateStrategyGlobal(cfg)
		}
		return s.Execute(cfg, executor.Default)
	},
}

var djangoGenRandomSecretKeyCmd = &cobra.Command{
	Use:   "gen-random-secret-key [service]",
	Short: "Generate a random Django secret key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			return err
		}

		var s strategy.Strategy
		if len(args) > 0 {
			var targetSvc *config.ServiceConfig
			for i := range cfg.Services {
				if cfg.Services[i].Name == args[0] {
					targetSvc = &cfg.Services[i]
					break
				}
			}
			if targetSvc == nil {
				return fmt.Errorf("service '%s' not found", args[0])
			}
			s = strategy.NewDjangoGenRandomSecretKeyStrategy(targetSvc)
		} else {
			s = strategy.NewDjangoGenRandomSecretKeyStrategyGlobal(cfg)
		}
		return s.Execute(cfg, executor.Default)
	},
}

func init() {
	djangoCmd.AddCommand(djangoCreateUserDevCmd)
	djangoCmd.AddCommand(djangoCollectStaticCmd)
	djangoCmd.AddCommand(djangoMigrateCmd)
	djangoCmd.AddCommand(djangoGenRandomSecretKeyCmd)
	rootCmd.AddCommand(djangoCmd)
}
