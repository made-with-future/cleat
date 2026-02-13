package task

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

func TestRubyCommands(t *testing.T) {
	rubyCfg := &config.RubyConfig{
		Enabled:      ptrBool(true),
		Rails:        true,
		RailsService: "rails-svc",
	}
	svc := &config.ServiceConfig{
		Name:   "backend",
		Dir:    "backend",
		Docker: ptrBool(true),
		Modules: []config.ModuleConfig{
			{
				Ruby: rubyCfg,
			},
		},
	}

	tests := []struct {
		name          string
		dockerEnabled bool
		task          Task
		wantCmd       []string
	}{
		{
			name:          "migrate local",
			dockerEnabled: false,
			task:          NewRubyAction(svc, rubyCfg, "migrate"),
			wantCmd:       []string{"bundle", "exec", "rails", "db:migrate"},
		},
		{
			name:          "migrate docker",
			dockerEnabled: true,
			task:          NewRubyAction(svc, rubyCfg, "migrate"),
			wantCmd:       []string{"docker", "--log-level", "error", "compose", "run", "--rm", "rails-svc", "bundle", "exec", "rails", "db:migrate"},
		},
		{
			name:          "bundle install local",
			dockerEnabled: false,
			task:          NewRubyInstall(svc, rubyCfg),
			wantCmd:       []string{"bundle", "install"},
		},
		{
			name:          "bundle install docker",
			dockerEnabled: true,
			task:          NewRubyInstall(svc, rubyCfg),
			wantCmd:       []string{"docker", "--log-level", "error", "compose", "run", "--rm", "rails-svc", "bundle", "install"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Docker: tt.dockerEnabled}
			sess := session.NewSession(cfg, nil)
			cmds := tt.task.Commands(sess)
			if len(cmds) == 0 {
				t.Fatal("expected at least one command")
			}
			got := cmds[0]
			if len(got) != len(tt.wantCmd) {
				t.Fatalf("got %v, want %v", got, tt.wantCmd)
			}
			for i := range got {
				if got[i] != tt.wantCmd[i] {
					t.Errorf("at index %d: got %q, want %q", i, got[i], tt.wantCmd[i])
				}
			}
		})
	}
}
