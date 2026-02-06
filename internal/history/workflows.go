package history

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/logger"
	"gopkg.in/yaml.v3"
)

func Slugify(name string) string {
	name = strings.ToLower(name)
	var result strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		} else if r == ' ' {
			result.WriteRune('-')
		}
	}
	s := result.String()
	// Collapse multiple hyphens
	s = regexp.MustCompile("-+").ReplaceAllString(s, "-")
	// Trim hyphens
	s = strings.Trim(s, "-")
	return s
}

func ValidateWorkflowName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}
	return nil
}

func GetUserWorkflowFilePath() (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	id := config.GetProjectID()

	return filepath.Join(home, ".cleat", id+".workflows.yaml"), nil
}

func modifyWorkflowFile(path string, op func([]config.Workflow) ([]config.Workflow, bool, error)) error {
	var workflows []config.Workflow
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &workflows); err != nil {
			return fmt.Errorf("failed to parse workflows in %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read workflows in %s: %w", path, err)
	}

	newWorkflows, modified, err := op(workflows)
	if err != nil {
		return err
	}
	if !modified {
		return nil
	}

	data, err := yaml.Marshal(newWorkflows)
	if err != nil {
		return fmt.Errorf("failed to marshal workflows: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflows to %s: %w", path, err)
	}
	return nil
}

func SaveWorkflowToProject(workflow config.Workflow) error {
	if workflow.ID == "" {
		workflow.ID = Slugify(workflow.Name)
	}

	// Save to project-local file. Prefer existing .yaml if it exists.
	root := config.FindProjectRoot()
	projectFile := filepath.Join(root, "cleat.workflows.yaml")
	// If .yaml doesn't exist but .yml DOES, use .yml as fallback
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(root, "cleat.workflows.yml")); err == nil {
			projectFile = filepath.Join(root, "cleat.workflows.yml")
		}
	}

	return modifyWorkflowFile(projectFile, func(workflows []config.Workflow) ([]config.Workflow, bool, error) {
		// Update existing or add new
		found := false
		for i, w := range workflows {
			id := w.ID
			if id == "" {
				id = Slugify(w.Name)
			}
			if id == workflow.ID {
				workflows[i] = workflow
				found = true
				break
			}
		}
		if !found {
			workflows = append(workflows, workflow)
		}
		return workflows, true, nil
	})
}

func SaveWorkflowToUser(workflow config.Workflow) error {
	if workflow.ID == "" {
		workflow.ID = Slugify(workflow.Name)
	}

	userFile, err := GetUserWorkflowFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(userFile), 0755); err != nil {
		return fmt.Errorf("failed to create user config directory: %w", err)
	}

	return modifyWorkflowFile(userFile, func(workflows []config.Workflow) ([]config.Workflow, bool, error) {
		// Update existing or add new
		found := false
		for i, w := range workflows {
			id := w.ID
			if id == "" {
				id = Slugify(w.Name)
			}
			if id == workflow.ID {
				workflows[i] = workflow
				found = true
				break
			}
		}
		if !found {
			workflows = append(workflows, workflow)
		}
		return workflows, true, nil
	})
}

func LoadWorkflows(cfg *config.Config) ([]config.Workflow, error) {
	workflowsMap := make(map[string]config.Workflow)

	// 1. Load from cleat.yaml/cleat.yml (if available in cfg)
	if cfg != nil {
		for _, w := range cfg.Workflows {
			if w.ID == "" {
				w.ID = Slugify(w.Name)
			}
			workflowsMap[w.ID] = w
		}
	}

	// 2. Load from cleat.workflows.yaml or .yml in project root
	root := config.FindProjectRoot()
	projectFiles := []string{
		filepath.Join(root, "cleat.workflows.yml"),
		filepath.Join(root, "cleat.workflows.yaml"),
	}
	for _, projectFile := range projectFiles {
		if data, err := os.ReadFile(projectFile); err == nil {
			var projectWorkflows []config.Workflow
			if err := yaml.Unmarshal(data, &projectWorkflows); err == nil {
				for _, w := range projectWorkflows {
					if w.ID == "" {
						w.ID = Slugify(w.Name)
					}
					workflowsMap[w.ID] = w
				}
			} else {
				return nil, fmt.Errorf("failed to parse project workflows in %s: %w", projectFile, err)
			}
		}
	}

	// 3. Load from user per-project file in home dir (overrides project-local)
	if userFile, err := GetUserWorkflowFilePath(); err == nil {
		if data, err := os.ReadFile(userFile); err == nil {
			var userWorkflows []config.Workflow
			if err := yaml.Unmarshal(data, &userWorkflows); err == nil {
				for _, w := range userWorkflows {
					if w.ID == "" {
						w.ID = Slugify(w.Name)
					}
					workflowsMap[w.ID] = w
				}
			} else {
				return nil, fmt.Errorf("failed to parse user workflows in %s: %w", userFile, err)
			}
		}
	}

	// Convert map back to slice and validate
	res := make([]config.Workflow, 0, len(workflowsMap))
	for _, w := range workflowsMap {
		if w.Name == "" {
			logger.Warn("skipping workflow with empty name", nil)
			continue
		}
		if len(w.Commands) == 0 {
			logger.Warn("skipping workflow with no commands", map[string]interface{}{"workflow": w.Name})
			continue
		}
		res = append(res, w)
	}

	// Sort by name for consistent UI
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res, nil
}

func DeleteWorkflow(idOrName string) error {
	root := config.FindProjectRoot()
	projectFiles := []string{
		filepath.Join(root, "cleat.workflows.yaml"),
		filepath.Join(root, "cleat.workflows.yml"),
	}

	op := func(workflows []config.Workflow) ([]config.Workflow, bool, error) {
		newWorkflows := []config.Workflow{}
		modified := false
		for _, w := range workflows {
			id := w.ID
			if id == "" {
				id = Slugify(w.Name)
			}
			if id != idOrName && w.Name != idOrName {
				newWorkflows = append(newWorkflows, w)
			} else {
				modified = true
			}
		}
		return newWorkflows, modified, nil
	}

	for _, projectFile := range projectFiles {
		if err := modifyWorkflowFile(projectFile, op); err != nil {
			return err
		}
	}

	// Also check user file
	userFile, err := GetUserWorkflowFilePath()
	if err == nil {
		if err := modifyWorkflowFile(userFile, op); err != nil {
			return err
		}
	}

	return nil
}
