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
func buildCommandTree(cfg *config.Config, workflows []history.Workflow) []CommandItem {
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

	tree = append(tree, CommandItem{Label: "build", Command: "build"})
	tree = append(tree, CommandItem{Label: "run", Command: "run"})

	if cfg.Docker {
		tree = append(tree, CommandItem{
			Label: "docker",
			Children: []CommandItem{
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
		if cfg.Terraform.UseFolders && len(cfg.Terraform.Envs) > 0 {
			var tfChildren []CommandItem
			for _, env := range cfg.Terraform.Envs {
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
		}

		if svc.AppYaml != "" {
			svcItem.Children = append(svcItem.Children, CommandItem{Label: "deploy", Command: fmt.Sprintf("gcp app-engine deploy:%s", svc.Name)})
			svcItem.Children = append(svcItem.Children, CommandItem{Label: "promote", Command: fmt.Sprintf("gcp app-engine promote:%s", svc.Name)})
		}

		if len(svcItem.Children) > 0 {
			tree = append(tree, svcItem)
		}
	}

	return tree
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
