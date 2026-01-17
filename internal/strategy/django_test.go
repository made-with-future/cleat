package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"strings"
	"testing"
)

func TestDjangoStrategy(t *testing.T) {
	var executedCommands []string
	runner := func(name string, args ...string) error {
		executedCommands = append(executedCommands, name+" "+strings.Join(args, " "))
		return nil
	}

	t.Run("Django local", func(t *testing.T) {
		executedCommands = nil
		s := &DjangoStrategy{}
		cfg := &config.Config{
			Django: true,
		}
		err := s.Run(cfg, runner)
		if err != nil {
			t.Fatalf("DjangoStrategy failed: %v", err)
		}
		if len(executedCommands) != 1 || executedCommands[0] != "python manage.py collectstatic --noinput" {
			t.Errorf("Unexpected commands: %v", executedCommands)
		}
	})

	t.Run("Django Docker", func(t *testing.T) {
		executedCommands = nil
		s := &DjangoStrategy{}
		cfg := &config.Config{
			Docker:        true,
			Django:        true,
			DjangoService: "web",
		}
		err := s.Run(cfg, runner)
		if err != nil {
			t.Fatalf("DjangoStrategy failed: %v", err)
		}
		if len(executedCommands) != 1 || executedCommands[0] != "docker compose run --rm web python manage.py collectstatic --noinput" {
			t.Errorf("Unexpected commands: %v", executedCommands)
		}
	})
}
