package detector

import (
	"os"
	"path/filepath"
	"strings"

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
		Build   interface{} `yaml:"build"`
		Image   string      `yaml:"image"`
		Command interface{} `yaml:"command"`
	}
	var dc struct {
		Services map[string]dcService `yaml:"services"`
	}
	if err := yaml.Unmarshal(dcData, &dc); err != nil {
		return err
	}

	for name, s := range dc.Services {
		buildContext := ""
		dockerfile := ""
		if s.Build != nil {
			if b, ok := s.Build.(string); ok {
				buildContext = b
			} else if b, ok := s.Build.(map[string]interface{}); ok {
				if context, ok := b["context"].(string); ok {
					buildContext = context
				}
				if df, ok := b["dockerfile"].(string); ok {
					dockerfile = df
				}
			}
		}

		command := ""
		if s.Command != nil {
			if c, ok := s.Command.(string); ok {
				command = c
			} else if c, ok := s.Command.([]interface{}); ok {
				var parts []string
				for _, p := range c {
					if ps, ok := p.(string); ok {
						parts = append(parts, ps)
					}
				}
				command = strings.Join(parts, " ")
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
				if cfg.Services[i].Dockerfile == "" && dockerfile != "" {
					cfg.Services[i].Dockerfile = dockerfile
				}
				if cfg.Services[i].Image == "" && s.Image != "" {
					cfg.Services[i].Image = s.Image
				}
				if cfg.Services[i].Command == "" && command != "" {
					cfg.Services[i].Command = command
				}
				found = true
				break
			}
		}
		if !found {
			cfg.Services = append(cfg.Services, schema.ServiceConfig{
				Name:       name,
				Docker:     ptrBool(true),
				Dir:        buildContext,
				Dockerfile: dockerfile,
				Image:      s.Image,
				Command:    command,
			})
		}
	}

	return nil
}

func ptrBool(b bool) *bool {
	return &b
}
