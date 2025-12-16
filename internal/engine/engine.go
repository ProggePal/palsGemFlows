package engine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"cli-gpt-flows/internal/analytics"
	"cli-gpt-flows/internal/gemini"
	"cli-gpt-flows/internal/templating"
	"cli-gpt-flows/internal/workflow"
)

type Dependencies struct {
	Gemini    *gemini.Client
	Analytics *analytics.Client
}

type Engine struct {
	deps Dependencies
}

func New(deps Dependencies) *Engine {
	return &Engine{deps: deps}
}

func (e *Engine) Run(ctx context.Context, wf workflow.Workflow) error {
	memory := map[string]string{}

	for _, raw := range wf.Steps {
		step, err := renderStep(raw, memory)
		if err != nil {
			return fmt.Errorf("render step %s: %w", raw.ID, err)
		}

		fmt.Printf("==> step %s (%s)\n", step.ID, step.Type)
		start := time.Now()

		var out string
		switch step.Type {
		case "input":
			out, err = runInput(step.Prompt)
		case "gemini":
			if e.deps.Gemini == nil {
				err = errors.New("gemini client is not configured")
				break
			}
			out, err = e.deps.Gemini.Generate(ctx, step.Model, step.SystemPrompt, step.UserPrompt)
		case "save":
			out, err = runSave(step.Filename, step.Content)
		case "clipboard":
			out, err = runClipboard(step.Content)
		default:
			err = fmt.Errorf("unsupported step type: %s", step.Type)
		}

		durationMs := time.Since(start).Milliseconds()
		if err != nil {
			return fmt.Errorf("step %s failed: %w", step.ID, err)
		}

		memory[step.ID] = out

		if e.deps.Analytics != nil {
			e.deps.Analytics.StepCompleted(wf.Name, step.ID, step.Type, durationMs)
		}

		fmt.Printf("<== completed %s in %dms\n\n", step.ID, durationMs)
	}

	return nil
}

func renderStep(step workflow.Step, memory map[string]string) (workflow.Step, error) {
	var err error
	step.Prompt, err = templating.RenderString(step.Prompt, memory)
	if err != nil {
		return workflow.Step{}, err
	}
	step.UserPrompt, err = templating.RenderString(step.UserPrompt, memory)
	if err != nil {
		return workflow.Step{}, err
	}
	step.SystemPrompt, err = templating.RenderString(step.SystemPrompt, memory)
	if err != nil {
		return workflow.Step{}, err
	}
	step.Model, err = templating.RenderString(step.Model, memory)
	if err != nil {
		return workflow.Step{}, err
	}
	step.Filename, err = templating.RenderString(step.Filename, memory)
	if err != nil {
		return workflow.Step{}, err
	}
	step.Content, err = templating.RenderString(step.Content, memory)
	if err != nil {
		return workflow.Step{}, err
	}
	return step, nil
}

func runInput(prompt string) (string, error) {
	if prompt == "" {
		prompt = "Input:" // fallback
	}
	fmt.Printf("%s ", prompt)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		// If user enters EOF without newline (Ctrl-D), ReadString returns data + io.EOF.
		// Treat that as valid input if we got any content.
		if !errors.Is(err, io.EOF) {
			return "", err
		}
	}

	line = strings.TrimRight(line, "\r\n")
	return line, nil
}

func runSave(filename string, content string) (string, error) {
	if filename == "" {
		return "", errors.New("filename is required")
	}
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", err
	}
	return filename, nil
}

func runClipboard(content string) (string, error) {
	if content == "" {
		return "", errors.New("content is required")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("cmd", "/c", "clip")
	default:
		// Prefer wl-copy if available, then xclip, then xsel.
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return "", fmt.Errorf("no clipboard helper found (install wl-copy, xclip, or xsel)")
		}
	}

	cmd.Stdin = strings.NewReader(content)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return "copied", nil
}
