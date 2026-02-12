package task

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

type GoAction struct {
	BaseTask
	Service *config.ServiceConfig
	GoCfg   *config.GoConfig
	Action  string
}

func NewGoAction(svc *config.ServiceConfig, g *config.GoConfig, action string) *GoAction {
	return &GoAction{
		BaseTask: BaseTask{
			TaskName:        fmt.Sprintf("go:%s", action),
			TaskDescription: fmt.Sprintf("Run 'go %s'", action),
		},
		Service: svc,
		GoCfg:   g,
		Action:  action,
	}
}

func (t *GoAction) ShouldRun(sess *session.Session) bool {
	return t.GoCfg != nil && t.GoCfg.IsEnabled()
}

func (t *GoAction) Run(sess *session.Session) error {
	desc := fmt.Sprintf("Running 'go %s' for service %s", t.Action, t.Service.Name)
	if sess.Config.Docker && t.Service.IsDocker() && t.GoCfg.Service != "" {
		desc += fmt.Sprintf(" via Docker (%s service)", t.GoCfg.Service)
	}
	PrintStep(desc)
	cmd := t.commandArgs(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() && t.GoCfg.Service != "" {
		dir = ""
	}
	if err := sess.Exec.RunWithDir(dir, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("go %s failed for service %s: %w", t.Action, t.Service.Name, err)
	}
	return nil
}

func (t *GoAction) commandArgs(sess *session.Session) []string {
	args := t.argsForAction()
	if sess.Config.Docker && t.Service.IsDocker() && t.GoCfg.Service != "" {
		base := []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.GoCfg.Service, "go"}
		return append(base, args...)
	}
	return append([]string{"go"}, args...)
}

func (t *GoAction) Commands(sess *session.Session) [][]string {
	return [][]string{t.commandArgs(sess)}
}

func (t *GoAction) argsForAction() []string {
	switch t.Action {
	case "build":
		return []string{"build", "./..."}
	case "test":
		return []string{"test", "./..."}
	case "test-coverage":
		return []string{"test", "-cover", "-coverprofile=coverage.out", "./..."}
	case "coverage-report":
		return []string{"tool", "cover", "-func=coverage.out"}
	case "fmt":
		return []string{"fmt", "./..."}
	case "vet":
		return []string{"vet", "./..."}
	case "mod-tidy":
		return []string{"mod", "tidy"}
	case "generate":
		return []string{"generate", "./..."}
	case "run":
		return []string{"run", "."}
	default:
		return []string{t.Action}
	}
}

type GoInstall struct {
	BaseTask
	Service *config.ServiceConfig
	GoCfg   *config.GoConfig
}

func NewGoInstall(svc *config.ServiceConfig, g *config.GoConfig) *GoInstall {
	return &GoInstall{
		BaseTask: BaseTask{
			TaskName:        "go:install",
			TaskDescription: "Build and install the Go binary",
		},
		Service: svc,
		GoCfg:   g,
	}
}

func (t *GoInstall) ShouldRun(sess *session.Session) bool {
	return t.GoCfg != nil && t.GoCfg.IsEnabled()
}

func (t *GoInstall) Requirements(sess *session.Session) []InputRequirement {
	defaultPath := "/usr/local/bin"
	if runtime.GOOS != "darwin" {
		home, _ := os.UserHomeDir()
		defaultPath = filepath.Join(home, ".local/bin")
	}
	return []InputRequirement{
		{
			Key:     "install_path",
			Prompt:  "Installation path",
			Default: defaultPath,
		},
	}
}

func (t *GoInstall) binName() string {
	name := t.Service.Name
	if name == "default" || name == "" {
		absDir, err := filepath.Abs(t.Service.Dir)
		if err == nil {
			return filepath.Base(absDir)
		}
		return "app"
	}
	return name
}

func (t *GoInstall) Run(sess *session.Session) error {
	installPath := sess.Inputs["install_path"]
	if installPath == "" {
		return fmt.Errorf("install_path not provided")
	}

	binName := t.binName()
	PrintStep(fmt.Sprintf("Installing %s to %s", binName, installPath))

	// 1. Build locally
	PrintSubStep(fmt.Sprintf("Building %s...", binName))
	buildArgs := []string{"build", "-o", binName, "."}
	if err := sess.Exec.RunWithDir(t.Service.Dir, "go", buildArgs...); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	// 2. Ensure install path exists
	PrintSubStep(fmt.Sprintf("Ensuring directory %s exists...", installPath))
	if err := sess.Exec.Run("mkdir", "-p", installPath); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// 3. Copy binary
	srcPath := filepath.Join(t.Service.Dir, binName)
	dstPath := filepath.Join(installPath, binName)
	PrintSubStep(fmt.Sprintf("Copying %s to %s...", binName, dstPath))
	if err := sess.Exec.Run("cp", srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	return nil
}

func (t *GoInstall) Commands(sess *session.Session) [][]string {
	binName := t.binName()
	installPath := sess.Inputs["install_path"]
	if installPath == "" {
		installPath = "<install_path>"
	}
	return [][]string{
		{"go", "build", "-o", binName, "."},
		{"mkdir", "-p", installPath},
		{"cp", filepath.Join(t.Service.Dir, binName), filepath.Join(installPath, binName)},
	}
}
