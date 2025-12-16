# Pals GemFlows

Run pre-made AI workflows from a folder of YAML files.

## For humans (quick start)

1) Get the zip bundle from the person who shared it.
2) Unzip it.
3) Start it:

- macOS: double-click `Run Pals-GemFlows.command`
- Or in Terminal:

```bash
./pals-gemflows run
```

It will show a list and you just pick a workflow.

First run: it will ask you for your Gemini API key and can save it for next time.

Then run a workflow (examples are included in the `workflows/` folder):

```bash
./pals-gemflows run scoping_application
```

For very long meeting transcripts: copy the full transcript to your clipboard first.
The `scoping_application` workflow reads the transcript from your clipboard to avoid terminal paste limits.

If you forget the workflow name, just run `./pals-gemflows` and it will list what's available.

## Technical details (for workflow authors)

The runtime looks for workflows in `./workflows` by default.

You can point to a different folder:

```bash
./pals-gemflows run --workflows-dir /path/to/workflows scoping_application
```

## Workflow location

- `workflows/<name>.yaml` or `workflows/<name>.yml`

## Step types

- `input`: prompts the user and reads input from stdin.
- `gemini`: calls Gemini (Google GenAI) using `GEMINI_API_KEY`.
- `save`: writes a file to disk.
- `clipboard`: copies `content` to your system clipboard.

Optional step fields:

- `multiline: true` (only for `input`): reads until EOF (Ctrl-D on macOS/Linux).
- `from_clipboard: true` (only for `input`): reads the step value from your clipboard (best for long texts).
- `parallel_group: <name>`: consecutive `gemini` steps with the same `parallel_group` run concurrently.

Example:

```yaml
  - id: copy_result
    type: clipboard
    content: "{{ ai_process }}"
```

## Templating (Memory & Injection)

Every step stores its output in memory under its `id`. Any later step can reference it using Mustache-style placeholders:

- `{{ step_id }}` or `{{step_id}}`

Example:

```yaml
steps:
  - id: get_topic
    type: input
    prompt: "What topic?"

  - id: draft_email
    type: gemini
    model: "gemini-1.5-flash"
    user_prompt: "Write a professional email about {{ get_topic }}."

  - id: polish_email
    type: gemini
    model: "gemini-1.5-flash"
    user_prompt: "Make this more friendly: {{ draft_email }}"
```

## Analytics

If `POSTHOG_API_KEY` is set, the engine emits `step_completed` after each step with:
- `workflow_name`, `step_id`, `step_type`, `duration_ms`, `user_machine`

## Packaging (for developers)

```bash
make release
```

This creates a zip in `dist/` containing the runtime + README + sample workflows.
