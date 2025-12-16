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

## Workflow location

- `workflows/<name>.yaml` or `workflows/<name>.yml`

## Step types

- `input`: prompts the user and reads one line from stdin.
- `gemini`: calls Gemini (Google GenAI) using `GEMINI_API_KEY`.
- `save`: writes a file to disk.
- `clipboard`: copies `content` to your system clipboard.

Example:

```yaml
	- id: copy_result
		type: clipboard
		content: "{{ ai_process }}"
```

## Analytics

If `POSTHOG_API_KEY` is set, the engine emits `step_completed` after each step with:
- `workflow_name`, `step_id`, `step_type`, `duration_ms`, `user_machine`
