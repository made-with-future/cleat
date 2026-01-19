package strategy

import "github.com/madewithfuture/cleat/internal/task"

func init() {
	Register("django create-user-dev", NewDjangoCreateUserDevStrategy)
	Register("django collectstatic", NewDjangoCollectStaticStrategy)
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
