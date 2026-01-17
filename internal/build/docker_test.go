package build

import (
	"github.com/madewithfuture/cleat/internal/config"
	"strings"
	"testing"
)

func TestDockerStrategy(t *testing.T) {
	var executedCommands []string
	runner := func(name string, args ...string) error {
		executedCommands = append(executedCommands, name+" "+strings.Join(args, " "))
		return nil
	}

	t.Run("Docker build", func(t *testing.T) {
		executedCommands = nil
		s := &DockerStrategy{}
		cfg := &config.Config{
			Docker: true,
		}
		err := s.Run(cfg, runner)
		if err != nil {
			t.Fatalf("DockerStrategy failed: %v", err)
		}
		if len(executedCommands) != 1 || executedCommands[0] != "docker compose build" {
			t.Errorf("Unexpected commands: %v", executedCommands)
		}
	})

	t.Run("No Docker", func(t *testing.T) {
		executedCommands = nil
		s := &DockerStrategy{}
		cfg := &config.Config{
			Docker: false,
		}
		err := s.Run(cfg, runner)
		if err != nil {
			t.Fatalf("DockerStrategy failed: %v", err)
		}
		if len(executedCommands) != 0 {
			t.Errorf("Expected no commands, got: %v", executedCommands)
		}
	})
}
