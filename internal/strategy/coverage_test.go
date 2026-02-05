package strategy

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestGetStrategyForCommand_AdditionalCoverage(t *testing.T) {
	cfg := &config.Config{
		Docker: true,
		Services: []config.ServiceConfig{
			{
				Name: "web",
				AppYaml: "web/app.yaml",
				Modules: []config.ModuleConfig{
					{Npm: &config.NpmConfig{Scripts: []string{"start", "test"}}},
				},
			},
			{
				Name: "api",
				Modules: []config.ModuleConfig{
					{Python: &config.PythonConfig{Django: true}},
				},
			},
		},
	}

	tests := []struct {
		name    string
		command string
		want    string
	}{
		{"npm run with service prefix", "npm run web:test", "npm:test"},
		{"npm run matching script in any service", "npm run test", "npm:test"},
		{"npm run fallback to first npm service", "npm run missing-script", "npm:missing-script"},
		{"docker down with service", "docker down:web", "docker down:web"},
		{"docker rebuild with service", "docker rebuild:web", "docker rebuild:web"},
		{"docker remove-orphans with service", "docker remove-orphans:web", "docker remove-orphans:web"},
		{"django runserver with service", "django runserver:api", "django runserver"},
		{"django migrate with service", "django migrate:api", "django migrate"},
		{"django makemigrations with service", "django makemigrations:api", "django makemigrations"},
		{"django collectstatic with service", "django collectstatic:api", "django collectstatic"},
		{"django create-user-dev with service", "django create-user-dev:api", "django create-user-dev"},
		{"django gen-random-secret-key with service", "django gen-random-secret-key:api", "django gen-random-secret-key"},
		{"terraform init-upgrade", "terraform init-upgrade:prod", "terraform:init:prod"},
		{"terraform apply-refresh", "terraform apply-refresh:prod", "terraform:apply:prod"},
		{"gcp app-engine deploy with service", "gcp app-engine deploy:web", "gcp:app-engine-deploy"},
		{"gcp app-engine promote root", "gcp app-engine promote", "gcp:app-engine-promote"},
		{"gcp app-engine promote with service", "gcp app-engine promote:web", "gcp:app-engine-promote"},
		{"nil config uses registry", "build", "build"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var currentCfg *config.Config
			if tt.name != "nil config uses registry" {
				currentCfg = cfg
			}
			s := GetStrategyForCommand(tt.command, currentCfg)
			if s == nil {
				t.Fatalf("expected strategy for %q, got nil", tt.command)
			}
			if s.Name() != tt.want {
				t.Errorf("got name %q, want %q", s.Name(), tt.want)
			}
		})
	}
}

func TestGetStrategyForCommand_EdgeCases(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{Name: "only-docker", Docker: ptrBool(true)},
		},
	}

	// Unknown command
	s := GetStrategyForCommand("unknown-cmd", cfg)
	if s != nil {
		t.Errorf("expected nil for unknown command, got %v", s)
	}

	// Docker command with missing service
	s = GetStrategyForCommand("docker down:nonexistent", cfg)
	if s != nil {
		t.Errorf("expected nil for docker down with missing service, got %v", s)
	}

	// Django command with missing service
	s = GetStrategyForCommand("django migrate:nonexistent", cfg)
	if s != nil {
		t.Errorf("expected nil for django migrate with missing service, got %v", s)
	}

	// GCP deploy without app.yaml or service match
	s = GetStrategyForCommand("gcp app-engine deploy", cfg)
	if s != nil {
		t.Errorf("expected nil for gcp deploy without app.yaml, got %v", s)
	}
}
