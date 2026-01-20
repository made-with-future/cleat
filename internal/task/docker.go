package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// DockerBuild builds Docker containers
type DockerBuild struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDockerBuild(svc *config.ServiceConfig) *DockerBuild {
	name := "docker:build"
	if svc != nil {
		name = fmt.Sprintf("docker:build:%s", svc.Name)
	}
	return &DockerBuild{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Build Docker containers",
			TaskDeps:        nil,
		},
		Service: svc,
	}
}

func (t *DockerBuild) ShouldRun(cfg *config.Config) bool {
	if t.Service != nil {
		return t.Service.Docker
	}
	return cfg.Docker
}

func (t *DockerBuild) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Building Docker containers")
	cmds := t.Commands(cfg)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Location
	}
	return exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DockerBuild) Commands(cfg *config.Config) [][]string {
	cmd := []string{"docker", "compose"}
	cmd = append(cmd, "build")
	return [][]string{cmd}
}

// DockerUp starts Docker containers
type DockerUp struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDockerUp(svc *config.ServiceConfig) *DockerUp {
	name := "docker:up"
	if svc != nil {
		name = fmt.Sprintf("docker:up:%s", svc.Name)
	}
	return &DockerUp{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Start Docker containers",
			TaskDeps:        nil,
		},
		Service: svc,
	}
}

func (t *DockerUp) ShouldRun(cfg *config.Config) bool {
	if t.Service != nil {
		return t.Service.Docker
	}
	return cfg.Docker
}

func (t *DockerUp) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Running project via Docker")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Location != "" {
		searchDir = t.Service.Location
	}
	if _, err := os.Stat(filepath.Join(searchDir, ".env/dev.env")); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
	}

	cmds := t.Commands(cfg)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Location
	}
	return exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DockerUp) Commands(cfg *config.Config) [][]string {
	cmdName := "docker"
	args := []string{"compose"}
	args = append(args, "up", "--remove-orphans")

	// 1Password integration
	searchDir := "."
	envFile := "./.env/dev.env"
	if t.Service != nil && t.Service.Location != "" {
		searchDir = t.Service.Location
		envFile = ".env/dev.env"
	}
	if _, err := os.Stat(filepath.Join(searchDir, ".env/dev.env")); err == nil {
		args = append([]string{"run", "--env-file", envFile, "--", "docker"}, args...)
		cmdName = "op"
	}

	return [][]string{append([]string{cmdName}, args...)}
}

// DockerDown stops Docker containers
type DockerDown struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDockerDown(svc *config.ServiceConfig) *DockerDown {
	name := "docker:down"
	if svc != nil {
		name = fmt.Sprintf("docker:down:%s", svc.Name)
	}
	return &DockerDown{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Stop Docker containers",
			TaskDeps:        nil,
		},
		Service: svc,
	}
}

func (t *DockerDown) ShouldRun(cfg *config.Config) bool {
	if t.Service != nil {
		return t.Service.Docker
	}
	return cfg.Docker
}

func (t *DockerDown) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Stopping Docker containers (all profiles)")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Location != "" {
		searchDir = t.Service.Location
	}
	if _, err := os.Stat(filepath.Join(searchDir, ".env/dev.env")); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
	}

	cmds := t.Commands(cfg)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Location
	}
	return exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DockerDown) Commands(cfg *config.Config) [][]string {
	cmdName := "docker"
	args := []string{"compose"}
	args = append(args, "--profile", "*", "down", "--remove-orphans")

	// 1Password integration
	searchDir := "."
	envFile := "./.env/dev.env"
	if t.Service != nil && t.Service.Location != "" {
		searchDir = t.Service.Location
		envFile = ".env/dev.env"
	}
	if _, err := os.Stat(filepath.Join(searchDir, ".env/dev.env")); err == nil {
		args = append([]string{"run", "--env-file", envFile, "--", "docker"}, args...)
		cmdName = "op"
	}

	return [][]string{append([]string{cmdName}, args...)}
}

// DockerRebuild stops all containers, removes images/volumes, and rebuilds without cache
type DockerRebuild struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDockerRebuild(svc *config.ServiceConfig) *DockerRebuild {
	name := "docker:rebuild"
	if svc != nil {
		name = fmt.Sprintf("docker:rebuild:%s", svc.Name)
	}
	return &DockerRebuild{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Stop, remove all and rebuild Docker containers (all profiles, no cache)",
			TaskDeps:        nil,
		},
		Service: svc,
	}
}

func (t *DockerRebuild) ShouldRun(cfg *config.Config) bool {
	if t.Service != nil {
		return t.Service.Docker
	}
	return cfg.Docker
}

func (t *DockerRebuild) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Rebuilding Docker containers (all profiles, no cache)")

	// 1Password integration check for logging
	searchDir := "."
	if t.Service != nil && t.Service.Location != "" {
		searchDir = t.Service.Location
	}
	if _, err := os.Stat(filepath.Join(searchDir, ".env/dev.env")); err == nil {
		fmt.Println("--> Detected .env/dev.env, using 1Password CLI (op)")
	}

	cmds := t.Commands(cfg)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Location
	}

	// 1. Down with --rmi all --volumes
	fmt.Println("--> Cleaning up: stopping containers and removing images/volumes")
	if err := exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return err
	}

	// 2. Build with --no-cache
	fmt.Println("--> Rebuilding: build without cache")
	return exec.RunWithDir(dir, cmds[1][0], cmds[1][1:]...)
}

func (t *DockerRebuild) Commands(cfg *config.Config) [][]string {
	// 1. Down
	downCmd := "docker"
	downArgs := []string{"compose"}
	downArgs = append(downArgs, "--profile", "*", "down", "--remove-orphans", "--rmi", "all", "--volumes")

	// 2. Build
	buildCmd := "docker"
	buildArgs := []string{"compose"}
	buildArgs = append(buildArgs, "--profile", "*", "build", "--no-cache")

	// 1Password integration
	searchDir := "."
	envFile := "./.env/dev.env"
	if t.Service != nil && t.Service.Location != "" {
		searchDir = t.Service.Location
		envFile = ".env/dev.env"
	}

	if _, err := os.Stat(filepath.Join(searchDir, ".env/dev.env")); err == nil {
		downArgs = append([]string{"run", "--env-file", envFile, "--", "docker"}, downArgs[1:]...)
		downCmd = "op"

		buildArgs = append([]string{"run", "--env-file", envFile, "--", "docker"}, buildArgs[1:]...)
		buildCmd = "op"
	}

	return [][]string{
		append([]string{downCmd}, downArgs...),
		append([]string{buildCmd}, buildArgs...),
	}
}
