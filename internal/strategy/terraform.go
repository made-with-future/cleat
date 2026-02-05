package strategy

import (
	"fmt"

	"github.com/madewithfuture/cleat/internal/task"
)

func NewTerraformStrategy(env string, action string, args []string) Strategy {
	name := "terraform:" + action
	if env != "" {
		name = fmt.Sprintf("terraform:%s:%s", action, env)
	}
	return NewBaseStrategy(name, []task.Task{
		task.NewTerraform(env, action, args),
	})
}