package detector

import (
	"testing"
	"github.com/madewithfuture/cleat/internal/config"
)

type mockDetector struct {
	called bool
}

func (m *mockDetector) Detect(baseDir string, cfg *config.Config) error {
	m.called = true
	return nil
}

func TestRegistry(t *testing.T) {
	d := &mockDetector{}
	Register(d)

	cfg := &config.Config{}
	err := DetectAll(".", cfg)
	if err != nil {
		t.Fatalf("DetectAll failed: %v", err)
	}

	if !d.called {
		t.Error("expected detector to be called")
	}
}
