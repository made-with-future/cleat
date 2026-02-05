package executor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/madewithfuture/cleat/internal/logger"
)

// Executor abstracts command execution for testability
type Executor interface {
	Run(name string, args ...string) error
	RunWithDir(dir string, name string, args ...string) error
	Prompt(message string, defaultValue string) (string, error)
}

// ShellExecutor runs real shell commands
type ShellExecutor struct{}

func (e *ShellExecutor) Run(name string, args ...string) error {
	return e.RunWithDir("", name, args...)
}

func (e *ShellExecutor) RunWithDir(dir string, name string, args ...string) error {
	logger.Debug("executing command", map[string]interface{}{
		"dir":  dir,
		"name": name,
		"args": args,
	})

	if dir != "" {
		fmt.Printf("Executing (in %s): %s %s\n", dir, name, strings.Join(args, " "))
	} else {
		fmt.Printf("Executing: %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		logger.Error("command execution failed", err, map[string]interface{}{
			"dir":  dir,
			"name": name,
			"args": args,
		})
	}
	return err
}

func (e *ShellExecutor) Prompt(message string, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", message, defaultValue)
	} else {
		fmt.Printf("%s: ", message)
	}
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		logger.Error("failed to read user input", err, nil)
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

// Default is the production executor
var Default Executor = &ShellExecutor{}