package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/madewithfuture/cleat/internal/config/schema"
)

type TerraformDetector struct{}

func (d *TerraformDetector) Detect(baseDir string, cfg *schema.Config) error {
	iacDirName := ".iac"
	if cfg.Terraform != nil && cfg.Terraform.Dir != "" {
		iacDirName = cfg.Terraform.Dir
	}
	iacDir := filepath.Join(baseDir, iacDirName)
	info, err := os.Stat(iacDir)

	// If not found in baseDir and it was the default ".iac", try searching deeper
	if (err != nil || !info.IsDir()) && iacDirName == ".iac" {
		foundDir := ""
		filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				return nil
			}
			// Skip common large/irrelevant directories
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "venv" || name == ".envs" || name == "testdata" || name == "vendor" {
				return filepath.SkipDir
			}
			if name == ".iac" {
				rel, err := filepath.Rel(baseDir, path)
				if err == nil {
					foundDir = rel
					return filepath.SkipAll // Stop searching
				}
			}
			return nil
		})
		if foundDir != "" {
			iacDir = filepath.Join(baseDir, foundDir)
			iacDirName = foundDir
			info, err = os.Stat(iacDir)
		}
	}

	if err != nil || !info.IsDir() {
		return nil
	}

	if cfg.Terraform == nil {
		cfg.Terraform = &schema.TerraformConfig{}
	}
	if cfg.Terraform.Dir == "" && iacDirName != ".iac" {
		cfg.Terraform.Dir = iacDirName
	}

	entries, err := os.ReadDir(iacDir)
	if err == nil {
		useFolders := false
		detectedEnvs := []string{}
		hasTfFiles := false

		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(iacDir, entry.Name())
				hasTfInSubDir := false
				filepath.Walk(subDir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
						hasTfInSubDir = true
						return filepath.SkipAll
					}
					return nil
				})

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
