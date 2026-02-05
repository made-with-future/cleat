package strategy

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestRegistryProvider(t *testing.T) {
	// Setup: Register a test strategy
	strategyName := "registry-test-cmd"
	Register(strategyName, func(cfg *config.Config) Strategy {
		return NewBaseStrategy(strategyName, nil)
	})

	provider := &RegistryProvider{}

	// Test CanHandle
	if !provider.CanHandle(strategyName) {
		t.Errorf("expected CanHandle(%q) to be true", strategyName)
	}
	if provider.CanHandle("unknown-cmd") {
		t.Error("expected CanHandle(\"unknown-cmd\") to be false")
	}

	// Test GetStrategy
	s := provider.GetStrategy(strategyName, nil)
	if s == nil {
		t.Fatal("expected GetStrategy to return a strategy")
	}
	if s.Name() != strategyName {
		t.Errorf("expected strategy name %q, got %q", strategyName, s.Name())
	}

	s = provider.GetStrategy("unknown-cmd", nil)
	if s != nil {
		t.Errorf("expected nil strategy for unknown command, got %v", s)
	}
}
