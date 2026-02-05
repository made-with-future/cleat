package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func NewDjangoRunServerStrategy(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("django runserver", []task.Task{
		task.NewDjangoRunServer(svc),
	})
}

func NewDjangoRunServerStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			if svc.Modules[j].Python != nil && svc.Modules[j].Python.Django {
				tasks = append(tasks, task.NewDjangoRunServer(svc))
				// Just the first one for global strategy to avoid conflicts
				return NewBaseStrategy("django runserver", tasks)
			}
		}
	}
	return NewBaseStrategy("django runserver", tasks)
}

func NewDjangoMigrateStrategy(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("django migrate", []task.Task{
		task.NewDjangoMigrate(svc),
	})
}

func NewDjangoMigrateStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			if svc.Modules[j].Python != nil && svc.Modules[j].Python.Django {
				tasks = append(tasks, task.NewDjangoMigrate(svc))
				return NewBaseStrategy("django migrate", tasks)
			}
		}
	}
	return NewBaseStrategy("django migrate", tasks)
}

func NewDjangoMakeMigrationsStrategy(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("django makemigrations", []task.Task{
		task.NewDjangoMakeMigrations(svc),
	})
}

func NewDjangoMakeMigrationsStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			if svc.Modules[j].Python != nil && svc.Modules[j].Python.Django {
				tasks = append(tasks, task.NewDjangoMakeMigrations(svc))
				return NewBaseStrategy("django makemigrations", tasks)
			}
		}
	}
	return NewBaseStrategy("django makemigrations", tasks)
}

func NewDjangoCollectStaticStrategy(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("django collectstatic", []task.Task{
		task.NewDjangoCollectStatic(svc),
	})
}

func NewDjangoCollectStaticStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			if svc.Modules[j].Python != nil && svc.Modules[j].Python.Django {
				tasks = append(tasks, task.NewDjangoCollectStatic(svc))
				return NewBaseStrategy("django collectstatic", tasks)
			}
		}
	}
	return NewBaseStrategy("django collectstatic", tasks)
}

func NewDjangoCreateUserDevStrategy(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("django create-user-dev", []task.Task{
		task.NewDjangoCreateUserDev(svc),
	})
}

func NewDjangoCreateUserDevStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			if svc.Modules[j].Python != nil && svc.Modules[j].Python.Django {
				tasks = append(tasks, task.NewDjangoCreateUserDev(svc))
				return NewBaseStrategy("django create-user-dev", tasks)
			}
		}
	}
	return NewBaseStrategy("django create-user-dev", tasks)
}

func NewDjangoGenRandomSecretKeyStrategy(svc *config.ServiceConfig) Strategy {
	return NewBaseStrategy("django gen-random-secret-key", []task.Task{
		task.NewDjangoGenRandomSecretKey(svc),
	})
}

func NewDjangoGenRandomSecretKeyStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			if svc.Modules[j].Python != nil && svc.Modules[j].Python.Django {
				tasks = append(tasks, task.NewDjangoGenRandomSecretKey(svc))
				return NewBaseStrategy("django gen-random-secret-key", tasks)
			}
		}
	}
	return NewBaseStrategy("django gen-random-secret-key", tasks)
}