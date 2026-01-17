package build

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
)

type DjangoStrategy struct{}

func (s *DjangoStrategy) Run(cfg *config.Config, runner Runner) error {
	if !cfg.Django {
		return nil
	}

	fmt.Println("==> Building Django project")
	if cfg.Docker {
		fmt.Printf("--> Running collectstatic via Docker (%s service)\n", cfg.DjangoService)
		err := runner("docker", "compose", "run", "--rm", cfg.DjangoService, "python", "manage.py", "collectstatic", "--noinput")
		if err != nil {
			return fmt.Errorf("failed to run collectstatic: %w", err)
		}
	} else {
		fmt.Println("--> Running collectstatic locally")
		managePy := "manage.py"
		if _, err := os.Stat("backend/manage.py"); err == nil {
			managePy = "backend/manage.py"
		}

		err := runner("python", managePy, "collectstatic", "--noinput")
		if err != nil {
			return fmt.Errorf("failed to run collectstatic: %w", err)
		}
	}
	return nil
}
