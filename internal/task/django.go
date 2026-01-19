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
		return exec.Run("docker", "compose", "run", "--rm", cfg.DjangoService,
			"python", "manage.py", "collectstatic", "--noinput")
	}

	fmt.Println("--> Running collectstatic locally")
	managePy := findManagePy()
	return exec.Run("python", managePy, "collectstatic", "--noinput")
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
	managePy := findManagePy()
	return exec.Run("python", managePy, "runserver")
}

func findManagePy() string {
	if _, err := os.Stat("backend/manage.py"); err == nil {
		return "backend/manage.py"
	}
	return "manage.py"
}
