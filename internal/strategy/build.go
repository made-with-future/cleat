package strategy

import "github.com/madewithfuture/cleat/internal/task"

func init() {
	Register("build", NewBuildStrategy)
}

// NewBuildStrategy creates the build command strategy
func NewBuildStrategy() Strategy {
	return NewBaseStrategy("build", []task.Task{
		task.NewDockerBuild(),
		task.NewNpmBuild(),
		task.NewDjangoCollectStatic(),
	})
}
