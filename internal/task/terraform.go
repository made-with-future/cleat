package task

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

type TerraformTask struct {
	BaseTask
	Env    string
	Action string
	Args   []string
}

func NewTerraformTask(env string, action string, args []string) *TerraformTask {
	name := fmt.Sprintf("terraform:%s", action)
	if env != "" {
		name = fmt.Sprintf("%s:%s", name, env)
	}

	description := fmt.Sprintf("Terraform %s", action)
	if env != "" {
		description = fmt.Sprintf("%s for %s", description, env)
	}

	return &TerraformTask{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: description,
			TaskDeps:        []string{"gcp:activate"},
		},
		Env:    env,
		Action: action,
		Args:   args,
	}
}

func (t *TerraformTask) ShouldRun(cfg *config.Config) bool {
	return cfg.Terraform != nil
}

func (t *TerraformTask) Run(cfg *config.Config, exec executor.Executor) error {
	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *TerraformTask) Commands(cfg *config.Config) [][]string {
	if cfg.Terraform == nil {
		return [][]string{}
	}

	chdir := ".iac"
	if cfg.Terraform.UseFolders && t.Env != "" {
		chdir = filepath.Join(".iac", t.Env)
	}

	// If Env is empty, we might not want to use environment-specific paths
	// but the requirements said we have a list of envs.
	if t.Env == "" && cfg.Terraform.UseFolders {
		// Fallback or error? For now follow the pattern.
		return [][]string{}
	}

	tfArgs := []string{fmt.Sprintf("-chdir=%s", chdir), t.Action}
	tfArgs = append(tfArgs, t.Args...)

	var cmd []string
	useOp := false
	if t.Env != "" {
		envFile := filepath.Join(".envs", t.Env+".env")
		if _, err := os.Stat(envFile); err == nil {
			if _, err := exec.LookPath("op"); err == nil {
				useOp = true
				cmd = []string{
					"op", "run", "--env-file=" + envFile, "--",
					"terraform",
				}
			}
		}
	}

	if !useOp {
		cmd = []string{"terraform"}
	}
	cmd = append(cmd, tfArgs...)

	return [][]string{cmd}
}
