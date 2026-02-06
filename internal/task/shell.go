package task

import (
	"fmt"
	"strings"

	"github.com/madewithfuture/cleat/internal/session"
)

// ShellTask executes a raw shell command
type ShellTask struct {
	BaseTask
	Command string
	Args    []string
}

func NewShellTask(fullCommand string) *ShellTask {
	parts := strings.Fields(fullCommand) // Split by whitespace
	if len(parts) == 0 {
		return &ShellTask{
			BaseTask: BaseTask{
				TaskName:        "shell:empty",
				TaskDescription: "Execute an empty shell command",
			},
			Command: "",
			Args:    nil,
		}
	}

	name := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	return &ShellTask{
		BaseTask: BaseTask{
			TaskName:        fmt.Sprintf("shell:%s", name),
			TaskDescription: fmt.Sprintf("Execute shell command: %s", fullCommand),
		},
		Command: name,
		Args:    args,
	}
}

func (t *ShellTask) ShouldRun(sess *session.Session) bool {
	return t.Command != ""
}

func (t *ShellTask) Run(sess *session.Session) error {
	if err := sess.Exec.Run(t.Command, t.Args...); err != nil {
		return fmt.Errorf("shell command '%s %s' failed: %w", t.Command, strings.Join(t.Args, " "), err)
	}
	return nil
}

func (t *ShellTask) Commands(sess *session.Session) [][]string {
	return [][]string{append([]string{t.Command}, t.Args...)}
}
