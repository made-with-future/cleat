package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfigTerraformAutoDetection(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string
		expectedTf     *TerraformConfig
		expectedEnvs   []string
		expectedTfEnvs []string
	}{
		{
			name: "no terraform",
			files: map[string]string{
				"cleat.yaml": "version: 1",
			},
			expectedTf:     nil,
			expectedEnvs:   nil,
			expectedTfEnvs: nil,
		},
		{
			name: "terraform single env (files in .iac)",
			files: map[string]string{
				"cleat.yaml":   "version: 1",
				".iac/main.tf": "",
			},
			expectedTf:     &TerraformConfig{UseFolders: false},
			expectedEnvs:   nil,
			expectedTfEnvs: nil,
		},
		{
			name: "terraform multi env (folders in .iac)",
			files: map[string]string{
				"cleat.yaml":           "version: 1",
				".iac/prod/main.tf":    "",
				".iac/staging/main.tf": "",
				".envs/prod.env":       "",
				".envs/staging.env":    "",
			},
			expectedTf:     &TerraformConfig{UseFolders: true, Envs: []string{"prod", "staging"}},
			expectedEnvs:   []string{"prod", "staging"},
			expectedTfEnvs: []string{"prod", "staging"},
		},
		{
			name: "terraform multi env with matching .envs",
			files: map[string]string{
				"cleat.yaml":        "version: 1",
				".iac/prod/main.tf": "",
				".envs/prod.env":    "",
			},
			expectedTf:     &TerraformConfig{UseFolders: true, Envs: []string{"prod"}},
			expectedEnvs:   []string{"prod"},
			expectedTfEnvs: []string{"prod"},
		},
		{
			name: "terraform folders with extra .envs (repro issue)",
			files: map[string]string{
				"cleat.yaml":        "version: 1",
				".iac/prod/main.tf": "",
				".envs/prod.env":    "",
				".envs/local.env":   "",
			},
			expectedTf:     &TerraformConfig{UseFolders: true, Envs: []string{"prod"}},
			expectedEnvs:   []string{"local", "prod"},
			expectedTfEnvs: []string{"prod"},
		},
		{
			name: "terraform multi env without matching .envs",
			files: map[string]string{
				"cleat.yaml":        "version: 1",
				".iac/prod/main.tf": "",
			},
			expectedTf:     &TerraformConfig{UseFolders: true, Envs: []string{"prod"}},
			expectedEnvs:   []string{"prod"},
			expectedTfEnvs: []string{"prod"},
		},
		{
			name: "terraform mixed mating",
			files: map[string]string{
				"cleat.yaml":           "version: 1",
				".iac/prod/main.tf":    "",
				".iac/staging/main.tf": "",
				".envs/prod.env":       "",
				// staging.env is missing
			},
			expectedTf:     &TerraformConfig{UseFolders: true, Envs: []string{"prod", "staging"}},
			expectedEnvs:   []string{"prod", "staging"},
			expectedTfEnvs: []string{"prod", "staging"},
		},
		{
			name: "terraform single env, no .envs",
			files: map[string]string{
				"cleat.yaml":   "version: 1",
				".iac/main.tf": "",
			},
			expectedTf:     &TerraformConfig{UseFolders: false},
			expectedEnvs:   nil,
			expectedTfEnvs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "cleat-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			cfg, err := LoadConfig(filepath.Join(tmpDir, "cleat.yaml"))
			if err != nil {
				t.Fatalf("LoadConfig failed: %v", err)
			}

			if !reflect.DeepEqual(cfg.Terraform, tt.expectedTf) {
				t.Errorf("expected TerraformConfig %+v, got %+v", tt.expectedTf, cfg.Terraform)
			}

			if !reflect.DeepEqual(cfg.Envs, tt.expectedEnvs) {
				t.Errorf("expected Envs %v, got %v", tt.expectedEnvs, cfg.Envs)
			}

			if cfg.Terraform != nil {
				if !reflect.DeepEqual(cfg.Terraform.Envs, tt.expectedTfEnvs) {
					t.Errorf("expected Terraform.Envs %v, got %v", tt.expectedTfEnvs, cfg.Terraform.Envs)
				}
			}
		})
	}
}
