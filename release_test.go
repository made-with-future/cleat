package cleat_test

import (
	"os"
	"strings"
	"testing"
)

func TestReleaseWorkflowExists(t *testing.T) {
	workflowPath := ".github/workflows/release.yml"
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Release workflow file not found at %s: %v", workflowPath, err)
	}

	workflowStr := string(content)

	// Verify key components
	checks := []string{
		"push:",
		"tags:",
		"- 'v*.*.*'",
		"contents: write",
		"runs-on: ubuntu-latest",
		"Extract Version",
		"GITHUB_REF_NAME",
		"strategy:",
		"matrix:",
		"os: [linux, darwin]",
		"arch: [amd64, arm64]",
		"exclude:",
		"go build",
		"-o cleat",
		"tar -czf",
		"cleat_${VERSION}_${GOOS}_${GOARCH}.tar.gz",
		"actions/upload-artifact",
		"needs: build",
		"actions/download-artifact",
		"gh release create",
		"gh release upload",
	}

	for _, check := range checks {
		if !strings.Contains(workflowStr, check) {
			t.Errorf("Workflow file missing expected content: %q", check)
		}
	}
}

func TestInstallScriptExists(t *testing.T) {
	scriptPath := "install.sh"
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Install script not found at %s: %v", scriptPath, err)
	}

	scriptStr := string(content)

	checks := []string{
		"uname -s", // OS detection
		"uname -m", // Arch detection
		"/usr/local/bin", // Darwin target
		".local/bin", // Linux target
		"curl", // Downloader
		"tar -xzf", // Decompression
		"cleat", // Binary name
	}

	for _, check := range checks {
		if !strings.Contains(scriptStr, check) {
			t.Errorf("Install script missing expected content: %q", check)
		}
	}
}

func TestReadmeInstallation(t *testing.T) {
	readmePath := "README.md"
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("README.md not found at %s: %v", readmePath, err)
	}

	readmeStr := string(content)

	checks := []string{
		"## Installation",
		"curl -fsSL https://raw.githubusercontent.com/made-with-future/cleat/main/install.sh | sh",
		"curl -fsSL https://raw.githubusercontent.com/made-with-future/cleat/main/install.sh | sh -s -- v1.0.0",
	}

	for _, check := range checks {
		if !strings.Contains(readmeStr, check) {
			t.Errorf("README.md missing expected content: %q", check)
		}
	}
}
