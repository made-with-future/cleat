package detector

import (
	"github.com/madewithfuture/cleat/internal/config/schema"
)

// Detector is the interface for project auto-discovery
type Detector interface {
	Detect(baseDir string, cfg *schema.Config) error
}

// DetectAll runs all registered detectors against the provided config in a specific order
func DetectAll(baseDir string, cfg *schema.Config) error {
	detectors := []Detector{
		&EnvDetector{},
		&DockerDetector{},
		&DjangoDetector{},
		&RubyDetector{},
		&NpmDetector{},
		&GoDetector{},
		&GcpDetector{},
		&TerraformDetector{},
	}

	for _, d := range detectors {
		if err := d.Detect(baseDir, cfg); err != nil {
			return err
		}
	}
	return nil
}
