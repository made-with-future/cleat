package history

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/madewithfuture/cleat/internal/config"
	"github.com/madewithfuture/cleat/internal/logger"
	"gopkg.in/yaml.v3"
)

func GetUserWorkflowFilePath() (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	root := config.FindProjectRoot()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute project root: %w", err)
	}

	hash := sha256.Sum256([]byte(absRoot))
	projectDirName := filepath.Base(absRoot)
	if projectDirName == "/" || projectDirName == "." || projectDirName == "" {
		projectDirName = "root"
	}

	id := fmt.Sprintf("%s-%x", projectDirName, hash[:8])

	return filepath.Join(home, ".cleat", id+".workflows.yaml"), nil
}

func SaveWorkflowToProject(workflow config.Workflow) error {
	// Save to project-local file. Prefer existing .yaml if it exists.
	root := config.FindProjectRoot()
	projectFile := filepath.Join(root, "cleat.workflows.yaml")
	// If .yaml doesn't exist but .yml DOES, use .yml as fallback
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(root, "cleat.workflows.yml")); err == nil {
			projectFile = filepath.Join(root, "cleat.workflows.yml")
		}
	}

	var workflows []config.Workflow
	if data, err := os.ReadFile(projectFile); err == nil {
		if err := yaml.Unmarshal(data, &workflows); err != nil {
			logger.Warn("failed to unmarshal project workflows before saving", map[string]interface{}{"path": projectFile, "error": err.Error()})
		}
	}

	// Update existing or add new
	found := false
	for i, w := range workflows {
		if w.Name == workflow.Name {
			workflows[i] = workflow
			found = true
			break
		}
	}
	if !found {
		workflows = append(workflows, workflow)
	}

	data, err := yaml.Marshal(workflows)
	if err != nil {
		return fmt.Errorf("failed to marshal workflows: %w", err)
	}

	if err := os.WriteFile(projectFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write project workflows: %w", err)
	}
	return nil
}

func SaveWorkflowToUser(workflow config.Workflow) error {
	userFile, err := GetUserWorkflowFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(userFile), 0755); err != nil {
		return fmt.Errorf("failed to create user config directory: %w", err)
	}

	var workflows []config.Workflow
	if data, err := os.ReadFile(userFile); err == nil {
		if err := yaml.Unmarshal(data, &workflows); err != nil {
			logger.Warn("failed to unmarshal user workflows before saving", map[string]interface{}{"path": userFile, "error": err.Error()})
		}
	}

	// Update existing or add new
	found := false
	for i, w := range workflows {
		if w.Name == workflow.Name {
			workflows[i] = workflow
			found = true
			break
		}
	}
	if !found {
		workflows = append(workflows, workflow)
	}

	data, err := yaml.Marshal(workflows)
	if err != nil {
		return fmt.Errorf("failed to marshal workflows: %w", err)
	}

	if err := os.WriteFile(userFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write user workflows: %w", err)
	}
	return nil
}

func LoadWorkflows(cfg *config.Config) ([]config.Workflow, error) {
	workflowsMap := make(map[string]config.Workflow)

	// 1. Load from cleat.yaml/cleat.yml (if available in cfg)
	if cfg != nil {
		for _, w := range cfg.Workflows {
			workflowsMap[w.Name] = w
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
					workflowsMap[w.Name] = w
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
					workflowsMap[w.Name] = w
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

func DeleteWorkflow(name string) error {
	root := config.FindProjectRoot()
	projectFiles := []string{
		filepath.Join(root, "cleat.workflows.yaml"),
		filepath.Join(root, "cleat.workflows.yml"),
	}

	for _, projectFile := range projectFiles {
		data, err := os.ReadFile(projectFile)
		if err != nil {
			continue
		}

		var workflows []config.Workflow
		if err := yaml.Unmarshal(data, &workflows); err != nil {
			logger.Warn("failed to unmarshal workflows during delete", map[string]interface{}{"path": projectFile, "error": err.Error()})
			continue
		}

		newWorkflows := []config.Workflow{}
		for _, w := range workflows {
			if w.Name != name {
				newWorkflows = append(newWorkflows, w)
			}
		}

		if len(newWorkflows) == len(workflows) {
			continue
		}

		newData, err := yaml.Marshal(newWorkflows)
		if err != nil {
			return fmt.Errorf("failed to marshal workflows after delete: %w", err)
		}

		if err := os.WriteFile(projectFile, newData, 0644); err != nil {
			return fmt.Errorf("failed to write project workflows after delete: %w", err)
		}
	}

	// Also check user file
	userFile, err := GetUserWorkflowFilePath()
	if err == nil {
		if data, err := os.ReadFile(userFile); err == nil {
			var workflows []config.Workflow
			if err := yaml.Unmarshal(data, &workflows); err == nil {
				newWorkflows := []config.Workflow{}
				for _, w := range workflows {
					if w.Name != name {
						newWorkflows = append(newWorkflows, w)
					}
				}
				if len(newWorkflows) != len(workflows) {
					newData, err := yaml.Marshal(newWorkflows)
					if err == nil {
						os.WriteFile(userFile, newData, 0644)
					}
				}
			}
		}
	}

	return nil
}