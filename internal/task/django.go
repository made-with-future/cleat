package task

import (
	"fmt"

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
	fmt.Printf("==> Running Django server for service %s\n", t.Service.Name)
	cmds := t.Commands(sess)
	return sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoRunServer) Commands(sess *session.Session) [][]string {
	pm := "uv"
	for _, mod := range t.Service.Modules {
		if mod.Python != nil && mod.Python.PackageManager != "" {
			pm = mod.Python.PackageManager
			break
		}
	}
	return [][]string{{pm, "run", "python", "manage.py", "runserver"}}
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
	fmt.Printf("==> Running Django migrations for service %s\n", t.Service.Name)
	cmds := t.Commands(sess)
	return sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoMigrate) Commands(sess *session.Session) [][]string {
	pm := "uv"
	for _, mod := range t.Service.Modules {
		if mod.Python != nil && mod.Python.PackageManager != "" {
			pm = mod.Python.PackageManager
			break
		}
	}
	return [][]string{{pm, "run", "python", "manage.py", "migrate"}}
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
	fmt.Printf("==> Creating Django migrations for service %s\n", t.Service.Name)
	cmds := t.Commands(sess)
	return sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoMakeMigrations) Commands(sess *session.Session) [][]string {
	pm := "uv"
	for _, mod := range t.Service.Modules {
		if mod.Python != nil && mod.Python.PackageManager != "" {
			pm = mod.Python.PackageManager
			break
		}
	}
	return [][]string{{pm, "run", "python", "manage.py", "makemigrations"}}
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
	fmt.Printf("==> Collecting static files for service %s\n", t.Service.Name)
	cmds := t.Commands(sess)
	return sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCollectStatic) Commands(sess *session.Session) [][]string {
	pm := "uv"
	for _, mod := range t.Service.Modules {
		if mod.Python != nil && mod.Python.PackageManager != "" {
			pm = mod.Python.PackageManager
			break
		}
	}
	return [][]string{{pm, "run", "python", "manage.py", "collectstatic", "--noinput"}}
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
	fmt.Printf("==> Creating Django superuser for service %s\n", t.Service.Name)
	cmds := t.Commands(sess)
	return sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoCreateUserDev) Commands(sess *session.Session) [][]string {
	pm := "uv"
	for _, mod := range t.Service.Modules {
		if mod.Python != nil && mod.Python.PackageManager != "" {
			pm = mod.Python.PackageManager
			break
		}
	}
	return [][]string{{pm, "run", "python", "manage.py", "shell", "-c", "from django.contrib.auth.models import User; User.objects.create_superuser('admin', 'admin@example.com', 'admin') if not User.objects.filter(username='admin').exists() else None"}}
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
	fmt.Printf("==> Generating Django secret key for service %s\n", t.Service.Name)
	cmds := t.Commands(sess)
	return sess.Exec.RunWithDir(t.Service.Dir, cmds[0][0], cmds[0][1:]...)
}

func (t *DjangoGenRandomSecretKey) Commands(sess *session.Session) [][]string {
	pm := "uv"
	for _, mod := range t.Service.Modules {
		if mod.Python != nil && mod.Python.PackageManager != "" {
			pm = mod.Python.PackageManager
			break
		}
	}
	return [][]string{{pm, "run", "python", "-c", "from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())"}}
}