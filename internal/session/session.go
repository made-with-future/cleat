package session

import (
	"github.com/madewithfuture/cleat/internal/config/schema"
	"github.com/madewithfuture/cleat/internal/executor"
)

// Session encapsulates the runtime state of a Cleat execution
type Session struct {
	Config *schema.Config
	Inputs map[string]string
	Exec   executor.Executor
}

// NewSession creates a new session with the provided configuration and executor
func NewSession(cfg *schema.Config, exec executor.Executor) *Session {
	return &Session{
		Config: cfg,
		Inputs: make(map[string]string),
		Exec:   exec,
	}
}
