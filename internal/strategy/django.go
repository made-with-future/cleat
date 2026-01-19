package strategy

import "github.com/madewithfuture/cleat/internal/task"

func init() {
	Register("django create-user-dev", NewDjangoCreateUserDevStrategy)
	Register("django collectstatic", NewDjangoCollectStaticStrategy)
	Register("django migrate", NewDjangoMigrateStrategy)
}

func NewDjangoCreateUserDevStrategy() Strategy {
	return NewBaseStrategy("django create-user-dev", []task.Task{
		task.NewDjangoCreateUserDev(),
	})
}

func NewDjangoCollectStaticStrategy() Strategy {
	return NewBaseStrategy("django collectstatic", []task.Task{
		task.NewDjangoCollectStatic(),
	})
}

func NewDjangoMigrateStrategy() Strategy {
	return NewBaseStrategy("django migrate", []task.Task{
		task.NewDjangoMigrate(),
	})
}
