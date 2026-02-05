package task

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

func TestNpmCommands(t *testing.T) {
	svc := &config.ServiceConfig{
		Name:   "frontend",
		Dir:    "frontend",
		Docker: ptrBool(true),
	}
	npm := &config.NpmConfig{
		Service: "frontend-svc",
		Scripts: []string{"build", "start"},
	}

	tests := []struct {
		name          string
		dockerEnabled bool
		task          Task
		wantCmd       []string
	}{
		{
			name:          "npm run build local",
			dockerEnabled: false,
			task:          NewNpmRun(svc, npm, "build"),
			wantCmd:       []string{"npm", "run", "build"},
		},
		{
			name:          "npm run build docker",
			dockerEnabled: true,
			task:          NewNpmRun(svc, npm, "build"),
			wantCmd:       []string{"docker", "--log-level", "error", "compose", "run", "--rm", "frontend-svc", "npm", "run", "build"},
		},
		{
			name:          "npm install local",
			dockerEnabled: false,
			task:          NewNpmInstall(svc, npm),
			wantCmd:       []string{"npm", "install"},
		},
		{
			name:          "npm install docker",
			dockerEnabled: true,
			task:          NewNpmInstall(svc, npm),
			wantCmd:       []string{"docker", "--log-level", "error", "compose", "run", "--rm", "frontend-svc", "npm", "install"},
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
