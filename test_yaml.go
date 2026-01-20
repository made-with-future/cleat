package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type PythonConfig struct {
	Django bool `yaml:"django"`
}

type NpmConfig struct {
	Scripts []string `yaml:"scripts"`
}

type ModuleConfig struct {
	Python *PythonConfig `yaml:"python.django,omitempty"`
	Npm    *NpmConfig    `yaml:"npm,omitempty"`
}

type ServiceConfig struct {
	Name     string         `yaml:"name"`
	Location string         `yaml:"location"`
	Modules  []ModuleConfig `yaml:"modules"`
}

type Config struct {
	Services []ServiceConfig `yaml:"services"`
}

func main() {
	data := `
services:
  - name: backend
    location: ./backend
    modules:
      - python.django:
          django: true
      - npm:
          scripts: [build]
`
	var cfg Config
	err := yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Config: %+v\n", cfg)
	for _, svc := range cfg.Services {
		fmt.Printf("Service: %s\n", svc.Name)
		for _, mod := range svc.Modules {
			if mod.Python != nil {
				fmt.Printf("  Module: Python (Django=%v)\n", mod.Python.Django)
			}
			if mod.Npm != nil {
				fmt.Printf("  Module: Npm (Scripts=%v)\n", mod.Npm.Scripts)
			}
		}
	}
}
