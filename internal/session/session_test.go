package session

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config/schema"
	"github.com/madewithfuture/cleat/internal/executor"
)

type mockExecutor struct {
	executor.Executor
}

func TestNewSession(t *testing.T) {
	cfg := &schema.Config{Version: 1}
	exec := &mockExecutor{}
	
	sess := NewSession(cfg, exec)
	
	if sess.Config != cfg {
		t.Errorf("expected config %v, got %v", cfg, sess.Config)
	}
	if sess.Exec != exec {
		t.Errorf("expected executor %v, got %v", exec, sess.Exec)
	}
	if sess.Inputs == nil {
		t.Error("expected inputs map to be initialized")
	}
}
