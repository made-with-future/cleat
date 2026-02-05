package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/session"
)

type DjangoRunServer struct {
	BaseTask
	Service *config.ServiceConfig
}


func NewDjangoRunServer(svc *config.ServiceConfig) *DjangoRunServer {
	return &DjangoRunServer{
		BaseTask: BaseTask{
			TaskName:        "django:runserver",
			TaskDescription: "Run Django development server",
		},
		Service: svc,
	}
}

func (t *DjangoRunServer) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		for _, mod := range t.Service.Modules {
			if mod.Python != nil && mod.Python.Django {
				return true
			}
		}
	}
	return false
}

func (t *DjangoRunServer) Run(sess *session.Session) error {
	if sess.Config.Docker && t.Service.IsDocker() {
		pyConfig := getPythonConfig(t.Service)
		fmt.Printf("==> Running Django runserver for service %s via Docker (%s service)\n", t.Service.Name, pyConfig.DjangoService)
	} else {
		fmt.Printf("==> Running Django server for service %s\n", t.Service.Name)
	}
	cmds := t.Commands(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() {
		dir = "" // Run from root when using docker compose
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("django server failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *DjangoRunServer) Commands(sess *session.Session) [][]string {
	pyConfig := getPythonConfig(t.Service)
	if sess.Config.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", pyConfig.DjangoService}
		cmd = append(cmd, pythonCommand(pyConfig)...)
		cmd = append(cmd, "manage.py", "runserver", "0.0.0.0:8000")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(pyConfig)
	cmd = append(cmd, managePy, "runserver")
	return [][]string{cmd}
}

type DjangoMigrate struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDjangoMigrate(svc *config.ServiceConfig) *DjangoMigrate {
	return &DjangoMigrate{
		BaseTask: BaseTask{
			TaskName:        "django:migrate",
			TaskDescription: "Run Django database migrations",
		},
		Service: svc,
	}
}

func (t *DjangoMigrate) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		for _, mod := range t.Service.Modules {
			if mod.Python != nil && mod.Python.Django {
				return true
			}
		}
	}
	return false
}

func (t *DjangoMigrate) Run(sess *session.Session) error {
	if sess.Config.Docker && t.Service.IsDocker() {
		pyConfig := getPythonConfig(t.Service)
		fmt.Printf("==> Running Django migrations for service %s via Docker (%s service)\n", t.Service.Name, pyConfig.DjangoService)
	} else {
		fmt.Printf("==> Running Django migrations for service %s\n", t.Service.Name)
	}
	cmds := t.Commands(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() {
		dir = ""
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("django migrations failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *DjangoMigrate) Commands(sess *session.Session) [][]string {
	pyConfig := getPythonConfig(t.Service)
	if sess.Config.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", pyConfig.DjangoService}
		cmd = append(cmd, pythonCommand(pyConfig)...)
		cmd = append(cmd, "manage.py", "migrate", "--noinput")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(pyConfig)
	cmd = append(cmd, managePy, "migrate", "--noinput")
	return [][]string{cmd}
}

type DjangoMakeMigrations struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDjangoMakeMigrations(svc *config.ServiceConfig) *DjangoMakeMigrations {
	return &DjangoMakeMigrations{
		BaseTask: BaseTask{
			TaskName:        "django:makemigrations",
			TaskDescription: "Create new Django migrations",
		},
		Service: svc,
	}
}

func (t *DjangoMakeMigrations) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		for _, mod := range t.Service.Modules {
			if mod.Python != nil && mod.Python.Django {
				return true
			}
		}
	}
	return false
}

func (t *DjangoMakeMigrations) Run(sess *session.Session) error {
	if sess.Config.Docker && t.Service.IsDocker() {
		pyConfig := getPythonConfig(t.Service)
		fmt.Printf("==> Creating Django migrations for service %s via Docker (%s service)\n", t.Service.Name, pyConfig.DjangoService)
	} else {
		fmt.Printf("==> Creating Django migrations for service %s\n", t.Service.Name)
	}
	cmds := t.Commands(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() {
		dir = ""
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("django makemigrations failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *DjangoMakeMigrations) Commands(sess *session.Session) [][]string {
	pyConfig := getPythonConfig(t.Service)
	if sess.Config.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", pyConfig.DjangoService}
		cmd = append(cmd, pythonCommand(pyConfig)...)
		cmd = append(cmd, "manage.py", "makemigrations")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(pyConfig)
	cmd = append(cmd, managePy, "makemigrations")
	return [][]string{cmd}
}

type DjangoCollectStatic struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDjangoCollectStatic(svc *config.ServiceConfig) *DjangoCollectStatic {
	return &DjangoCollectStatic{
		BaseTask: BaseTask{
			TaskName:        "django:collectstatic",
			TaskDescription: "Collect Django static files",
		},
		Service: svc,
	}
}

func (t *DjangoCollectStatic) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		for _, mod := range t.Service.Modules {
			if mod.Python != nil && mod.Python.Django {
				return true
			}
		}
	}
	return false
}

func (t *DjangoCollectStatic) Run(sess *session.Session) error {
	if sess.Config.Docker && t.Service.IsDocker() {
		pyConfig := getPythonConfig(t.Service)
		fmt.Printf("==> Collecting static files for service %s via Docker (%s service)\n", t.Service.Name, pyConfig.DjangoService)
	} else {
		fmt.Printf("==> Collecting static files for service %s\n", t.Service.Name)
	}
	cmds := t.Commands(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() {
		dir = ""
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("django collectstatic failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *DjangoCollectStatic) Commands(sess *session.Session) [][]string {
	pyConfig := getPythonConfig(t.Service)
	if sess.Config.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", pyConfig.DjangoService}
		cmd = append(cmd, pythonCommand(pyConfig)...)
		cmd = append(cmd, "manage.py", "collectstatic", "--noinput", "--clear")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(pyConfig)
	cmd = append(cmd, managePy, "collectstatic", "--noinput", "--clear")
	return [][]string{cmd}
}

type DjangoCreateUserDev struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDjangoCreateUserDev(svc *config.ServiceConfig) *DjangoCreateUserDev {
	return &DjangoCreateUserDev{
		BaseTask: BaseTask{
			TaskName:        "django:create-user-dev",
			TaskDescription: "Create a development superuser in Django",
		},
		Service: svc,
	}
}

func (t *DjangoCreateUserDev) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		for _, mod := range t.Service.Modules {
			if mod.Python != nil && mod.Python.Django {
				return true
			}
		}
	}
	return false
}

func (t *DjangoCreateUserDev) Run(sess *session.Session) error {
	if sess.Config.Docker && t.Service.IsDocker() {
		pyConfig := getPythonConfig(t.Service)
		fmt.Printf("==> Creating Django superuser for service %s via Docker (%s service)\n", t.Service.Name, pyConfig.DjangoService)
	} else {
		fmt.Printf("==> Creating Django superuser for service %s\n", t.Service.Name)
	}
	cmds := t.Commands(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() {
		dir = ""
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("django create-user-dev failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *DjangoCreateUserDev) Commands(sess *session.Session) [][]string {
	pyConfig := getPythonConfig(t.Service)
	if sess.Config.Docker && t.Service.IsDocker() {
		cmd := []string{
			"docker", "--log-level", "error", "compose", "run",
			"-e", "DJANGO_SUPERUSER_USERNAME=dev",
			"-e", "DJANGO_SUPERUSER_PASSWORD=dev",
			"--rm",
			pyConfig.DjangoService,
		}
		cmd = append(cmd, pythonCommand(pyConfig)...)
		cmd = append(cmd, "manage.py", "createsuperuser", "--email", "dev@madewithfuture.com", "--noinput")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(pyConfig)
	cmd = append(cmd, managePy, "shell", "-c", "from django.contrib.auth.models import User; User.objects.create_superuser('admin', 'admin@example.com', 'admin') if not User.objects.filter(username='admin').exists() else None")
	return [][]string{cmd}
}

type DjangoGenRandomSecretKey struct {
	BaseTask
	Service *config.ServiceConfig
}

func NewDjangoGenRandomSecretKey(svc *config.ServiceConfig) *DjangoGenRandomSecretKey {
	return &DjangoGenRandomSecretKey{
		BaseTask: BaseTask{
			TaskName:        "django:gen-random-secret-key",
			TaskDescription: "Generate a random Django secret key",
		},
		Service: svc,
	}
}

func (t *DjangoGenRandomSecretKey) ShouldRun(sess *session.Session) bool {
	if t.Service != nil {
		for _, mod := range t.Service.Modules {
			if mod.Python != nil && mod.Python.Django {
				return true
			}
		}
	}
	return false
}

func (t *DjangoGenRandomSecretKey) Run(sess *session.Session) error {
	if sess.Config.Docker && t.Service.IsDocker() {
		pyConfig := getPythonConfig(t.Service)
		fmt.Printf("==> Generating Django secret key for service %s via Docker (%s service)\n", t.Service.Name, pyConfig.DjangoService)
	} else {
		fmt.Printf("==> Generating Django secret key for service %s\n", t.Service.Name)
	}
	cmds := t.Commands(sess)
	dir := t.Service.Dir
	if sess.Config.Docker && t.Service.IsDocker() {
		dir = ""
	}
	if err := sess.Exec.RunWithDir(dir, cmds[0][0], cmds[0][1:]...); err != nil {
		return fmt.Errorf("django gen-random-secret-key failed for service %s: %w", t.Service.Name, err)
	}
	return nil
}

func (t *DjangoGenRandomSecretKey) Commands(sess *session.Session) [][]string {
	pyCmd := "from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())"
	pyConfig := getPythonConfig(t.Service)

	if sess.Config.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", pyConfig.DjangoService}
		cmd = append(cmd, pythonCommand(pyConfig)...)
		cmd = append(cmd, "-c", pyCmd)
		return [][]string{cmd}
	}

	cmd := pythonCommand(pyConfig)
	cmd = append(cmd, "-c", pyCmd)
	return [][]string{cmd}
}

func getPythonConfig(svc *config.ServiceConfig) *config.PythonConfig {
	if svc == nil {
		return nil
	}
	for _, mod := range svc.Modules {
		if mod.Python != nil {
			return mod.Python
		}
	}
	return nil
}

func pythonCommand(p *config.PythonConfig) []string {
	if p != nil && p.PackageManager == "pip" {
		return []string{"python"}
	}
	return []string{"uv", "run", "python"}
}

func findManagePy(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, "backend/manage.py")); err == nil {
		return "backend/manage.py"
	}
	return "manage.py"
}
