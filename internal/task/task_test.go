package task

import (
	"os"
	"reflect"
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestBuild(t *testing.T) {
	oldRunner := CommandRunner
	defer func() { CommandRunner = oldRunner }()

	var captured [][]string
	CommandRunner = func(name string, args ...string) error {
		captured = append(captured, append([]string{name}, args...))
		return nil
	}

	tests := []struct {
		name     string
		cfg      *config.Config
		expected [][]string
	}{
		{
			name: "NPM only",
			cfg: &config.Config{
				Npm: config.NpmConfig{
					Scripts: []string{"build"},
				},
			},
			expected: [][]string{
				{"npm", "run", "build"},
			},
		},
		{
			name: "Django only (local)",
			cfg: &config.Config{
				Django: true,
			},
			expected: [][]string{
				{"python", "manage.py", "collectstatic", "--noinput"},
			},
		},
		{
			name: "Docker only",
			cfg: &config.Config{
				Docker: true,
			},
			expected: [][]string{
				{"docker", "compose", "build"},
			},
		},
		{
			name: "Full project (Docker)",
			cfg: &config.Config{
				Docker:        true,
				Django:        true,
				DjangoService: "web",
				Npm: config.NpmConfig{
					Scripts: []string{"build"},
					Service: "web",
				},
			},
			expected: [][]string{
				{"docker", "compose", "run", "--rm", "web", "npm", "run", "build"},
				{"docker", "compose", "run", "--rm", "web", "python", "manage.py", "collectstatic", "--noinput"},
				{"docker", "compose", "build"},
			},
		},
		{
			name:     "Empty config",
			cfg:      &config.Config{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured = nil
			err := Build(tt.cfg)
			if err != nil {
				t.Fatalf("Build failed: %v", err)
			}
			if !reflect.DeepEqual(captured, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, captured)
			}
		})
	}
}

func TestRun(t *testing.T) {
	oldRunner := CommandRunner
	defer func() { CommandRunner = oldRunner }()

	var captured [][]string
	CommandRunner = func(name string, args ...string) error {
		captured = append(captured, append([]string{name}, args...))
		return nil
	}

	// Create a temp directory for filesystem checks
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	tests := []struct {
		name     string
		cfg      *config.Config
		setup    func()
		expected [][]string
		wantErr  bool
	}{
		{
			name: "Docker",
			cfg: &config.Config{
				Docker: true,
			},
			expected: [][]string{
				{"docker", "compose", "up", "--remove-orphans"},
			},
		},
		{
			name: "Docker with op",
			cfg: &config.Config{
				Docker: true,
			},
			setup: func() {
				os.MkdirAll(".env", 0755)
				os.WriteFile(".env/dev.env", []byte("FOO=BAR"), 0644)
			},
			expected: [][]string{
				{"op", "run", "--env-file", "./.env/dev.env", "--", "docker", "compose", "up", "--remove-orphans"},
			},
		},
		{
			name: "Django local",
			cfg: &config.Config{
				Django: true,
			},
			expected: [][]string{
				{"python", "manage.py", "runserver"},
			},
		},
		{
			name: "Django local in backend/",
			cfg: &config.Config{
				Django: true,
			},
			setup: func() {
				os.MkdirAll("backend", 0755)
				os.WriteFile("backend/manage.py", []byte(""), 0644)
			},
			expected: [][]string{
				{"python", "backend/manage.py", "runserver"},
			},
		},
		{
			name: "NPM local",
			cfg: &config.Config{
				Npm: config.NpmConfig{
					Scripts: []string{"dev"},
				},
			},
			expected: [][]string{
				{"npm", "start"},
			},
		},
		{
			name: "NPM local in frontend/",
			cfg: &config.Config{
				Npm: config.NpmConfig{
					Scripts: []string{"dev"},
				},
			},
			setup: func() {
				os.MkdirAll("frontend", 0755)
				os.WriteFile("frontend/package.json", []byte("{}"), 0644)
			},
			expected: [][]string{
				{"npm", "--prefix", "frontend", "start"},
			},
		},
		{
			name:    "No run command",
			cfg:     &config.Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp dir for each test
			entries, _ := os.ReadDir(".")
			for _, e := range entries {
				os.RemoveAll(e.Name())
			}

			if tt.setup != nil {
				tt.setup()
			}
			captured = nil
			err := Run(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(captured, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, captured)
			}
		})
	}
}

func TestRunNpmScript(t *testing.T) {
	oldRunner := CommandRunner
	defer func() { CommandRunner = oldRunner }()

	var captured [][]string
	CommandRunner = func(name string, args ...string) error {
		captured = append(captured, append([]string{name}, args...))
		return nil
	}

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	tests := []struct {
		name     string
		cfg      *config.Config
		script   string
		setup    func()
		expected [][]string
	}{
		{
			name: "Docker",
			cfg: &config.Config{
				Docker: true,
				Npm: config.NpmConfig{
					Service: "node",
				},
			},
			script: "build",
			expected: [][]string{
				{"docker", "compose", "run", "--rm", "node", "npm", "run", "build"},
			},
		},
		{
			name:   "Local",
			cfg:    &config.Config{},
			script: "test",
			expected: [][]string{
				{"npm", "run", "test"},
			},
		},
		{
			name:   "Local in frontend/",
			cfg:    &config.Config{},
			script: "test",
			setup: func() {
				os.MkdirAll("frontend", 0755)
				os.WriteFile("frontend/package.json", []byte("{}"), 0644)
			},
			expected: [][]string{
				{"npm", "--prefix", "frontend", "run", "test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp dir
			entries, _ := os.ReadDir(".")
			for _, e := range entries {
				os.RemoveAll(e.Name())
			}

			if tt.setup != nil {
				tt.setup()
			}
			captured = nil
			err := RunNpmScript(tt.cfg, tt.script)
			if err != nil {
				t.Fatalf("RunNpmScript failed: %v", err)
			}
			if !reflect.DeepEqual(captured, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, captured)
			}
		})
	}
}

func TestBuildError(t *testing.T) {
	oldRunner := CommandRunner
	defer func() { CommandRunner = oldRunner }()

	CommandRunner = func(name string, args ...string) error {
		return os.ErrPermission
	}

	cfg := &config.Config{Docker: true}
	err := Build(cfg)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
