package strategy

import "github.com/madewithfuture/cleat/internal/task"

func init() {
	Register("run", NewRunStrategy)
}

// NewRunStrategy creates the run command strategy
func NewRunStrategy() Strategy {
	return NewBaseStrategy("run", []task.Task{
		task.NewDockerUp(),
		task.NewDjangoRunServer(),
		task.NewNpmStart(),
	})
}

// NewNpmScriptStrategy creates a strategy for running a single npm script
func NewNpmScriptStrategy(script string) Strategy {
	return NewBaseStrategy("npm:"+script, []task.Task{
		task.NewNpmRun(script),
	})
}
