package cmd

import "testing"

func TestVersionVariable(t *testing.T) {
	if Version == "" {
		t.Error("Version variable should not be empty")
	}
}
