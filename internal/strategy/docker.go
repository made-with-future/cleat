package strategy

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
)

type DockerStrategy struct{}

func (s *DockerStrategy) Run(cfg *config.Config, runner Runner) error {
	if !cfg.Docker {
		return nil
	}

	fmt.Println("==> Building Docker images")
	err := runner("docker", "compose", "build")
	if err != nil {
		return fmt.Errorf("failed to build docker images: %w", err)
	}
	return nil
}
