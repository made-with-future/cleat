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

	if cfg.Docker {
		fmt.Printf("--> Running collectstatic via Docker (%s service)\n", t.Python.DjangoService)
	} else {
		fmt.Println("--> Running collectstatic locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCollectStatic) Commands(cfg *config.Config) [][]string {
	if cfg.Docker {
		return [][]string{{"docker", "compose", "run", "--rm", t.Python.DjangoService,
			"uv", "run", "python", "manage.py", "collectstatic", "--noinput", "--clear"}}
	}

	managePy := findManagePy(t.Service.Location)
	return [][]string{{"uv", "run", "python", managePy, "collectstatic", "--noinput", "--clear"}}
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
	return t.Service != nil && t.Python != nil && t.Python.Django && !cfg.Docker
}

func (t *DjangoRunServer) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Running Django project '%s' locally\n", t.Service.Name)
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoRunServer) Commands(cfg *config.Config) [][]string {
	managePy := findManagePy(t.Service.Location)
	return [][]string{{"uv", "run", "python", managePy, "runserver"}}
}

func findManagePy(location string) string {
	if _, err := os.Stat(filepath.Join(location, "backend/manage.py")); err == nil {
		return filepath.Join(location, "backend/manage.py")
	}
	return filepath.Join(location, "manage.py")
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
	return t.Service != nil && t.Python != nil && t.Python.Django && cfg.Docker
}

func (t *DjangoCreateUserDev) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Printf("==> Creating Django dev superuser for service '%s'\n", t.Service.Name)
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCreateUserDev) Commands(cfg *config.Config) [][]string {
	return [][]string{{
		"docker", "compose", "run",
		"-e", "DJANGO_SUPERUSER_USERNAME=dev",
		"-e", "DJANGO_SUPERUSER_PASSWORD=dev",
		"--rm",
		t.Python.DjangoService,
		"uv", "run", "python", "manage.py", "createsuperuser",
		"--email", "dev@madewithfuture.com",
		"--noinput",
	}}
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

	if cfg.Docker {
		fmt.Printf("--> Running migrate via Docker (%s service)\n", t.Python.DjangoService)
	} else {
		fmt.Println("--> Running migrate locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoMigrate) Commands(cfg *config.Config) [][]string {
	if cfg.Docker {
		return [][]string{{"docker", "compose", "run", "--rm", t.Python.DjangoService,
			"uv", "run", "python", "manage.py", "migrate", "--noinput"}}
	}

	managePy := findManagePy(t.Service.Location)
	return [][]string{{"uv", "run", "python", managePy, "migrate", "--noinput"}}
}
