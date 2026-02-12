package ui

import (
	"fmt"
	"strings"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/history"
)

const defaultConfigTemplate = `# Cleat configuration
# See https://github.com/madewithfuture/cleat for documentation

version: 1
docker: true
services:
  - name: backend
    dir: .
    modules:
      - python:
          django: true
          django_service: backend
  - name: frontend
    dir: ./frontend
    modules:
      - npm:
          service: backend-node
          scripts:
            - build
`

// buildCommandTree creates the commands tree from config and workflows
func buildCommandTree(cfg *config.Config, workflows []config.Workflow) []CommandItem {
	var tree []CommandItem

	if len(workflows) > 0 {
		var workflowChildren []CommandItem
		for _, w := range workflows {
			workflowChildren = append(workflowChildren, CommandItem{
				Label:   w.Name,
				Command: "workflow:" + w.Name,
			})
		}
		tree = append(tree, CommandItem{
			Label:    "workflows",
			Children: workflowChildren,
		})
	}

	isFlattened := len(cfg.Services) == 1 && (cfg.Services[0].Name == "default" || cfg.Services[0].Name == "")

	hasDocker := cfg.Docker
	if !hasDocker {
		for i := range cfg.Services {
			if cfg.Services[i].IsDocker() {
				hasDocker = true
				break
			}
		}
	}

	if hasDocker && !isFlattened {
		tree = append(tree, CommandItem{
			Label: "docker",
			Children: []CommandItem{
				{Label: "up", Command: "docker up"},
				{Label: "down", Command: "docker down"},
				{Label: "rebuild", Command: "docker rebuild"},
				{Label: "remove-orphans", Command: "docker remove-orphans"},
			},
		})
	}

	if cfg.GoogleCloudPlatform != nil {
		gcpChildren := []CommandItem{
			{Label: "activate", Command: "gcp activate"},
			{Label: "adc-login", Command: "gcp adc-login"},
			{Label: "adc-impersonate-login", Command: "gcp adc-impersonate-login"},
			{Label: "init", Command: "gcp init"},
			{Label: "set-config", Command: "gcp set-config"},
		}
		if cfg.AppYaml != "" {
			gcpChildren = append(gcpChildren, CommandItem{Label: "deploy", Command: "gcp app-engine deploy"})
			gcpChildren = append(gcpChildren, CommandItem{Label: "promote", Command: "gcp app-engine promote"})
		}
		gcpChildren = append(gcpChildren, CommandItem{Label: "console", Command: "gcp console"})
		tree = append(tree, CommandItem{
			Label:    "gcp",
			Children: gcpChildren,
		})
	}

	if cfg.Terraform != nil {
		tfEnvs := cfg.Terraform.Envs
		if !cfg.Terraform.UseFolders {
			// Merge with general environments for UI options
			for _, e := range cfg.Envs {
				found := false
				for _, existing := range tfEnvs {
					if existing == e {
						found = true
						break
					}
				}
				if !found {
					tfEnvs = append(tfEnvs, e)
				}
			}
		}

		if len(tfEnvs) > 0 {
			var tfChildren []CommandItem
			for _, env := range tfEnvs {
				tfChildren = append(tfChildren, CommandItem{
					Label: env,
					Children: []CommandItem{
						{Label: "init", Command: "terraform init:" + env},
						{Label: "init-upgrade", Command: "terraform init-upgrade:" + env},
						{Label: "plan", Command: "terraform plan:" + env},
						{Label: "apply", Command: "terraform apply:" + env},
						{Label: "apply-refresh", Command: "terraform apply-refresh:" + env},
					},
				})
			}
			tree = append(tree, CommandItem{
				Label:    "terraform",
				Children: tfChildren,
			})
		} else {
			tree = append(tree, CommandItem{
				Label: "terraform",
				Children: []CommandItem{
					{Label: "init", Command: "terraform init"},
					{Label: "init-upgrade", Command: "terraform init-upgrade"},
					{Label: "plan", Command: "terraform plan"},
					{Label: "apply", Command: "terraform apply"},
					{Label: "apply-refresh", Command: "terraform apply-refresh"},
				},
			})
		}
	}

	foundGo := false
	for i := range cfg.Services {
		for j := range cfg.Services[i].Modules {
			if cfg.Services[i].Modules[j].Go != nil {
				foundGo = true
				break
			}
		}
		if foundGo {
			break
		}
	}

	if foundGo && !isFlattened {
		tree = append(tree, CommandItem{
			Label: "go",
			Children: []CommandItem{
				{Label: "build", Command: "go build"},
				{Label: "test", Command: "go test"},
				{Label: "fmt", Command: "go fmt"},
				{Label: "vet", Command: "go vet"},
				{Label: "mod tidy", Command: "go mod tidy"},
				{Label: "generate", Command: "go generate"},
				{Label: "run", Command: "go run"},
				{Label: "coverage", Command: "go coverage"},
				{Label: "install", Command: "go install"},
			},
		})
	}

	for i := range cfg.Services {
		svc := &cfg.Services[i]
		svcItem := CommandItem{
			Label: svc.Name,
		}

		for j := range svc.Modules {
			mod := &svc.Modules[j]

			// Python/Django
			if mod.Python != nil && mod.Python.Django {
				var djangoChildren []CommandItem
				if cfg.Docker {
					djangoChildren = append(djangoChildren, CommandItem{Label: "create-user-dev", Command: fmt.Sprintf("django create-user-dev:%s", svc.Name)})
				}
				djangoChildren = append(djangoChildren, CommandItem{Label: "collectstatic", Command: fmt.Sprintf("django collectstatic:%s", svc.Name)})
				djangoChildren = append(djangoChildren, CommandItem{Label: "makemigrations", Command: fmt.Sprintf("django makemigrations:%s", svc.Name)})
				djangoChildren = append(djangoChildren, CommandItem{Label: "migrate", Command: fmt.Sprintf("django migrate:%s", svc.Name)})
				djangoChildren = append(djangoChildren, CommandItem{Label: "gen-random-secret-key", Command: fmt.Sprintf("django gen-random-secret-key:%s", svc.Name)})

				svcItem.Children = append(svcItem.Children, CommandItem{
					Label:    "django",
					Children: djangoChildren,
				})
			}

			// NPM
			if mod.Npm != nil {
				npmItem := CommandItem{
					Label: "npm",
				}
				npmItem.Children = append(npmItem.Children, CommandItem{
					Label:   "install",
					Command: fmt.Sprintf("npm install:%s", svc.Name),
				})
				for _, script := range mod.Npm.Scripts {
					npmItem.Children = append(npmItem.Children, CommandItem{
						Label:   fmt.Sprintf("run %s", script),
						Command: fmt.Sprintf("npm run %s:%s", svc.Name, script),
					})
				}
				svcItem.Children = append(svcItem.Children, npmItem)
			}

			// Go
			if mod.Go != nil {
				goItem := CommandItem{
					Label: "go",
				}
				goItem.Children = append(goItem.Children, []CommandItem{
					{Label: "build", Command: fmt.Sprintf("go build:%s", svc.Name)},
					{Label: "test", Command: fmt.Sprintf("go test:%s", svc.Name)},
					{Label: "fmt", Command: fmt.Sprintf("go fmt:%s", svc.Name)},
					{Label: "vet", Command: fmt.Sprintf("go vet:%s", svc.Name)},
					{Label: "mod tidy", Command: fmt.Sprintf("go mod tidy:%s", svc.Name)},
					{Label: "generate", Command: fmt.Sprintf("go generate:%s", svc.Name)},
					{Label: "run", Command: fmt.Sprintf("go run:%s", svc.Name)},
					{Label: "coverage", Command: fmt.Sprintf("go coverage:%s", svc.Name)},
					{Label: "install", Command: fmt.Sprintf("go install:%s", svc.Name)},
				}...)
				svcItem.Children = append(svcItem.Children, goItem)
			}
		}

		if svc.IsDocker() {
			svcItem.Children = append(svcItem.Children, CommandItem{
				Label: "docker",
				Children: []CommandItem{
					{Label: "up", Command: fmt.Sprintf("docker up:%s", svc.Name)},
					{Label: "down", Command: fmt.Sprintf("docker down:%s", svc.Name)},
					{Label: "rebuild", Command: fmt.Sprintf("docker rebuild:%s", svc.Name)},
					{Label: "remove-orphans", Command: fmt.Sprintf("docker remove-orphans:%s", svc.Name)},
				},
			})
		}

		if svc.AppYaml != "" {
			svcItem.Children = append(svcItem.Children, CommandItem{Label: "deploy", Command: fmt.Sprintf("gcp app-engine deploy:%s", svc.Name)})
			svcItem.Children = append(svcItem.Children, CommandItem{Label: "promote", Command: fmt.Sprintf("gcp app-engine promote:%s", svc.Name)})
		}

		if len(svcItem.Children) > 0 {
			if len(cfg.Services) == 1 && (svc.Name == "default" || svc.Name == "") {
				tree = append(tree, svcItem.Children...)
			} else {
				tree = append(tree, svcItem)
			}
		}
	}

	// Add recent commands, filtered by what's actually in the tree
	if recentCmds, err := history.GetTopCommands(3); err == nil && len(recentCmds) > 0 {
		validCommands := make(map[string]bool)
		collectCommands(tree, validCommands)

		var recentChildren []CommandItem
		for _, cmd := range recentCmds {
			if !validCommands[cmd.Command] {
				continue
			}
			label := cmd.Command
			if strings.HasPrefix(label, "workflow:") {
				label = fmt.Sprintf("Workflow: %s", strings.TrimPrefix(label, "workflow:"))
			}
			recentChildren = append(recentChildren, CommandItem{
				Label:   label,
				Command: cmd.Command,
			})
		}
		if len(recentChildren) > 0 {
			recentNode := CommandItem{
				Label:    "recent",
				Children: recentChildren,
				Expanded: true,
			}
			// Prepend recent commands
			tree = append([]CommandItem{recentNode}, tree...)
		}
	}

	return tree
}

func collectCommands(items []CommandItem, commands map[string]bool) {
	for i := range items {
		if items[i].Command != "" {
			commands[items[i].Command] = true
		}
		collectCommands(items[i].Children, commands)
	}
}

// matches checks if an item matches the filter text
func matches(item *CommandItem, text string) bool {
	if text == "" {
		return true
	}
	text = strings.ToLower(text)
	return strings.Contains(strings.ToLower(item.Label), text) ||
		strings.Contains(strings.ToLower(item.Command), text)
}

// anyDescendantMatches checks if any descendant of an item matches the filter
func anyDescendantMatches(item *CommandItem, text string) bool {
	for i := range item.Children {
		if matches(&item.Children[i], text) || anyDescendantMatches(&item.Children[i], text) {
			return true
		}
	}
	return false
}
