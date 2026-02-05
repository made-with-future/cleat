package detector

import (
	"github.com/madewithfuture/cleat/internal/config"
)

// Detector is the interface for project auto-discovery
type Detector interface {
	Detect(baseDir string, cfg *config.Config) error
}

var registry []Detector

// Register adds a detector to the global registry
func Register(d Detector) {
	registry = append(registry, d)
}

// DetectAll runs all registered detectors against the provided config
func DetectAll(baseDir string, cfg *config.Config) error {
	for _, d := range registry {
		if err := d.Detect(baseDir, cfg); err != nil {
			return err
		}
	}
	return nil
}
