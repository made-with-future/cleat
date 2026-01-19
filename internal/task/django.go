package task

import (
	"fmt"
	"os"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/executor"
)

// DjangoCollectStatic runs Django's collectstatic command
type DjangoCollectStatic struct{ BaseTask }

func NewDjangoCollectStatic() *DjangoCollectStatic {
	return &DjangoCollectStatic{
		BaseTask: BaseTask{
			TaskName:        "django:collectstatic",
			TaskDescription: "Collect Django static files",
			TaskDeps:        []string{"docker:build", "npm:build"}, // Static files often come from npm build
		},
	}
}

func (t *DjangoCollectStatic) ShouldRun(cfg *config.Config) bool {
	return cfg.Django
}

func (t *DjangoCollectStatic) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Collecting Django static files")

	if cfg.Docker {
		fmt.Printf("--> Running collectstatic via Docker (%s service)\n", cfg.DjangoService)
	} else {
		fmt.Println("--> Running collectstatic locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCollectStatic) Commands(cfg *config.Config) [][]string {
	if cfg.Docker {
		return [][]string{{"docker", "compose", "run", "--rm", cfg.DjangoService,
			"uv", "run", "python", "manage.py", "collectstatic", "--noinput", "--clear"}}
	}

	managePy := findManagePy()
	return [][]string{{"uv", "run", "python", managePy, "collectstatic", "--noinput", "--clear"}}
}

// DjangoRunServer runs Django's development server
type DjangoRunServer struct{ BaseTask }

func NewDjangoRunServer() *DjangoRunServer {
	return &DjangoRunServer{
		BaseTask: BaseTask{
			TaskName:        "django:runserver",
			TaskDescription: "Start Django development server",
			TaskDeps:        nil,
		},
	}
}

func (t *DjangoRunServer) ShouldRun(cfg *config.Config) bool {
	return cfg.Django && !cfg.Docker
}

func (t *DjangoRunServer) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Running Django project locally")
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoRunServer) Commands(cfg *config.Config) [][]string {
	managePy := findManagePy()
	return [][]string{{"uv", "run", "python", managePy, "runserver"}}
}

func findManagePy() string {
	if _, err := os.Stat("backend/manage.py"); err == nil {
		return "backend/manage.py"
	}
	return "manage.py"
}

// DjangoCreateUserDev creates a Django superuser for development
type DjangoCreateUserDev struct{ BaseTask }

func NewDjangoCreateUserDev() *DjangoCreateUserDev {
	return &DjangoCreateUserDev{
		BaseTask: BaseTask{
			TaskName:        "django:create-user-dev",
			TaskDescription: "Create a Django superuser (dev/dev)",
			TaskDeps:        nil,
		},
	}
}

func (t *DjangoCreateUserDev) ShouldRun(cfg *config.Config) bool {
	return cfg.Django && cfg.Docker
}

func (t *DjangoCreateUserDev) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Creating Django dev superuser")
	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCreateUserDev) Commands(cfg *config.Config) [][]string {
	return [][]string{{
		"docker", "compose", "run",
		"-e", "DJANGO_SUPERUSER_USERNAME=dev",
		"-e", "DJANGO_SUPERUSER_PASSWORD=dev",
		"--rm",
		cfg.DjangoService,
		"uv", "run", "python", "manage.py", "createsuperuser",
		"--email", "dev@madewithfuture.com",
		"--noinput",
	}}
}

// DjangoMigrate runs Django migrations
type DjangoMigrate struct{ BaseTask }

func NewDjangoMigrate() *DjangoMigrate {
	return &DjangoMigrate{
		BaseTask: BaseTask{
			TaskName:        "django:migrate",
			TaskDescription: "Run Django migrations",
			TaskDeps:        []string{"docker:build"},
		},
	}
}

func (t *DjangoMigrate) ShouldRun(cfg *config.Config) bool {
	return cfg.Django
}

func (t *DjangoMigrate) Run(cfg *config.Config, exec executor.Executor) error {
	fmt.Println("==> Running Django migrations")

	if cfg.Docker {
		fmt.Printf("--> Running migrate via Docker (%s service)\n", cfg.DjangoService)
	} else {
		fmt.Println("--> Running migrate locally")
	}

	cmds := t.Commands(cfg)
	return exec.Run(cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoMigrate) Commands(cfg *config.Config) [][]string {
	if cfg.Docker {
		return [][]string{{"docker", "compose", "run", "--rm", cfg.DjangoService,
			"uv", "run", "python", "manage.py", "migrate", "--noinput"}}
	}

	managePy := findManagePy()
	return [][]string{{"uv", "run", "python", managePy, "migrate", "--noinput"}}
}
