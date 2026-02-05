package task

import (
	"fmt"
	"runtime"

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

type GCPADCImpersonateLogin struct {
	BaseTask
}

func NewGCPADCImpersonateLogin() *GCPADCImpersonateLogin {
	return &GCPADCImpersonateLogin{
		BaseTask: BaseTask{
			TaskName:        "gcp:adc-impersonate-login",
			TaskDescription: "Login to Google Cloud Platform with service account impersonation (ADC)",
		},
	}
}

func (t *GCPADCImpersonateLogin) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPADCImpersonateLogin) Run(cfg *config.Config, exec executor.Executor) error {
	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *GCPADCImpersonateLogin) Requirements(cfg *config.Config) []InputRequirement {
	if cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ImpersonateServiceAccount != "" {
		return nil
	}
	return []InputRequirement{
		{
			Key:    "gcp:impersonate-service-account",
			Prompt: "Enter service account email to impersonate",
		},
	}
}

func (t *GCPADCImpersonateLogin) Commands(cfg *config.Config) [][]string {
	if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
		return [][]string{}
	}
	email := cfg.Inputs["gcp:impersonate-service-account"]
	if email == "" && cfg.GoogleCloudPlatform.ImpersonateServiceAccount != "" {
		email = cfg.GoogleCloudPlatform.ImpersonateServiceAccount
	}
	if email == "" {
		return [][]string{}
	}

	return [][]string{
		{"gcloud", "config", "configurations", "activate", cfg.GoogleCloudPlatform.ProjectName},
		{"gcloud", "auth", "application-default", "login", "--impersonate-service-account", email, "--project", cfg.GoogleCloudPlatform.ProjectName},
		{"gcloud", "auth", "login", "--impersonate-service-account", email, "--project", cfg.GoogleCloudPlatform.ProjectName},
	}
}

type GCPAppEngineDeploy struct {
	BaseTask
	AppYaml string
}

func NewGCPAppEngineDeploy(appYaml string) *GCPAppEngineDeploy {
	return &GCPAppEngineDeploy{
		BaseTask: BaseTask{
			TaskName:        "gcp:app-engine-deploy",
			TaskDescription: fmt.Sprintf("Deploy to Google App Engine using %s", appYaml),
		},
		AppYaml: appYaml,
	}
}

func (t *GCPAppEngineDeploy) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPAppEngineDeploy) Run(cfg *config.Config, exec executor.Executor) error {
	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *GCPAppEngineDeploy) Requirements(cfg *config.Config) []InputRequirement {
	return []InputRequirement{
		{
			Key:    "gcp:version",
			Prompt: "Enter version name, or return to skip",
		},
	}
}

func (t *GCPAppEngineDeploy) Commands(cfg *config.Config) [][]string {
	version := cfg.Inputs["gcp:version"]
	cmd := []string{"gcloud", "app", "deploy", t.AppYaml}
	if version != "" {
		cmd = append(cmd, "--version", version)
	}
	return [][]string{cmd}
}

type GCPAppEnginePromote struct {
	BaseTask
	Service string
}

func NewGCPAppEnginePromote(service string) *GCPAppEnginePromote {
	name := "gcp:app-engine-promote"
	desc := "Promote a version to receive all traffic"
	if service != "" && service != "default" {
		name = fmt.Sprintf("gcp:app-engine-promote:%s", service)
		desc = fmt.Sprintf("Promote a version to receive all traffic for service %s", service)
	}
	return &GCPAppEnginePromote{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: desc,
		},
		Service: service,
	}
}

func (t *GCPAppEnginePromote) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPAppEnginePromote) Run(cfg *config.Config, exec executor.Executor) error {
	commands := t.Commands(cfg)
	for _, cmd := range commands {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (t *GCPAppEnginePromote) Requirements(cfg *config.Config) []InputRequirement {
	return []InputRequirement{
		{
			Key:    "gcp:promote_version",
			Prompt: "Enter version to promote",
		},
	}
}

func (t *GCPAppEnginePromote) Commands(cfg *config.Config) [][]string {
	version := cfg.Inputs["gcp:promote_version"]
	if version == "" {
		return nil
	}
	cmd := []string{"gcloud", "app", "versions", "migrate", version}
	return [][]string{cmd}
}

type GCPConsole struct {
	BaseTask
}

func NewGCPConsole() *GCPConsole {
	return &GCPConsole{
		BaseTask: BaseTask{
			TaskName:        "gcp:console",
			TaskDescription: "Open Google Cloud Console in browser",
		},
	}
}

func (t *GCPConsole) ShouldRun(cfg *config.Config) bool {
	return cfg.GoogleCloudPlatform != nil && cfg.GoogleCloudPlatform.ProjectName != ""
}

func (t *GCPConsole) Run(cfg *config.Config, exec executor.Executor) error {
	url := fmt.Sprintf("https://console.cloud.google.com/home/dashboard?project=%s", cfg.GoogleCloudPlatform.ProjectName)
	fmt.Printf("Opening %s\n", url)

	cmds := t.Commands(cfg)
	if len(cmds) == 0 {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *GCPConsole) Commands(cfg *config.Config) [][]string {
	if cfg.GoogleCloudPlatform == nil || cfg.GoogleCloudPlatform.ProjectName == "" {
		return nil
	}
	url := fmt.Sprintf("https://console.cloud.google.com/home/dashboard?project=%s", cfg.GoogleCloudPlatform.ProjectName)

	switch runtime.GOOS {
	case "linux":
		return [][]string{{"xdg-open", url}}
	case "darwin":
		return [][]string{{"open", url}}
	case "windows":
		return [][]string{{"cmd", "/c", "start", url}}
	default:
		return nil
	}
}
