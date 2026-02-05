package task

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
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

func (t *DockerBuild) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		return t.Service.IsDocker()
	}
	return sess.Config.Docker
}

func (t *DockerBuild) Run(sess *session.Session) error {
	fmt.Println("==> Building Docker containers")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, displayEnv := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		fmt.Printf("--> Detected %s, using 1Password CLI (op)\n", displayEnv)
	}

	cmds := t.Commands(sess)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Dir
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}
	return nil
}

func (t *DockerBuild) Commands(sess *session.Session) [][]string {
	cmdName := "docker"
	args := []string{"--log-level", "error", "compose"}
	if t.Service == nil {
		args = append(args, "--profile", "*")
	}
	args = append(args, "build")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, _ := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		args = append([]string{"run", "--env-file", execPath, "--", "docker"}, args...)
		cmdName = "op"
	}

	return [][]string{append([]string{cmdName}, args...)}
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

func (t *DockerUp) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		return t.Service.IsDocker()
	}
	return sess.Config.Docker
}

func (t *DockerUp) Run(sess *session.Session) error {
	fmt.Println("==> Running project via Docker")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, displayEnv := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		fmt.Printf("--> Detected %s, using 1Password CLI (op)\n", displayEnv)
	}

	cmds := t.Commands(sess)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Dir
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("docker up failed: %w", err)
	}
	return nil
}

func (t *DockerUp) Commands(sess *session.Session) [][]string {
	cmdName := "docker"
	args := []string{"--log-level", "error", "compose"}
	args = append(args, "up", "--remove-orphans")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, _ := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		args = append([]string{"run", "--env-file", execPath, "--", "docker"}, args...)
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

func (t *DockerDown) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		return t.Service.IsDocker()
	}
	return sess.Config.Docker
}

func (t *DockerDown) Run(sess *session.Session) error {
	fmt.Println("==> Stopping Docker containers (all profiles)")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, displayEnv := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		fmt.Printf("--> Detected %s, using 1Password CLI (op)\n", displayEnv)
	}

	cmds := t.Commands(sess)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Dir
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("docker down failed: %w", err)
	}
	return nil
}

func (t *DockerDown) Commands(sess *session.Session) [][]string {
	cmdName := "docker"
	args := []string{"compose"}
	args = append(args, "--profile", "*", "down", "--remove-orphans")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, _ := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		args = append([]string{"run", "--env-file", execPath, "--", "docker"}, args...)
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

func (t *DockerRebuild) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		return t.Service.IsDocker()
	}
	return sess.Config.Docker
}

func (t *DockerRebuild) Run(sess *session.Session) error {
	fmt.Println("==> Rebuilding Docker containers (all profiles, no cache)")

	// 1Password integration check for logging
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, displayEnv := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		fmt.Printf("--> Detected %s, using 1Password CLI (op)\n", displayEnv)
	}

	cmds := t.Commands(sess)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Dir
	}

	// 1. Down with --rmi all --volumes
	fmt.Println("--> Cleaning up: stopping containers and removing images/volumes")
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("docker cleanup failed during rebuild: %w", err)
	}

	// 2. Build with --no-cache
	fmt.Println("--> Rebuilding: build without cache")
	if err := sess.Exec.RunWithDir(dir, cmds[1][0], cmds[1][1:]...); err != nil {
		return fmt.Errorf("docker rebuild failed: %w", err)
	}
	return nil
}

func (t *DockerRebuild) Commands(sess *session.Session) [][]string {
	// 1. Down
	downCmd := "docker"
	downArgs := []string{"compose"}
	downArgs = append(downArgs, "--profile", "*", "down", "--remove-orphans", "--rmi", "all", "--volumes")

	// 2. Build
	buildCmd := "docker"
	buildArgs := []string{"--log-level", "error", "compose"}
	buildArgs = append(buildArgs, "--profile", "*", "build", "--no-cache")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}

	if execPath, absPath, _ := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		downArgs = append([]string{"run", "--env-file", execPath, "--", "docker"}, downArgs...)
		downCmd = "op"

		buildArgs = append([]string{"run", "--env-file", execPath, "--", "docker"}, buildArgs...)
		buildCmd = "op"
	}

	return [][]string{
		append([]string{downCmd}, downArgs...),
		append([]string{buildCmd}, buildArgs...),
	}
}

// DockerRemoveOrphans removes orphan Docker containers
type DockerRemoveOrphans struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDockerRemoveOrphans(svc *config.ServiceConfig) *DockerRemoveOrphans {
	name := "docker:remove-orphans"
	if svc != nil {
		name = fmt.Sprintf("docker:remove-orphans:%s", svc.Name)
	}
	return &DockerRemoveOrphans{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Remove orphan Docker containers (all profiles)",
			TaskDeps:        nil,
		},
		Service: svc,
	}
}

func (t *DockerRemoveOrphans) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		return t.Service.IsDocker()
	}
	return sess.Config.Docker
}

func (t *DockerRemoveOrphans) Run(sess *session.Session) error {
	fmt.Println("==> Removing orphan Docker containers (all profiles)")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, displayEnv := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		fmt.Printf("--> Detected %s, using 1Password CLI (op)\n", displayEnv)
	}

	cmds := t.Commands(sess)
	dir := ""
	if t.Service != nil {
		dir = t.Service.Dir
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("docker remove-orphans failed: %w", err)
	}
	return nil
}

func (t *DockerRemoveOrphans) Commands(sess *session.Session) [][]string {
	cmdName := "docker"
	args := []string{"compose"}
	args = append(args, "--profile", "*", "down", "--remove-orphans")

	// 1Password integration
	searchDir := "."
	if t.Service != nil && t.Service.Dir != "" {
		searchDir = t.Service.Dir
	}
	if execPath, absPath, _ := DetectEnvFile(searchDir); execPath != "" && FileUsesOp(absPath) {
		args = append([]string{"run", "--env-file", execPath, "--", "docker"}, args...)
		cmdName = "op"
	}

	return [][]string{append([]string{cmdName}, args...)}
}
