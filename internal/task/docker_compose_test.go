package task

import (
	"strings"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
	"github.com/madewithfuture/cleat/internal/session"
)

// TestDockerCommandsUseCompose ensures all docker commands use "docker compose" not raw "docker"
func TestDockerCommandsUseCompose(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{Name: "backend"},
		},
	}
	exec := &executor.ShellExecutor{}
	sess := session.NewSession(cfg, exec)

	tests := []struct {
		name     string
		taskFunc func() Task
	}{
		{
			name:     "DockerBuild",
			taskFunc: func() Task { return NewDockerBuild(nil) },
		},
		{
			name:     "DockerBuild with service",
			taskFunc: func() Task { return NewDockerBuild(&cfg.Services[0]) },
		},
		{
			name:     "DockerUp",
			taskFunc: func() Task { return NewDockerUp(nil) },
		},
		{
			name:     "DockerUp with service",
			taskFunc: func() Task { return NewDockerUp(&cfg.Services[0]) },
		},
		{
			name:     "DockerDown",
			taskFunc: func() Task { return NewDockerDown(nil) },
		},
		{
			name:     "DockerDown with service",
			taskFunc: func() Task { return NewDockerDown(&cfg.Services[0]) },
		},
		{
			name:     "DockerRebuild",
			taskFunc: func() Task { return NewDockerRebuild(nil) },
		},
		{
			name:     "DockerRebuild with service",
			taskFunc: func() Task { return NewDockerRebuild(&cfg.Services[0]) },
		},
		{
			name:     "DockerRemoveOrphans",
			taskFunc: func() Task { return NewDockerRemoveOrphans(nil) },
		},
		{
			name:     "DockerRemoveOrphans with service",
			taskFunc: func() Task { return NewDockerRemoveOrphans(&cfg.Services[0]) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := tt.taskFunc()
			commands := task.Commands(sess)

			for i, cmdParts := range commands {
				if len(cmdParts) == 0 {
					t.Errorf("%s: command %d is empty", tt.name, i)
					continue
				}

				cmdName := cmdParts[0]
				args := cmdParts[1:]

				// Allow "op" as command name for 1Password integration
				if cmdName != "docker" && cmdName != "op" {
					t.Errorf("%s: expected command to be 'docker' or 'op', got %q", tt.name, cmdName)
					continue
				}

				// Find where "docker" appears in the args (for op commands)
				dockerArgIndex := -1
				if cmdName == "op" {
					for idx, arg := range args {
						if arg == "docker" {
							dockerArgIndex = idx
							break
						}
					}
					if dockerArgIndex == -1 {
						t.Errorf("%s: op command should eventually call docker", tt.name)
						continue
					}
					// Adjust args to start after "docker"
					args = args[dockerArgIndex+1:]
				}

				// Verify "compose" appears in the arguments
				foundCompose := false
				for _, arg := range args {
					if arg == "compose" {
						foundCompose = true
						break
					}
				}

				if !foundCompose {
					t.Errorf("%s: command %d does not use 'docker compose' - args: %v",
						tt.name, i, strings.Join(args, " "))
				}
			}
		})
	}
}

// TestDockerCommandsNeverUseRawDocker ensures we never use raw docker commands like "docker ps", "docker images", etc.
func TestDockerCommandsNeverUseRawDocker(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{Name: "backend"},
		},
	}
	exec := &executor.ShellExecutor{}
	sess := session.NewSession(cfg, exec)

	// List of raw docker commands we should NEVER use
	forbiddenDockerCommands := []string{
		"ps", "images", "pull", "push", "run", "exec",
		"start", "stop", "restart", "rm", "rmi",
		"container", "image", "volume", "network",
	}

	tasks := []Task{
		NewDockerBuild(nil),
		NewDockerUp(nil),
		NewDockerDown(nil),
		NewDockerRebuild(nil),
		NewDockerRemoveOrphans(nil),
	}

	for _, task := range tasks {
		commands := task.Commands(sess)
		for _, cmdParts := range commands {
			if len(cmdParts) < 2 {
				continue
			}

			// Get the args after "docker" (or after "op ... docker")
			args := cmdParts[1:]
			cmdName := cmdParts[0]
			if cmdName == "op" {
				// Find docker in args
				for i, arg := range args {
					if arg == "docker" && i+1 < len(args) {
						args = args[i+1:]
						break
					}
				}
			}

			// Check if any forbidden command appears directly after "docker"
			// (which would indicate raw docker usage)
			if len(args) > 0 {
				for _, forbidden := range forbiddenDockerCommands {
					if args[0] == forbidden {
						t.Errorf("Task %s uses forbidden raw docker command: docker %s\nCommand: %v",
							task.Name(), forbidden, strings.Join(cmdParts, " "))
					}
				}
			}
		}
	}
}

// TestDockerComposeSubcommands verifies we use proper docker compose subcommands
func TestDockerComposeSubcommands(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{Name: "backend"},
		},
	}
	exec := &executor.ShellExecutor{}
	sess := session.NewSession(cfg, exec)

	tests := []struct {
		name               string
		task               Task
		expectedSubcommand string
	}{
		{
			name:               "DockerBuild uses compose build",
			task:               NewDockerBuild(nil),
			expectedSubcommand: "build",
		},
		{
			name:               "DockerUp uses compose up",
			task:               NewDockerUp(nil),
			expectedSubcommand: "up",
		},
		{
			name:               "DockerDown uses compose down",
			task:               NewDockerDown(nil),
			expectedSubcommand: "down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := tt.task.Commands(sess)
			if len(commands) == 0 {
				t.Fatal("expected at least one command")
			}

			cmdParts := commands[0]
			args := cmdParts[1:]

			// Handle op wrapper
			if cmdParts[0] == "op" {
				for i, arg := range args {
					if arg == "docker" && i+1 < len(args) {
						args = args[i+1:]
						break
					}
				}
			}

			// Verify compose is present
			foundCompose := false
			composeIndex := -1
			for i, arg := range args {
				if arg == "compose" {
					foundCompose = true
					composeIndex = i
					break
				}
			}

			if !foundCompose {
				t.Errorf("expected 'compose' in command args: %v", strings.Join(args, " "))
				return
			}

			// Verify the expected subcommand appears after compose
			foundSubcommand := false
			for i := composeIndex + 1; i < len(args); i++ {
				if args[i] == tt.expectedSubcommand {
					foundSubcommand = true
					break
				}
			}

			if !foundSubcommand {
				t.Errorf("expected '%s' after 'compose' in command args: %v",
					tt.expectedSubcommand, strings.Join(args, " "))
			}
		})
	}
}
