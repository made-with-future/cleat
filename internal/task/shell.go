package task

import (
	"fmt"
	"runtime"

	"github.com/madewithfuture/cleat/internal/session"
)

// ShellTask executes a raw shell command using the system shell
type ShellTask struct {
	BaseTask
	FullCommand string
}

func NewShellTask(fullCommand string) *ShellTask {
	return &ShellTask{
		BaseTask: BaseTask{
			TaskName:        "shell:run",
			TaskDescription: fmt.Sprintf("Execute shell command: %s", fullCommand),
		},
		FullCommand: fullCommand,
	}
}

func (t *ShellTask) ShouldRun(sess *session.Session) bool {
	return t.FullCommand != ""
}

func (t *ShellTask) Run(sess *session.Session) error {
	shell := "sh"
	shellArg := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellArg = "/c"
	}

	// We use the executor's Run method, but we pass the shell as the command
	if err := sess.Exec.Run(shell, shellArg, t.FullCommand); err != nil {
		return fmt.Errorf("shell command failed: %w", err)
	}
	return nil
}

func (t *ShellTask) Commands(sess *session.Session) [][]string {
	shell := "sh"
	shellArg := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellArg = "/c"
	}
	return [][]string{{shell, shellArg, t.FullCommand}}
}
