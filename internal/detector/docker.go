package detector

import (
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config/schema"
	"gopkg.in/yaml.v3"
)

type DockerDetector struct{}

func (d *DockerDetector) Detect(baseDir string, cfg *schema.Config) error {
	dockerComposeFile := ""
	if _, err := os.Stat(filepath.Join(baseDir, "docker-compose.yaml")); err == nil {
		dockerComposeFile = "docker-compose.yaml"
	} else if _, err := os.Stat(filepath.Join(baseDir, "docker-compose.yml")); err == nil {
		dockerComposeFile = "docker-compose.yml"
	}

	if dockerComposeFile == "" {
		return nil
	}

	cfg.Docker = true
	dcPath := filepath.Join(baseDir, dockerComposeFile)
	dcData, err := os.ReadFile(dcPath)
	if err != nil {
		return nil
	}

	type dcService struct {
		Build interface{} `yaml:"build"`
	}
	var dc struct {
		Services map[string]dcService `yaml:"services"`
	}
	if err := yaml.Unmarshal(dcData, &dc); err != nil {
		return nil
	}

	for name, s := range dc.Services {
		buildContext := ""
		if s.Build != nil {
			if b, ok := s.Build.(string); ok {
				buildContext = b
			} else if b, ok := s.Build.(map[string]interface{}); ok {
				if context, ok := b["context"].(string); ok {
					buildContext = context
				}
			}
		}

		found := false
		for i := range cfg.Services {
			if cfg.Services[i].Name == name {
				if cfg.Services[i].Docker == nil {
					cfg.Services[i].Docker = ptrBool(true)
				}
				if cfg.Services[i].Dir == "" && buildContext != "" {
					cfg.Services[i].Dir = buildContext
				}
				found = true
				break
			}
		}
		if !found {
			cfg.Services = append(cfg.Services, schema.ServiceConfig{
				Name:   name,
				Docker: ptrBool(true),
				Dir:    buildContext,
			})
		}
	}

	return nil
}

func ptrBool(b bool) *bool {
	return &b
}