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
