# CLI GPT Flows (MVP)

This repo builds a single Go binary (`my-tool`) that runs YAML workflows from `workflows/`.

## Run

```bash
export GEMINI_API_KEY="..."
export POSTHOG_API_KEY="..."  # optional
# Optional override for PostHog cloud region/self-hosted ingest
# export POSTHOG_ENDPOINT="https://eu.i.posthog.com"

go run ./cmd/my-tool run grammar_fix
```

You can also point to a different workflows folder:

```bash
go run ./cmd/my-tool run --workflows-dir /path/to/workflows grammar_fix
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

## Sharing with colleagues

Create a zip bundle containing the runtime + docs + example workflows:

```bash
make release
```

Share the generated zip in `dist/`. Your colleague can unzip, set env vars (see `.env.example`), and run:

```bash
./my-tool run grammar_fix
```

Workflow authoring guide: [docs/WORKFLOWS.md](docs/WORKFLOWS.md)
