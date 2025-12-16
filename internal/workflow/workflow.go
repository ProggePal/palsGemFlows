package workflow

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Step struct {
	ID           string `yaml:"id"`
	Type         string `yaml:"type"`
	Prompt       string `yaml:"prompt"`
	Multiline    bool   `yaml:"multiline"`
	FromClipboard bool  `yaml:"from_clipboard"`
	UserPrompt   string `yaml:"user_prompt"`
	SystemPrompt string `yaml:"system_prompt"`
	Model        string `yaml:"model"`
	Filename     string `yaml:"filename"`
	Content      string `yaml:"content"`
	ParallelGroup string `yaml:"parallel_group"`
}

type Workflow struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
}

func LoadFromWorkflowsDir(dir string, key string) (Workflow, error) {
	candidates := []string{}
	if filepath.Ext(key) == "" {
		candidates = append(candidates,
			filepath.Join(dir, key+".yaml"),
			filepath.Join(dir, key+".yml"),
		)
	} else {
		candidates = append(candidates, filepath.Join(dir, key))
	}

	var lastErr error
	for _, path := range candidates {
		wf, err := LoadFromFile(path)
		if err == nil {
			return wf, nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = errors.New("no workflow candidates")
	}
	return Workflow{}, lastErr
}

func LoadFromFile(path string) (Workflow, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Workflow{}, fmt.Errorf("read %s: %w", path, err)
	}

	var wf Workflow
	if err := yaml.Unmarshal(b, &wf); err != nil {
		return Workflow{}, fmt.Errorf("parse yaml %s: %w", path, err)
	}
	if err := validateWorkflow(wf); err != nil {
		return Workflow{}, fmt.Errorf("invalid workflow %s: %w", path, err)
	}
	return wf, nil
}

// ListKeys returns workflow names (file name without extension) found in a directory.
func ListKeys(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	seen := map[string]struct{}{}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		low := strings.ToLower(name)
		if !strings.HasSuffix(low, ".yaml") && !strings.HasSuffix(low, ".yml") {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if base == "" {
			continue
		}
		if _, ok := seen[base]; ok {
			continue
		}
		seen[base] = struct{}{}
		out = append(out, base)
	}
	return out, nil
}

func validateWorkflow(wf Workflow) error {
	if wf.Name == "" {
		return errors.New("name is required")
	}
	if len(wf.Steps) == 0 {
		return errors.New("steps is required")
	}

	seenIDs := map[string]struct{}{}
	for i, s := range wf.Steps {
		if s.ID == "" {
			return fmt.Errorf("steps[%d].id is required", i)
		}
		if _, ok := seenIDs[s.ID]; ok {
			return fmt.Errorf("duplicate step id: %s", s.ID)
		}
		seenIDs[s.ID] = struct{}{}

		switch s.Type {
		case "input", "gemini", "save", "clipboard":
		default:
			return fmt.Errorf("steps[%d].type must be one of: input, gemini, save, clipboard", i)
		}
	}
	return nil
}
