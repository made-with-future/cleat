package history

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madewithfuture/cleat/internal/config"
)

type Workflow struct {
	Name     string         `json:"name"`
	Commands []HistoryEntry `json:"commands"`
}

func getWorkflowFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	root := config.FindProjectRoot()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(absRoot))
	projectDirName := filepath.Base(absRoot)
	if projectDirName == "/" || projectDirName == "." || projectDirName == "" {
		projectDirName = "root"
	}

	id := fmt.Sprintf("%s-%x", projectDirName, hash[:8])

	return filepath.Join(home, ".cleat", id+".workflows.json"), nil
}

func SaveWorkflow(workflow Workflow) error {
	workflows, _ := LoadWorkflows()

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

	data, err := json.MarshalIndent(workflows, "", "  ")
	if err != nil {
		return err
	}

	workflowFile, err := getWorkflowFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(workflowFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(workflowFile, data, 0644)
}

func LoadWorkflows() ([]Workflow, error) {
	workflowFile, err := getWorkflowFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(workflowFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Workflow{}, nil
		}
		return nil, err
	}

	var workflows []Workflow
	if err := json.Unmarshal(data, &workflows); err != nil {
		return nil, err
	}

	return workflows, nil
}

func DeleteWorkflow(name string) error {
	workflows, err := LoadWorkflows()
	if err != nil {
		return err
	}

	newWorkflows := []Workflow{}
	for _, w := range workflows {
		if w.Name != name {
			newWorkflows = append(newWorkflows, w)
		}
	}

	data, err := json.MarshalIndent(newWorkflows, "", "  ")
	if err != nil {
		return err
	}

	workflowFile, err := getWorkflowFilePath()
	if err != nil {
		return err
	}

	return os.WriteFile(workflowFile, data, 0644)
}
