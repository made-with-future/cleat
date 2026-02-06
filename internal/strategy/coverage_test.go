package strategy

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
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
		{"docker down with service", "docker down:web", "docker down"},
		{"docker rebuild with service", "docker rebuild:web", "docker rebuild"},
		{"docker remove-orphans with service", "docker remove-orphans:web", "docker remove-orphans"},
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
			var currentSess *session.Session
			if tt.name != "nil config uses registry" {
				currentSess = session.NewSession(cfg, nil)
			}
			s := GetStrategyForCommand(tt.command, currentSess)
			if s == nil {
				t.Fatalf("expected strategy for %q, got nil", tt.command)
			}
			if s.Name() != tt.want {
				t.Errorf("got name %q, want %q", s.Name(), tt.want)
			}
		})
	}
}

func ptrBool(b bool) *bool {
	return &b
}

func TestGetStrategyForCommand_EdgeCases(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{Name: "only-docker", Docker: ptrBool(true)},
		},
	}
	sess := session.NewSession(cfg, nil)

	// Unknown command
	s := GetStrategyForCommand("unknown-cmd", sess)
	if s == nil || s.Name() != "passthrough:unknown-cmd" {
		t.Errorf("expected passthrough:unknown-cmd for unknown command, got %v", s)
	}

	// Docker command with missing service
	s = GetStrategyForCommand("docker down:nonexistent", sess)
	if s == nil || s.Name() != "passthrough:docker down:nonexistent" {
		t.Errorf("expected passthrough:docker down:nonexistent for docker down with missing service, got %v", s)
	}

	// Django command with missing service
	s = GetStrategyForCommand("django migrate:nonexistent", sess)
	if s == nil || s.Name() != "passthrough:django migrate:nonexistent" {
		t.Errorf("expected passthrough:django migrate:nonexistent for django migrate with missing service, got %v", s)
	}

	// GCP deploy without app.yaml or service match
	s = GetStrategyForCommand("gcp app-engine deploy", sess)
	if s == nil || s.Name() != "passthrough:gcp app-engine deploy" {
		t.Errorf("expected passthrough:gcp app-engine deploy for gcp deploy without app.yaml, got %v", s)
	}
}