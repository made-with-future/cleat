package task

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/strategy"
)

var CommandRunner strategy.Runner = RunCommand

func Build(cfg *config.Config) error {
	strategies := []strategy.Strategy{
		&strategy.NpmStrategy{},
		&strategy.DjangoStrategy{},
		&strategy.DockerStrategy{},
	}

	for _, s := range strategies {
		if err := s.Run(cfg, strategy.Runner(CommandRunner)); err != nil {
			return err
		}
	}

	if !cfg.Django && !cfg.Docker && len(cfg.Npm.Scripts) == 0 {
		fmt.Println("No build steps defined for this project type in cleat.yaml")
	} else {
		fmt.Println("==> Build completed successfully")
	}

	return nil
}

func Run(cfg *config.Config) error {
	if cfg.Docker {
		fmt.Println("==> Running project via Docker")
		cmdName := "docker"
		args := []string{"compose", "up", "--remove-orphans"}

		if _, err := os.Stat(".env/dev.env"); err == nil {
			fmt.Println("--> Detected .env/dev.env, checking for 1Password CLI (op)")
			args = append([]string{"run", "--env-file", "./.env/dev.env", "--", "docker"}, args...)
			cmdName = "op"
		}

		return CommandRunner(cmdName, args...)
	}

	if cfg.Django {
		fmt.Println("==> Running Django project locally")
		managePy := "manage.py"
		if _, err := os.Stat("backend/manage.py"); err == nil {
			managePy = "backend/manage.py"
		}
		return CommandRunner("python", managePy, "runserver")
	}

	if len(cfg.Npm.Scripts) > 0 {
		fmt.Println("==> Running frontend (NPM) locally")
		args := []string{"start"}
		if _, err := os.Stat("frontend/package.json"); err == nil {
			args = append([]string{"--prefix", "frontend"}, args...)
		}
		return CommandRunner("npm", args...)
	}

	return fmt.Errorf("no run command defined for this project type in cleat.yaml")
}

func RunNpmScript(cfg *config.Config, script string) error {
	if cfg.Docker {
		fmt.Printf("==> Running npm run %s via Docker (%s service)\n", script, cfg.Npm.Service)
		return CommandRunner("docker", "compose", "run", "--rm", cfg.Npm.Service, "npm", "run", script)
	}

	fmt.Printf("==> Running npm run %s locally\n", script)
	args := []string{"run", script}
	if _, err := os.Stat("frontend/package.json"); err == nil {
		args = append([]string{"--prefix", "frontend"}, args...)
	}
	return CommandRunner("npm", args...)
}

func RunCommand(name string, args ...string) error {
	fmt.Printf("Executing: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
