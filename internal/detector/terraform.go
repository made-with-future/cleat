package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
)

func init() {
	Register(&TerraformDetector{})
}

type TerraformDetector struct{}

func (d *TerraformDetector) Detect(baseDir string, cfg *config.Config) error {
	iacDirName := ".iac"
	if cfg.Terraform != nil && cfg.Terraform.Dir != "" {
		iacDirName = cfg.Terraform.Dir
	}
	iacDir := filepath.Join(baseDir, iacDirName)
	info, err := os.Stat(iacDir)
	if err != nil || !info.IsDir() {
		return nil
	}

	if cfg.Terraform == nil {
		cfg.Terraform = &config.TerraformConfig{}
	}

	// Check for subdirectories (multiple envs) or .tf files (single env)
	entries, err := os.ReadDir(iacDir)
	if err == nil {
		useFolders := false
		detectedEnvs := []string{}
		hasTfFiles := false

		for _, entry := range entries {
			if entry.IsDir() {
				// Check if subdirectory contains .tf files
				subDir := filepath.Join(iacDir, entry.Name())
				subEntries, _ := os.ReadDir(subDir)
				hasTfInSubDir := false
				for _, subEntry := range subEntries {
					if !subEntry.IsDir() && strings.HasSuffix(subEntry.Name(), ".tf") {
						hasTfInSubDir = true
						break
					}
				}

				if hasTfInSubDir {
					useFolders = true
					detectedEnvs = append(detectedEnvs, entry.Name())
				}
			} else if strings.HasSuffix(entry.Name(), ".tf") {
				hasTfFiles = true
			}
		}

		if useFolders {
			cfg.Terraform.UseFolders = true
			if cfg.Terraform.Envs == nil {
				cfg.Terraform.Envs = detectedEnvs
			}
			// Ensure detected terraform envs are also in global envs
			for _, env := range cfg.Terraform.Envs {
				found := false
				for _, existing := range cfg.Envs {
					if existing == env {
						found = true
						break
					}
				}
				if !found {
					cfg.Envs = append(cfg.Envs, env)
				}
			}
		} else if hasTfFiles {
			cfg.Terraform.UseFolders = false
		}
	}

	return nil
}
