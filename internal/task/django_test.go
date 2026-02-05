package task

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

func TestDjangoCommands(t *testing.T) {
	svc := &config.ServiceConfig{
		Name:   "backend",
		Dir:    "backend",
		Docker: ptrBool(true),
		Modules: []config.ModuleConfig{
			{
				Python: &config.PythonConfig{
					Django:         true,
					DjangoService:  "django-svc",
					PackageManager: "uv",
				},
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
			task:          NewDjangoMigrate(svc),
			wantCmd:       []string{"uv", "run", "python", "manage.py", "migrate", "--noinput"},
		},
		{
			name:          "migrate docker",
			dockerEnabled: true,
			task:          NewDjangoMigrate(svc),
			wantCmd:       []string{"docker", "--log-level", "error", "compose", "run", "--rm", "django-svc", "uv", "run", "python", "manage.py", "migrate", "--noinput"},
		},
		{
			name:          "runserver local",
			dockerEnabled: false,
			task:          NewDjangoRunServer(svc),
			wantCmd:       []string{"uv", "run", "python", "manage.py", "runserver"},
		},
		{
			name:          "runserver docker",
			dockerEnabled: true,
			task:          NewDjangoRunServer(svc),
			wantCmd:       []string{"docker", "--log-level", "error", "compose", "run", "--rm", "django-svc", "uv", "run", "python", "manage.py", "runserver", "0.0.0.0:8000"},
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

func ptrBool(b bool) *bool {
	return &b
}
