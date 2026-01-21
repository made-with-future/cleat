package strategy

import (
	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

func init() {
	Register("django create-user-dev", NewDjangoCreateUserDevStrategyGlobal)
	Register("django collectstatic", NewDjangoCollectStaticStrategyGlobal)
	Register("django migrate", NewDjangoMigrateStrategyGlobal)
	Register("django makemigrations", NewDjangoMakeMigrationsStrategyGlobal)
	Register("django gen-random-secret-key", NewDjangoGenRandomSecretKeyStrategyGlobal)
}

func NewDjangoCreateUserDevStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	if cfg != nil {
		for i := range cfg.Services {
			svc := &cfg.Services[i]
			for j := range svc.Modules {
				mod := &svc.Modules[j]
				if mod.Python != nil {
					tasks = append(tasks, task.NewDjangoCreateUserDev(svc, mod.Python))
				}
			}
		}
	}
	return NewBaseStrategy("django create-user-dev", tasks)
}

func NewDjangoCreateUserDevStrategy(svc *config.ServiceConfig) Strategy {
	var tasks []task.Task
	for i := range svc.Modules {
		mod := &svc.Modules[i]
		if mod.Python != nil {
			tasks = append(tasks, task.NewDjangoCreateUserDev(svc, mod.Python))
		}
	}
	return NewBaseStrategy("django create-user-dev", tasks)
}

func NewDjangoCollectStaticStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Python != nil {
				tasks = append(tasks, task.NewDjangoCollectStatic(svc, mod.Python))
			}
		}
	}
	return NewBaseStrategy("django collectstatic", tasks)
}

func NewDjangoCollectStaticStrategy(svc *config.ServiceConfig) Strategy {
	var tasks []task.Task
	for i := range svc.Modules {
		mod := &svc.Modules[i]
		if mod.Python != nil {
			tasks = append(tasks, task.NewDjangoCollectStatic(svc, mod.Python))
		}
	}
	return NewBaseStrategy("django collectstatic", tasks)
}

func NewDjangoMigrateStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Python != nil {
				tasks = append(tasks, task.NewDjangoMigrate(svc, mod.Python))
			}
		}
	}
	return NewBaseStrategy("django migrate", tasks)
}

func NewDjangoMigrateStrategy(svc *config.ServiceConfig) Strategy {
	var tasks []task.Task
	for i := range svc.Modules {
		mod := &svc.Modules[i]
		if mod.Python != nil {
			tasks = append(tasks, task.NewDjangoMigrate(svc, mod.Python))
		}
	}
	return NewBaseStrategy("django migrate", tasks)
}

func NewDjangoMakeMigrationsStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Python != nil {
				tasks = append(tasks, task.NewDjangoMakeMigrations(svc, mod.Python))
			}
		}
	}
	return NewBaseStrategy("django makemigrations", tasks)
}

func NewDjangoMakeMigrationsStrategy(svc *config.ServiceConfig) Strategy {
	var tasks []task.Task
	for i := range svc.Modules {
		mod := &svc.Modules[i]
		if mod.Python != nil {
			tasks = append(tasks, task.NewDjangoMakeMigrations(svc, mod.Python))
		}
	}
	return NewBaseStrategy("django makemigrations", tasks)
}

func NewDjangoGenRandomSecretKeyStrategyGlobal(cfg *config.Config) Strategy {
	var tasks []task.Task
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		for j := range svc.Modules {
			mod := &svc.Modules[j]
			if mod.Python != nil {
				tasks = append(tasks, task.NewDjangoGenRandomSecretKey(svc, mod.Python))
			}
		}
	}
	return NewBaseStrategy("django gen-random-secret-key", tasks)
}

func NewDjangoGenRandomSecretKeyStrategy(svc *config.ServiceConfig) Strategy {
	var tasks []task.Task
	for i := range svc.Modules {
		mod := &svc.Modules[i]
		if mod.Python != nil {
			tasks = append(tasks, task.NewDjangoGenRandomSecretKey(svc, mod.Python))
		}
	}
	return NewBaseStrategy("django gen-random-secret-key", tasks)
}
