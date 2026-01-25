package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// DjangoCollectStatic runs Django's collectstatic command
type DjangoCollectStatic struct {
	BaseTask
	Service *config.ServiceConfig
	Python  *config.PythonConfig
}

func NewDjangoCollectStatic(svc *config.ServiceConfig, python *config.PythonConfig) *DjangoCollectStatic {
	name := "django:collectstatic"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("django:collectstatic:%s", svc.Name)
	}
	return &DjangoCollectStatic{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Collect Django static files",
			TaskDeps:        []string{"docker:build", "npm:build"}, // Static files often come from npm build
		},
		Service: svc,
		Python:  python,
	}
}

func (t *DjangoCollectStatic) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Python != nil && t.Python.Django
}

func (t *DjangoCollectStatic) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Collecting Django static files for service '%s'\n", t.Service.Name)

	if cfg.Docker && t.Service.IsDocker() {
		fmt.Printf("--> Running collectstatic via Docker (%s service)\n", t.Python.DjangoService)
	} else {
		fmt.Println("--> Running collectstatic locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCollectStatic) Commands(cfg *config.Config) [][]string {
	if cfg.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.Python.DjangoService}
		cmd = append(cmd, pythonCommand(t.Python)...)
		cmd = append(cmd, "manage.py", "collectstatic", "--noinput", "--clear")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(t.Python)
	cmd = append(cmd, managePy, "collectstatic", "--noinput", "--clear")
	return [][]string{cmd}
}

// DjangoRunServer runs Django's development server
type DjangoRunServer struct {
	BaseTask
	Service *config.ServiceConfig
	Python  *config.PythonConfig
}

func NewDjangoRunServer(svc *config.ServiceConfig, python *config.PythonConfig) *DjangoRunServer {
	name := "django:runserver"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("django:runserver:%s", svc.Name)
	}
	return &DjangoRunServer{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Start Django development server",
			TaskDeps:        nil,
		},
		Service: svc,
		Python:  python,
	}
}

func (t *DjangoRunServer) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Python != nil && t.Python.Django
}

func (t *DjangoRunServer) Run(cfg *config.Config, exec executor.Executor) error {
	if cfg.Docker && t.Service.IsDocker() {
		fmt.Printf("==> Running Django runserver for service '%s' via Docker (%s service)\n", t.Service.Name, t.Python.DjangoService)
	} else {
		fmt.Printf("==> Running Django runserver for service '%s' locally\n", t.Service.Name)
	}
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoRunServer) Commands(cfg *config.Config) [][]string {
	if cfg.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.Python.DjangoService}
		cmd = append(cmd, pythonCommand(t.Python)...)
		cmd = append(cmd, "manage.py", "runserver", "0.0.0.0:8000")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(t.Python)
	cmd = append(cmd, managePy, "runserver")
	return [][]string{cmd}
}

func pythonCommand(p *config.PythonConfig) []string {
	if p != nil && p.PackageManager == "pip" {
		return []string{"python"}
	}
	return []string{"uv", "run", "python"}
}

func findManagePy(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, "backend/manage.py")); err == nil {
		return filepath.Join(dir, "backend/manage.py")
	}
	return filepath.Join(dir, "manage.py")
}

// DjangoCreateUserDev creates a Django superuser for development
type DjangoCreateUserDev struct {
	BaseTask
	Service *config.ServiceConfig
	Python  *config.PythonConfig
}

func NewDjangoCreateUserDev(svc *config.ServiceConfig, python *config.PythonConfig) *DjangoCreateUserDev {
	name := "django:create-user-dev"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("django:create-user-dev:%s", svc.Name)
	}
	return &DjangoCreateUserDev{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Create a Django superuser (dev/dev)",
			TaskDeps:        nil,
		},
		Service: svc,
		Python:  python,
	}
}

func (t *DjangoCreateUserDev) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Python != nil && t.Python.Django && cfg.Docker && t.Service.IsDocker()
}

func (t *DjangoCreateUserDev) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Creating Django dev superuser for service '%s'\n", t.Service.Name)
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCreateUserDev) Commands(cfg *config.Config) [][]string {
	cmd := []string{
		"docker", "--log-level", "error", "compose", "run",
		"-e", "DJANGO_SUPERUSER_USERNAME=dev",
		"-e", "DJANGO_SUPERUSER_PASSWORD=dev",
		"--rm",
		t.Python.DjangoService,
	}
	cmd = append(cmd, pythonCommand(t.Python)...)
	cmd = append(cmd,
		"manage.py", "createsuperuser",
		"--email", "dev@madewithfuture.com",
		"--noinput",
	)
	return [][]string{cmd}
}

// DjangoMigrate runs Django migrations
type DjangoMigrate struct {
	BaseTask
	Service *config.ServiceConfig
	Python  *config.PythonConfig
}

func NewDjangoMigrate(svc *config.ServiceConfig, python *config.PythonConfig) *DjangoMigrate {
	name := "django:migrate"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("django:migrate:%s", svc.Name)
	}
	return &DjangoMigrate{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Run Django migrations",
			TaskDeps:        []string{"docker:build"},
		},
		Service: svc,
		Python:  python,
	}
}

func (t *DjangoMigrate) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Python != nil && t.Python.Django
}

func (t *DjangoMigrate) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Running Django migrations for service '%s'\n", t.Service.Name)

	if cfg.Docker && t.Service.IsDocker() {
		fmt.Printf("--> Running migrate via Docker (%s service)\n", t.Python.DjangoService)
	} else {
		fmt.Println("--> Running migrate locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoMigrate) Commands(cfg *config.Config) [][]string {
	if cfg.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.Python.DjangoService}
		cmd = append(cmd, pythonCommand(t.Python)...)
		cmd = append(cmd, "manage.py", "migrate", "--noinput")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(t.Python)
	cmd = append(cmd, managePy, "migrate", "--noinput")
	return [][]string{cmd}
}

// DjangoMakeMigrations runs Django's makemigrations command
type DjangoMakeMigrations struct {
	BaseTask
	Service *config.ServiceConfig
	Python  *config.PythonConfig
}

func NewDjangoMakeMigrations(svc *config.ServiceConfig, python *config.PythonConfig) *DjangoMakeMigrations {
	name := "django:makemigrations"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("django:makemigrations:%s", svc.Name)
	}
	return &DjangoMakeMigrations{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Create new migrations based on model changes",
			TaskDeps:        []string{"docker:build"},
		},
		Service: svc,
		Python:  python,
	}
}

func (t *DjangoMakeMigrations) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Python != nil && t.Python.Django
}

func (t *DjangoMakeMigrations) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Running Django makemigrations for service '%s'\n", t.Service.Name)

	if cfg.Docker && t.Service.IsDocker() {
		fmt.Printf("--> Running makemigrations via Docker (%s service)\n", t.Python.DjangoService)
	} else {
		fmt.Println("--> Running makemigrations locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoMakeMigrations) Commands(cfg *config.Config) [][]string {
	if cfg.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.Python.DjangoService}
		cmd = append(cmd, pythonCommand(t.Python)...)
		cmd = append(cmd, "manage.py", "makemigrations")
		return [][]string{cmd}
	}

	managePy := findManagePy(t.Service.Dir)
	cmd := pythonCommand(t.Python)
	cmd = append(cmd, managePy, "makemigrations")
	return [][]string{cmd}
}

// DjangoGenRandomSecretKey generates a random Django secret key
type DjangoGenRandomSecretKey struct {
	BaseTask
	Service *config.ServiceConfig
	Python  *config.PythonConfig
}

func NewDjangoGenRandomSecretKey(svc *config.ServiceConfig, python *config.PythonConfig) *DjangoGenRandomSecretKey {
	name := "django:gen-random-secret-key"
	if svc != nil && svc.Name != "default" {
		name = fmt.Sprintf("django:gen-random-secret-key:%s", svc.Name)
	}
	return &DjangoGenRandomSecretKey{
		BaseTask: BaseTask{
			TaskName:        name,
			TaskDescription: "Generate a random Django secret key",
			TaskDeps:        nil,
		},
		Service: svc,
		Python:  python,
	}
}

func (t *DjangoGenRandomSecretKey) ShouldRun(cfg *config.Config) bool {
	return t.Service != nil && t.Python != nil && t.Python.Django
}

func (t *DjangoGenRandomSecretKey) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Generating random Django secret key for service '%s'\n", t.Service.Name)

	if cfg.Docker && t.Service.IsDocker() {
		fmt.Printf("--> Running via Docker (%s service)\n", t.Python.DjangoService)
	} else {
		fmt.Println("--> Running locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoGenRandomSecretKey) Commands(cfg *config.Config) [][]string {
	pyCmd := "from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())"

	if cfg.Docker && t.Service.IsDocker() {
		cmd := []string{"docker", "--log-level", "error", "compose", "run", "--rm", t.Python.DjangoService}
		cmd = append(cmd, pythonCommand(t.Python)...)
		cmd = append(cmd, "-c", pyCmd)
		return [][]string{cmd}
	}

	cmd := pythonCommand(t.Python)
	cmd = append(cmd, "-c", pyCmd)
	return [][]string{cmd}
}
