package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Executor abstracts command execution for testability
type Executor interface {
	Run(name string, args ...string) error
}

// ShellExecutor runs real shell commands
type ShellExecutor struct{}

func (e *ShellExecutor) Run(name string, args ...string) error {
	fmt.Printf("Executing: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Default is the production executor
var Default Executor = &ShellExecutor{}
