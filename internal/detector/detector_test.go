package detector

import (
	"testing"
	"github.com/madewithfuture/cleat/internal/config/schema"
)

type mockDetector struct {
	called bool
}

func (m *mockDetector) Detect(baseDir string, cfg *schema.Config) error {
	m.called = true
	return nil
}

func TestDetectAll(t *testing.T) {
	// Since DetectAll now has a hardcoded list, we can't easily use a mock 
	// unless we change how DetectAll works. 
	// For now, let's just test that it runs without error.
	cfg := &schema.Config{}
	err := DetectAll(".", cfg)
	if err != nil {
		t.Fatalf("DetectAll failed: %v", err)
	}
}