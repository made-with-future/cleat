package strategy

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/task"
)

type TerraformStrategy struct {
	BaseStrategy
}

func NewTerraformStrategy(env string, action string, args []string) *TerraformStrategy {
	name := fmt.Sprintf("terraform:%s", action)
	if env != "" {
		name = fmt.Sprintf("%s:%s", name, env)
	}

	return &TerraformStrategy{
		BaseStrategy: *NewBaseStrategy(name, []task.Task{
			task.NewTerraformTask(env, action, args),
		}),
	}
}

func (s *TerraformStrategy) ResolveTasks(cfg *config.Config) ([]task.Task, error) {
	return s.buildExecutionPlan(cfg)
}
