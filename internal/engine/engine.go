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
	"sync"
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

	for i := 0; i < len(wf.Steps); {
		raw := wf.Steps[i]
		if raw.ParallelGroup != "" {
			group := raw.ParallelGroup
			j := i
			for j < len(wf.Steps) && wf.Steps[j].ParallelGroup == group {
				j++
			}
			if err := e.runParallelGroup(ctx, wf, wf.Steps[i:j], memory); err != nil {
				return err
			}
			i = j
			continue
		}

		step, err := renderStep(raw, memory)
		if err != nil {
			return fmt.Errorf("render step %s: %w", raw.ID, err)
		}

		out, durationMs, err := e.executeStep(ctx, step)
		if err != nil {
			return fmt.Errorf("step %s failed: %w", step.ID, err)
		}

		memory[step.ID] = out
		if e.deps.Analytics != nil {
			e.deps.Analytics.StepCompleted(wf.Name, step.ID, step.Type, durationMs)
		}
		i++
	}

	return nil
}

func (e *Engine) runParallelGroup(ctx context.Context, wf workflow.Workflow, raws []workflow.Step, memory map[string]string) error {
	if len(raws) == 0 {
		return nil
	}
	group := raws[0].ParallelGroup
	fmt.Printf("==> parallel group %q (%d steps)\n", group, len(raws))

	steps := make([]workflow.Step, 0, len(raws))
	for _, raw := range raws {
		if raw.Type != "gemini" {
			return fmt.Errorf("parallel_group %q only supports gemini steps for now (got %s for %s)", group, raw.Type, raw.ID)
		}
		step, err := renderStep(raw, memory)
		if err != nil {
			return fmt.Errorf("render step %s: %w", raw.ID, err)
		}
		steps = append(steps, step)
	}

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		id         string
		stepType   string
		out        string
		durationMs int64
		err        error
	}

	results := make(chan result, len(steps))
	var wg sync.WaitGroup
	for _, step := range steps {
		step := step
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, durationMs, err := e.executeStep(childCtx, step)
			if err != nil {
				cancel()
			}
			results <- result{id: step.ID, stepType: step.Type, out: out, durationMs: durationMs, err: err}
		}()
	}

	wg.Wait()
	close(results)

	for r := range results {
		if r.err != nil {
			return fmt.Errorf("step %s failed: %w", r.id, r.err)
		}
		memory[r.id] = r.out
		if e.deps.Analytics != nil {
			e.deps.Analytics.StepCompleted(wf.Name, r.id, r.stepType, r.durationMs)
		}
	}

	fmt.Printf("<== completed parallel group %q\n\n", group)
	return nil
}

func (e *Engine) executeStep(ctx context.Context, step workflow.Step) (string, int64, error) {
	fmt.Printf("==> step %s (%s)\n", step.ID, step.Type)
	start := time.Now()

	var (
		out string
		err error
	)
	switch step.Type {
	case "input":
		if step.FromClipboard {
			out, err = runInputFromClipboard(step.Prompt)
		} else if step.Multiline {
			out, err = runInputMultiline(step.Prompt)
		} else {
			out, err = runInput(step.Prompt)
		}
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
		return "", durationMs, err
	}

	fmt.Printf("<== completed %s in %dms\n\n", step.ID, durationMs)
	return out, durationMs, nil
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

func runInputMultiline(prompt string) (string, error) {
	if prompt == "" {
		prompt = "Paste input (end with Ctrl-D):"
	}
	fmt.Printf("%s\n", prompt)

	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	text := strings.TrimRight(string(b), "\r\n")
	return text, nil
}

func runInputFromClipboard(prompt string) (string, error) {
	if prompt == "" {
		prompt = "Copy the text you want to use, then press Enter to read from clipboard:"
	}
	fmt.Printf("%s\n", prompt)
	fmt.Print("Press Enter when ready (or Ctrl-C to cancel): ")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	data, err := readClipboard()
	if err != nil {
		return "", err
	}
	data = strings.TrimRight(data, "\r\n")
	if strings.TrimSpace(data) == "" {
		return "", errors.New("clipboard is empty")
	}
	return data, nil
}

func readClipboard() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("pbpaste").Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	case "windows":
		// PowerShell is available on modern Windows installations.
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-Clipboard -Raw").Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	default:
		// Linux: prefer wl-paste, then xclip, then xsel.
		if _, err := exec.LookPath("wl-paste"); err == nil {
			out, err := exec.Command("wl-paste", "-n").Output()
			if err != nil {
				return "", err
			}
			return string(out), nil
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
			if err != nil {
				return "", err
			}
			return string(out), nil
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			out, err := exec.Command("xsel", "--clipboard", "--output").Output()
			if err != nil {
				return "", err
			}
			return string(out), nil
		}
		return "", fmt.Errorf("no clipboard helper found (install wl-paste, xclip, or xsel)")
	}
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
