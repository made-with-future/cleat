package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

func getTerraformDir(cfg *config.Config) string {
	dir := ".iac"
	if cfg.Terraform != nil && cfg.Terraform.Dir != "" {
		dir = cfg.Terraform.Dir
	}
	return dir
}

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
	// Log detected env file if it uses op
	if t.Env != "" {
		envFile, displayEnv := t.detectEnvFile(cfg)
		if envFile != "" && FileUsesOp(envFile) {
			fmt.Printf("--> Detected %s, using 1Password CLI (op)\n", displayEnv)
		}
	}

	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *TerraformTask) detectEnvFile(cfg *config.Config) (path string, display string) {
	cwd, _ := os.Getwd()
	baseDir := filepath.Dir(cfg.SourcePath)
	iacDir := getTerraformDir(cfg)
	targetDir := filepath.Join(baseDir, iacDir)
	if cfg.Terraform != nil && cfg.Terraform.UseFolders && t.Env != "" {
		targetDir = filepath.Join(baseDir, iacDir, t.Env)
	}

	// 1. Check for .env in the target directory
	tfEnv := filepath.Join(targetDir, ".env")
	if _, err := os.Stat(tfEnv); err == nil {
		relPath, err := filepath.Rel(cwd, tfEnv)
		if err != nil {
			relPath = tfEnv
		}

		display = iacDir + "/.env"
		if cfg.Terraform != nil && cfg.Terraform.UseFolders && t.Env != "" {
			display = filepath.Join(iacDir, t.Env, ".env")
		}
		return relPath, display
	}

	// 2. Fallback to .envs/{{.Env}}.env
	envFile := filepath.Join(baseDir, ".envs", t.Env+".env")
	if _, err := os.Stat(envFile); err == nil {
		relPath, err := filepath.Rel(cwd, envFile)
		if err != nil {
			relPath = envFile
		}
		return relPath, filepath.Join(".envs", t.Env+".env")
	}

	return "", ""
}

func (t *TerraformTask) Commands(cfg *config.Config) [][]string {
	if cfg.Terraform == nil {
		return [][]string{}
	}

	cwd, _ := os.Getwd()
	baseDir := filepath.Dir(cfg.SourcePath)
	iacDir := getTerraformDir(cfg)
	targetDir := filepath.Join(baseDir, iacDir)
	if cfg.Terraform.UseFolders && t.Env != "" {
		targetDir = filepath.Join(baseDir, iacDir, t.Env)
	}

	// If Env is empty, we might not want to use environment-specific paths
	// but the requirements said we have a list of envs.
	if t.Env == "" && cfg.Terraform.UseFolders {
		return [][]string{}
	}

	// Make targetDir relative to CWD for shorter commands
	relTargetDir, err := filepath.Rel(cwd, targetDir)
	if err == nil {
		targetDir = relTargetDir
	}

	tfArgs := []string{fmt.Sprintf("-chdir=%s", targetDir), t.Action}
	tfArgs = append(tfArgs, t.Args...)

	var cmd []string
	useOp := false
	if t.Env != "" {
		envFile, _ := t.detectEnvFile(cfg)
		if envFile != "" && FileUsesOp(envFile) {
			useOp = true
			cmd = []string{
				"op", "run", "--env-file=" + envFile, "--no-masking", "--",
				"terraform",
			}
		}
	}

	if !useOp {
		cmd = []string{"terraform"}
	}
	cmd = append(cmd, tfArgs...)

	return [][]string{cmd}
}
