package task

import (
	"fmt"
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
	dir := ".iac"
	if sess.Config.Terraform != nil && sess.Config.Terraform.Dir != "" {
		dir = sess.Config.Terraform.Dir
	}
	if sess.Config.Terraform != nil && sess.Config.Terraform.UseFolders && t.Env != "" {
		dir = filepath.Join(dir, t.Env)
	}

	fmt.Printf("==> Running terraform %s in %s\n", t.Action, dir)
	cmds := t.Commands(sess)
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("terraform %s failed in %s: %w", t.Action, dir, err)
	}
	return nil
}

func (t *Terraform) Commands(sess *session.Session) [][]string {
	cmd := append([]string{"terraform", t.Action}, t.Args...)
	return [][]string{cmd}
}
