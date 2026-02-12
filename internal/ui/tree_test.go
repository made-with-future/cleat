package ui

import (
	"testing"

	"github.com/madewithfuture/cleat/internal/config"
)

func TestBuildCommandTree_FlattenDefaultService(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "default",
				Modules: []config.ModuleConfig{
					{Go: &config.GoConfig{}},
				},
			},
		},
	}

	tree := buildCommandTree(cfg, nil)

	// We expect "go" to be at the root, NOT inside "default"
	foundGoAtRoot := false
	foundDefaultService := false

	for _, item := range tree {
		if item.Label == "default" {
			foundDefaultService = true
		}
		if item.Label == "go" {
			foundGoAtRoot = true
		}
	}

	if foundDefaultService {
		t.Error("Should not have 'default' service node when it's the only service")
	}
	if !foundGoAtRoot {
		t.Error("Should have 'go' node at the root of the tree")
	}

	// Verify no duplicates
	goCount := 0
	for _, item := range tree {
		if item.Label == "go" {
			goCount++
		}
	}
	if goCount != 1 {
		t.Errorf("Expected exactly 1 'go' node at root, found %d", goCount)
	}
}

func TestBuildCommandTree_MultipleServices_NoFlattening(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "default",
				Modules: []config.ModuleConfig{
					{Go: &config.GoConfig{}},
				},
			},
			{
				Name: "other",
				Modules: []config.ModuleConfig{
					{Npm: &config.NpmConfig{}},
				},
			},
		},
	}

	tree := buildCommandTree(cfg, nil)

	foundDefaultService := false
	foundOtherService := false

	for _, item := range tree {
		if item.Label == "default" {
			foundDefaultService = true
		}
		if item.Label == "other" {
			foundOtherService = true
		}
	}

	if !foundDefaultService {
		t.Error("Should have 'default' service node when multiple services exist")
	}
	if !foundOtherService {
		t.Error("Should have 'other' service node when multiple services exist")
	}
}
