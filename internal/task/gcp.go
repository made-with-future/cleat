package task

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

type GCPActivate struct {
	BaseTask
}

func NewGCPActivate() *GCPActivate {
	return &GCPActivate{
		BaseTask: BaseTask{
			TaskName:        "gcp:activate",
			TaskDescription: "Activate Google Cloud Platform project",
		},
	}
}

func (t *GCPActivate) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPActivate) Run(cfg *config.Config, exec executor.Executor) error {
	if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
		return fmt.Errorf("google_cloud_platform.project_name is not configured")
	}

	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *GCPActivate) Commands(cfg *config.Config) [][]string {
	if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
		return [][]string{}
	}
	return [][]string{
		{"gcloud", "config", "configurations", "activate", cfg.GoogleCloudPlatform.ProjectName},
		{"gcloud", "auth", "application-default", "set-quota-project", cfg.GoogleCloudPlatform.ProjectName},
	}
}

type GCPInit struct {
	BaseTask
}

func NewGCPInit() *GCPInit {
	return &GCPInit{
		BaseTask: BaseTask{
			TaskName:        "gcp:init",
			TaskDescription: "Initialize Google Cloud Platform configuration",
		},
	}
}

func (t *GCPInit) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPInit) Run(cfg *config.Config, exec executor.Executor) error {
	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *GCPInit) Commands(cfg *config.Config) [][]string {
	if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
		return [][]string{}
	}
	return [][]string{
		{"gcloud", "config", "configurations", "create", cfg.GoogleCloudPlatform.ProjectName},
	}
}

type GCPSetConfig struct {
	BaseTask
}

func NewGCPSetConfig() *GCPSetConfig {
	return &GCPSetConfig{
		BaseTask: BaseTask{
			TaskName:        "gcp:set-config",
			TaskDescription: "Set Google Cloud Platform configuration properties",
		},
	}
}

func (t *GCPSetConfig) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPSetConfig) Run(cfg *config.Config, exec executor.Executor) error {
	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *GCPSetConfig) Requirements(cfg *config.Config) []InputRequirement {
	if cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != "" && cfg.GoogleCloudPlatform.Account == "" {
		return []InputRequirement{
			{
				Key:    "gcp:account",
				Prompt: "Enter your GCP account email",
			},
		}
	}
	return nil
}

func (t *GCPSetConfig) Commands(cfg *config.Config) [][]string {
	if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
		return [][]string{}
	}
	commands := [][]string{}
	account := cfg.GoogleCloudPlatform.Account
	if account == "" {
		account = cfg.Inputs["gcp:account"]
	}

	if account != "" {
		commands = append(commands, []string{"gcloud", "config", "set", "account", account})
	}
	commands = append(commands, [][]string{
		{"gcloud", "config", "set", "project", cfg.GoogleCloudPlatform.ProjectName},
		{"gcloud", "config", "set", "app/promote_by_default", "false"},
		{"gcloud", "config", "set", "billing/quota_project", cfg.GoogleCloudPlatform.ProjectName},
	}...)
	return commands
}

type GCPADCLogin struct {
	BaseTask
}

func NewGCPADCLogin() *GCPADCLogin {
	return &GCPADCLogin{
		BaseTask: BaseTask{
			TaskName:        "gcp:adc-login",
			TaskDescription: "Login to Google Cloud Platform and set application default credentials",
		},
	}
}

func (t *GCPADCLogin) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPADCLogin) Run(cfg *config.Config, exec executor.Executor) error {
	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *GCPADCLogin) Commands(cfg *config.Config) [][]string {
	if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
		return [][]string{}
	}
	return [][]string{
		{"gcloud", "config", "configurations", "activate", cfg.GoogleCloudPlatform.ProjectName},
		{"gcloud", "auth", "application-default", "login", "--project", cfg.GoogleCloudPlatform.ProjectName},
		{"gcloud", "auth", "login", "--project", cfg.GoogleCloudPlatform.ProjectName},
		{"gcloud", "auth", "application-default", "set-quota-project", cfg.GoogleCloudPlatform.ProjectName},
	}
}
