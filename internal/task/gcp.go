package task

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/session"
)

type GCPCreateProject struct {
	BaseTask
}

func NewGCPCreateProject() *GCPCreateProject {
	return &GCPCreateProject{
		BaseTask: BaseTask{
			TaskName:        "gcp:create-project",
			TaskDescription: "Create GCP project",
		},
	}
}

func (t *GCPCreateProject) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil
}

func (t *GCPCreateProject) Run(sess *session.Session) error {
	PrintStep(fmt.Sprintf("Creating GCP project %s", sess.Config.GoogleCloudPlatform.ProjectName))
	cmds := t.Commands(sess)
	for _, cmd := range cmds {
		if err := sess.Exec.Run(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("gcp create-project failed: %w", err)
		}
	}
	return nil
}

func (t *GCPCreateProject) Commands(sess *session.Session) [][]string {
	if sess.Config.GoogleCloudPlatform == nil {
		return nil
	}
	return [][]string{
		{"gcloud", "projects", "create", sess.Config.GoogleCloudPlatform.ProjectName},
	}
}

type GCPInit struct {
	BaseTask
}

func NewGCPInit() *GCPInit {
	return &GCPInit{
		BaseTask: BaseTask{
			TaskName:        "gcp:init",
			TaskDescription: "Initialize Google Cloud SDK",
		},
	}
}

func (t *GCPInit) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil
}

func (t *GCPInit) Run(sess *session.Session) error {
	PrintStep("Initializing Google Cloud SDK")
	cmds := t.Commands(sess)
	for _, cmd := range cmds {
		if err := sess.Exec.Run(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("gcp init failed: %w", err)
		}
	}
	return nil
}

func (t *GCPInit) Commands(sess *session.Session) [][]string {
	if sess.Config.GoogleCloudPlatform == nil {
		return nil
	}
	cmds := [][]string{{"gcloud", "config", "set", "project", sess.Config.GoogleCloudPlatform.ProjectName}}
	if sess.Config.GoogleCloudPlatform.Account != "" {
		cmds = append(cmds, []string{"gcloud", "config", "set", "account", sess.Config.GoogleCloudPlatform.Account})
	}
	return cmds
}

type GCPActivate struct {
	BaseTask
}

func NewGCPActivate() *GCPActivate {
	return &GCPActivate{
		BaseTask: BaseTask{
			TaskName:        "gcp:activate",
			TaskDescription: "Activate GCP account and project",
			TaskDeps:        nil,
		},
	}
}

func (t *GCPActivate) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil
}

func (t *GCPActivate) Run(sess *session.Session) error {
	PrintStep(fmt.Sprintf("Activating project %s", sess.Config.GoogleCloudPlatform.ProjectName))
	cmds := t.Commands(sess)
	for _, cmd := range cmds {
		if err := sess.Exec.Run(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("gcp activate failed: %w", err)
		}
	}
	return nil
}

func (t *GCPActivate) Commands(sess *session.Session) [][]string {
	if sess.Config.GoogleCloudPlatform == nil {
		return nil
	}
	cmds := [][]string{
		{"gcloud", "config", "set", "project", sess.Config.GoogleCloudPlatform.ProjectName},
	}
	if sess.Config.GoogleCloudPlatform.Account != "" {
		cmds = append(cmds, []string{"gcloud", "config", "set", "account", sess.Config.GoogleCloudPlatform.Account})
	}
	return cmds
}

type GCPSetConfig struct {
	BaseTask
}

func NewGCPSetConfig() *GCPSetConfig {
	return &GCPSetConfig{
		BaseTask: BaseTask{
			TaskName:        "gcp:set-config",
			TaskDescription: "Set GCP project configuration",
		},
	}
}

func (t *GCPSetConfig) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil
}

func (t *GCPSetConfig) Requirements(sess *session.Session) []InputRequirement {
	var reqs []InputRequirement
	if sess.Config.GoogleCloudPlatform != nil && sess.Config.GoogleCloudPlatform.Account == "" {
		if _, ok := sess.Inputs["gcp:account"]; !ok {
			reqs = append(reqs, InputRequirement{
				Key:    "gcp:account",
				Prompt: "Enter GCP account email",
			})
		}
	}
	return reqs
}

func (t *GCPSetConfig) Run(sess *session.Session) error {
	PrintStep("Setting GCP project configuration")
	cmds := t.Commands(sess)
	for _, cmd := range cmds {
		if err := sess.Exec.Run(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("gcp set-config failed: %w", err)
		}
	}
	return nil
}

func (t *GCPSetConfig) Commands(sess *session.Session) [][]string {
	if sess.Config.GoogleCloudPlatform == nil {
		return nil
	}
	project := sess.Config.GoogleCloudPlatform.ProjectName
	account := sess.Config.GoogleCloudPlatform.Account
	if a, ok := sess.Inputs["gcp:account"]; ok && a != "" {
		account = a
	}

	cmds := [][]string{
		{"gcloud", "config", "set", "project", project},
	}
	if account != "" {
		cmds = append(cmds, []string{"gcloud", "config", "set", "account", account})
	}
	cmds = append(cmds, []string{"gcloud", "config", "set", "app/promote_by_default", "false"})
	cmds = append(cmds, []string{"gcloud", "config", "set", "billing/quota_project", project})
	return cmds
}

type GCPADCLogin struct {
	BaseTask
}

func NewGCPADCLogin() *GCPADCLogin {
	return &GCPADCLogin{
		BaseTask: BaseTask{
			TaskName:        "gcp:adc-login",
			TaskDescription: "Login to GCP and set Application Default Credentials",
		},
	}
}

func (t *GCPADCLogin) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil
}

func (t *GCPADCLogin) Run(sess *session.Session) error {
	PrintStep("Logging in to GCP")
	cmds := t.Commands(sess)
	for _, cmd := range cmds {
		if err := sess.Exec.Run(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("gcp adc-login failed: %w", err)
		}
	}
	return nil
}

func (t *GCPADCLogin) Commands(sess *session.Session) [][]string {
	if sess.Config.GoogleCloudPlatform == nil {
		return nil
	}
	project := sess.Config.GoogleCloudPlatform.ProjectName
	return [][]string{
		{"gcloud", "config", "configurations", "activate", project},
		{"gcloud", "auth", "application-default", "login", "--project", project},
		{"gcloud", "auth", "login", "--project", project},
		{"gcloud", "auth", "application-default", "set-quota-project", project},
	}
}

type GCPAdcImpersonateLogin struct {
	BaseTask
}

func NewGCPAdcImpersonateLogin() *GCPAdcImpersonateLogin {
	return &GCPAdcImpersonateLogin{
		BaseTask: BaseTask{
			TaskName:        "gcp:adc-impersonate-login",
			TaskDescription: "Login with Application Default Credentials and service account impersonation",
		},
	}
}

func (t *GCPAdcImpersonateLogin) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil && (sess.Config.GoogleCloudPlatform.ImpersonateServiceAccount != "" || sess.Inputs["gcp:impersonate-service-account"] != "")
}

func (t *GCPAdcImpersonateLogin) Requirements(sess *session.Session) []InputRequirement {
	var reqs []InputRequirement
	if sess.Config.GoogleCloudPlatform != nil && sess.Config.GoogleCloudPlatform.ImpersonateServiceAccount == "" {
		if _, ok := sess.Inputs["gcp:impersonate-service-account"]; !ok {
			reqs = append(reqs, InputRequirement{
				Key:    "gcp:impersonate-service-account",
				Prompt: "Enter service account to impersonate",
			})
		}
	}
	return reqs
}

func (t *GCPAdcImpersonateLogin) Run(sess *session.Session) error {
	sa := sess.Config.GoogleCloudPlatform.ImpersonateServiceAccount
	if a, ok := sess.Inputs["gcp:impersonate-service-account"]; ok && a != "" {
		sa = a
	}
	PrintStep(fmt.Sprintf("Logging in with impersonation: %s", sa))
	cmds := t.Commands(sess)
	for _, cmd := range cmds {
		if err := sess.Exec.Run(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("gcp adc-impersonate-login failed: %w", err)
		}
	}
	return nil
}

func (t *GCPAdcImpersonateLogin) Commands(sess *session.Session) [][]string {
	if sess.Config.GoogleCloudPlatform == nil {
		return nil
	}
	project := sess.Config.GoogleCloudPlatform.ProjectName
	sa := sess.Config.GoogleCloudPlatform.ImpersonateServiceAccount
	if a, ok := sess.Inputs["gcp:impersonate-service-account"]; ok && a != "" {
		sa = a
	}
	if sa == "" {
		return nil
	}

	return [][]string{
		{"gcloud", "config", "configurations", "activate", project},
		{"gcloud", "auth", "application-default", "login", "--impersonate-service-account", sa, "--project", project},
		{"gcloud", "auth", "login", "--impersonate-service-account", sa, "--project", project},
	}
}

type GCPConsole struct {
	BaseTask
}

func NewGCPConsole() *GCPConsole {
	return &GCPConsole{
		BaseTask: BaseTask{
			TaskName:        "gcp:console",
			TaskDescription: "Open Google Cloud Console",
		},
	}
}

func (t *GCPConsole) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil
}

func (t *GCPConsole) Run(sess *session.Session) error {
	PrintStep("Opening Google Cloud Console")
	cmds := t.Commands(sess)
	if err := sess.Exec.Run(cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("failed to open gcp console: %w", err)
	}
	return nil
}

func (t *GCPConsole) Commands(sess *session.Session) [][]string {
	if sess.Config.GoogleCloudPlatform == nil {
		return nil
	}
	url := fmt.Sprintf("https://console.cloud.google.com/home/dashboard?project=%s", sess.Config.GoogleCloudPlatform.ProjectName)
	return [][]string{{"open", url}}
}

type GCPAppEngineDeploy struct {
	BaseTask
	AppYaml string
}

func NewGCPAppEngineDeploy(appYaml string) *GCPAppEngineDeploy {
	return &GCPAppEngineDeploy{
		BaseTask: BaseTask{
			TaskName:        "gcp:app-engine-deploy",
			TaskDescription: "Deploy to App Engine",
		},
		AppYaml: appYaml,
	}
}

func (t *GCPAppEngineDeploy) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil && t.AppYaml != ""
}

func (t *GCPAppEngineDeploy) Requirements(sess *session.Session) []InputRequirement {
	return []InputRequirement{
		{
			Key:    "gcp:version",
			Prompt: "Enter version name, or return to skip",
		},
	}
}

func (t *GCPAppEngineDeploy) Run(sess *session.Session) error {
	PrintStep(fmt.Sprintf("Deploying %s to App Engine", t.AppYaml))
	version := sess.Inputs["gcp:version"]
	if version != "" {
		PrintSubStep(fmt.Sprintf("Version: %s", version))
	}
	cmds := t.Commands(sess)
	if err := sess.Exec.Run(cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("gcp app-engine deploy failed: %w", err)
	}
	return nil
}

func (t *GCPAppEngineDeploy) Commands(sess *session.Session) [][]string {
	cmd := []string{"gcloud", "app", "deploy", t.AppYaml}
	version := sess.Inputs["gcp:version"]
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
	desc := "Promote a version to receive all traffic"
	name := "gcp:app-engine-promote"
	if service != "" {
		desc = fmt.Sprintf("Promote a version to receive all traffic for service %s", service)
		name = fmt.Sprintf("gcp:app-engine-promote:%s", service)
	}
	return &GCPAppEnginePromote{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: desc,
		},
		Service: service,
	}
}

func (t *GCPAppEnginePromote) ShouldRun(sess *session.Session) bool {
	return sess.Config.GoogleCloudPlatform != nil
}

func (t *GCPAppEnginePromote) Requirements(sess *session.Session) []InputRequirement {
	return []InputRequirement{
		{
			Key:    "gcp:promote_version",
			Prompt: "Enter version to promote",
		},
	}
}

func (t *GCPAppEnginePromote) Run(sess *session.Session) error {
	version := sess.Inputs["gcp:promote_version"]
	if version == "" {
		return fmt.Errorf("no version specified for promotion")
	}
	PrintStep(fmt.Sprintf("Promoting App Engine version %s", version))
	cmds := t.Commands(sess)
	if err := sess.Exec.Run(cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("gcp app-engine promote failed for version %s: %w", version, err)
	}
	return nil
}

func (t *GCPAppEnginePromote) Commands(sess *session.Session) [][]string {
	version := sess.Inputs["gcp:promote_version"]
	if version == "" {
		return nil
	}
	cmd := []string{"gcloud", "app", "versions", "migrate", version}
	if t.Service != "" {
		cmd = append(cmd, "--service", t.Service)
	}
	return [][]string{cmd}
}
