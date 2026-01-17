package strategy

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
)

type NpmStrategy struct{}

func (s *NpmStrategy) Run(cfg *config.Config, runner Runner) error {
	if len(cfg.Npm.Scripts) == 0 {
		return nil
	}

	fmt.Println("==> Building frontend (NPM)")
	for _, script := range cfg.Npm.Scripts {
		if cfg.Docker {
			fmt.Printf("--> Running npm run %s via Docker (%s service)\n", script, cfg.Npm.Service)
			err := runner("docker", "compose", "run", "--rm", cfg.Npm.Service, "npm", "run", script)
			if err != nil {
				return fmt.Errorf("failed to run npm script %s: %w", script, err)
			}
		} else {
			fmt.Printf("--> Running npm run %s locally\n", script)
			args := []string{"run", script}
			if _, err := os.Stat("frontend/package.json"); err == nil {
				args = append([]string{"--prefix", "frontend"}, args...)
			}
			err := runner("npm", args...)
			if err != nil {
				return fmt.Errorf("failed to run npm script %s: %w", script, err)
			}
		}
	}
	return nil
}
