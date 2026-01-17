package build

import (
	"github.com/madewithfuture/cleat/internal/config"
	"strings"
	"testing"
)

func TestNpmStrategy(t *testing.T) {
	var executedCommands []string
	runner := func(name string, args ...string) error {
		executedCommands = append(executedCommands, name+" "+strings.Join(args, " "))
		return nil
	}

	t.Run("NPM local", func(t *testing.T) {
		executedCommands = nil
		s := &NpmStrategy{}
		cfg := &config.Config{
			Npm: config.NpmConfig{
				Scripts: []string{"build"},
			},
		}
		err := s.Run(cfg, runner)
		if err != nil {
			t.Fatalf("NpmStrategy failed: %v", err)
		}
		if len(executedCommands) != 1 || executedCommands[0] != "npm run build" {
			t.Errorf("Unexpected commands: %v", executedCommands)
		}
	})

	t.Run("NPM Docker", func(t *testing.T) {
		executedCommands = nil
		s := &NpmStrategy{}
		cfg := &config.Config{
			Docker: true,
			Npm: config.NpmConfig{
				Service: "node",
				Scripts: []string{"build"},
			},
		}
		err := s.Run(cfg, runner)
		if err != nil {
			t.Fatalf("NpmStrategy failed: %v", err)
		}
		if len(executedCommands) != 1 || executedCommands[0] != "docker compose run --rm node npm run build" {
			t.Errorf("Unexpected commands: %v", executedCommands)
		}
	})
}
