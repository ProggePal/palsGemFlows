# Workflow Authoring Guide

A workflow is a YAML file placed in your workflows directory (default `./workflows`).

Run a workflow:

```bash
./my-tool run <workflow_name>
# or
./my-tool run --workflows-dir /path/to/workflows <workflow_name>
```

## File naming

- `workflows/<name>.yaml` or `workflows/<name>.yml`
- You run it with `my-tool run <name>` (extension optional)

## Schema

Top-level:

```yaml
name: "Human readable name"
description: "Optional description"
steps:
  - id: some_step
    type: input|gemini|save|clipboard
    ...
```

Each step output is saved in memory under its `id`.

## Templating (Data Passing)

Use Mustache-style placeholders to reference earlier outputs:

- `{{ step_id }}`
- `{{step_id}}`

If a placeholder references a missing step id, the run fails (MVP behavior).

## Step types

### 1) `input`
Prompts the user and reads input.

```yaml
- id: transcript
  type: input
  prompt: "Paste transcript:"
```

Multiline input (read until EOF / Ctrl-D):

```yaml
- id: transcript
  type: input
  multiline: true
  prompt: "Paste transcript, end with Ctrl-D:"
```

### 2) `gemini`
Calls Gemini and returns generated text.

```yaml
- id: ai_process
  type: gemini
  model: "gemini-1.5-flash"
  system_prompt: "You are a professional editor."
  user_prompt: "Fix the grammar: {{ transcript }}"
```

### 3) `save`
Writes `content` to a file. The step output is the filename.

```yaml
- id: save_result
  type: save
  filename: "result.md"
  content: "{{ ai_process }}"
```

### 4) `clipboard`
Copies `content` to the system clipboard. The step output is `copied`.

```yaml
- id: copy_result
  type: clipboard
  content: "{{ ai_process }}"
```

## Parallel Gemini steps (optional)

If you set `parallel_group` on consecutive `gemini` steps, they run concurrently:

```yaml
- id: ui
  type: gemini
  parallel_group: analyze
  model: "gemini-1.5-flash"
  user_prompt: "UI summary for: {{ transcript }}"

- id: technical
  type: gemini
  parallel_group: analyze
  model: "gemini-1.5-flash"
  user_prompt: "Technical summary for: {{ transcript }}"
```

All outputs are stored in memory and can be referenced in later steps.

## Environment variables

- `GEMINI_API_KEY` (required for `gemini` steps)
- `POSTHOG_API_KEY` (optional analytics)
- `POSTHOG_ENDPOINT` (optional; defaults to `https://us.i.posthog.com`)
- `MY_TOOL_WORKFLOWS_DIR` (optional default workflows dir)
