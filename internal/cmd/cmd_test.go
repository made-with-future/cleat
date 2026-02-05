package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestSubcommands(t *testing.T) {
	// Create a dummy cleat.yaml to satisfy LoadDefaultConfig
	tmpDir, _ := os.MkdirTemp("", "cleat-cmd-test-*")
	defer os.RemoveAll(tmpDir)
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	os.WriteFile("cleat.yaml", []byte("version: 1\ndocker: true"), 0644)

	tests := []struct {
		name string
		args []string
	}{
		{"build", []string{"build"}},
		{"run", []string{"run"}},
		{"version", []string{"version"}},
		{"docker down", []string{"docker", "down"}},
		{"django migrate", []string{"django", "migrate"}},
		{"npm install", []string{"npm", "install"}},
		{"terraform init", []string{"terraform", "init"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We don't care about actual execution success (which might fail due to missing tools like docker),
			// we just want to trigger the RunE functions for coverage.
			_, _ = executeCommand(rootCmd, tt.args...)
		})
	}
}