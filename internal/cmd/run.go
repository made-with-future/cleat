package cmd

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the project",
	Long:  `Runs the project based on cleat.yaml. If Docker is enabled, it runs 'docker compose up'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("cleat.yaml")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no cleat.yaml found in current directory")
			}
			return fmt.Errorf("error loading config: %w", err)
		}

		return runProject(cfg)
	},
}

func runProject(cfg *config.Config) error {
	if cfg.Docker {
		fmt.Println("==> Running project via Docker")
		cmdName := "docker"
		args := []string{"compose", "up", "--remove-orphans"}

		// Opinionated: support 1Password CLI if .env/dev.env exists
		if _, err := os.Stat(".env/dev.env"); err == nil {
			fmt.Println("--> Detected .env/dev.env, checking for 1Password CLI (op)")
			// Note: We use a simple check here, if 'op' is not in path it will just fail later
			// or we can check now.
			args = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, args...)
			cmdName = "op"
		}

		return runner(cmdName, args...)
	}

	// Local run logic
	if cfg.Django {
		fmt.Println("==> Running Django project locally")
		managePy := "manage.py"
		if _, err := os.Stat("backend/manage.py"); err == nil {
			managePy = "backend/manage.py"
		}
		return runner("python", managePy, "runserver")
	}

	if len(cfg.Npm.Scripts) > 0 {
		fmt.Println("==> Running frontend (NPM) locally")
		args := []string{"start"}
		if _, err := os.Stat("frontend/package.json"); err == nil {
			args = append([]string{"--prefix", "frontend"}, args...)
		}
		return runner("npm", args...)
	}

	return fmt.Errorf("no run command defined for this project type in cleat.yaml")
}

func init() {
	rootCmd.AddCommand(runCmd)
}
