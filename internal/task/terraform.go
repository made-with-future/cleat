package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/session"
)

type Terraform struct {
	BaseTask
	Env    string
	Action string
	Args   []string
}

func NewTerraform(env string, action string, args []string) *Terraform {
	name := fmt.Sprintf("terraform:%s", action)
	if env != "" {
		name = fmt.Sprintf("terraform:%s:%s", action, env)
	}
	return &Terraform{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: fmt.Sprintf("Run terraform %s", action),
		},
		Env:    env,
		Action: action,
		Args:   args,
	}
}

func (t *Terraform) ShouldRun(sess *session.Session) bool {
	return sess.Config.Terraform != nil
}

func (t *Terraform) Run(sess *session.Session) error {
	baseDir := "."
	if sess.Config.SourcePath != "" {
		baseDir = filepath.Dir(sess.Config.SourcePath)
	}

	tfDir := t.getTfDir(sess)

	// Ensure folder matches env if UseFolders is true
	if sess.Config.Terraform != nil && sess.Config.Terraform.UseFolders && t.Env != "" {
		absTfDir := tfDir
		if !filepath.IsAbs(absTfDir) {
			absTfDir = filepath.Join(baseDir, tfDir)
		}
		info, err := os.Stat(absTfDir)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("terraform folder for environment '%s' not found: %s", t.Env, tfDir)
		}
	}

	PrintStep(fmt.Sprintf("Running terraform %s in %s", t.Action, tfDir))

	// 1Password integration message
	if absPath, displayPath := t.getEnvFile(sess); absPath != "" && FileUsesOp(absPath) {
		PrintSubStep(fmt.Sprintf("Detected %s, using 1Password CLI (op)", displayPath))
	}

	cmds := t.Commands(sess)
	// We run from baseDir (cleat root) and use -chdir in the command
	if err := sess.Exec.RunWithDir(baseDir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("terraform %s failed: %w", t.Action, err)
	}
	return nil
}

func (t *Terraform) Commands(sess *session.Session) [][]string {
	tfDir := t.getTfDir(sess)

	// Basic terraform command with -chdir
	cmd := []string{"terraform", "-chdir=" + tfDir, t.Action}
	cmd = append(cmd, t.Args...)

	if absPath, displayPath := t.getEnvFile(sess); absPath != "" && FileUsesOp(absPath) {
		// Use displayPath because it is already relative to cleat root
		wrappedCmd := []string{"op", "run", "--env-file=" + displayPath, "--no-masking", "--"}
		wrappedCmd = append(wrappedCmd, cmd...)
		return [][]string{wrappedCmd}
	}

	return [][]string{cmd}
}

func (t *Terraform) getEnvFile(sess *session.Session) (absPath string, displayPath string) {
	baseDir := "."
	if sess.Config.SourcePath != "" {
		baseDir = filepath.Dir(sess.Config.SourcePath)
	}

	// If t.Env is set (e.g. "prod"), try to find .envs/prod.env
	if t.Env != "" {
		path := filepath.Join(baseDir, ".envs", t.Env+".env")
		if _, err := os.Stat(path); err == nil {
			return path, filepath.Join(".envs", t.Env+".env")
		}
		// If t.Env is set, we don't want to fall back to other environment files
		// as that could lead to using wrong secrets.
		return "", ""
	}

	// Fallback to DetectEnvFile logic only if t.Env is empty
	_, absPath, displayPath = DetectEnvFile(baseDir)
	return absPath, displayPath
}

func (t *Terraform) getTfDir(sess *session.Session) string {
	tfDir := ".iac"
	if sess.Config.Terraform != nil && sess.Config.Terraform.Dir != "" {
		tfDir = sess.Config.Terraform.Dir
	}
	if sess.Config.Terraform != nil && sess.Config.Terraform.UseFolders && t.Env != "" {
		tfDir = filepath.Join(tfDir, t.Env)
	}
	return tfDir
}
