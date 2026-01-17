package strategy

import "github.com/madewithfuture/cleat/internal/config"

type Runner func(name string, args ...string) error

type Strategy interface {
	Run(cfg *config.Config, runner Runner) error
}
