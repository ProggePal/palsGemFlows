# Pals GemFlows

Run pre-made AI workflows from a folder of YAML files.

## For humans (quick start)

1) Get the zip bundle from the person who shared it.
2) Unzip it.
3) Start it:

Note: In this README, the words `bash` and `powershell` are just labels for formatting. You do **not** type `bash`/`powershell` — you only type the commands inside the box.

- macOS: double-click `Run Pals-GemFlows.command`
- Or in macOS Terminal (copy/paste this, then press Enter):

```bash
cd /path/to/unzipped/folder && ./pals-gemflows run
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

## Install on another computer

You receive a `.zip` bundle containing:

- `pals-gemflows` (the executable)
- `Run Pals-GemFlows.command` (macOS one-click launcher)
- `README.md`, `.env.example`, and `workflows/` examples

### macOS

1) Unzip the bundle.
2) Run it:

- Easiest: double-click `Run Pals-GemFlows.command`
- Or Terminal:

```bash
cd /path/to/unzipped/folder
./pals-gemflows run
```

If macOS blocks it (Gatekeeper):

- Try: right-click `Run Pals-GemFlows.command` → **Open** → **Open**.
- Or: System Settings → **Privacy & Security** → scroll to the warning → **Open Anyway**.

If the file was downloaded and macOS still blocks it, you can remove the quarantine attribute for the unzipped folder:

```bash
cd /path/to/unzipped/folder
xattr -dr com.apple.quarantine .
```

### Windows

1) Unzip the bundle.
2) Run it from PowerShell:

```powershell
cd path\to\unzipped\folder
.\pals-gemflows.exe run
```

If Windows SmartScreen warns you, choose **More info** → **Run anyway**.

### Linux

1) Unzip the bundle.
2) Ensure it’s executable and run:

```bash
cd /path/to/unzipped/folder
chmod +x ./pals-gemflows
./pals-gemflows run
```

## Technical details (for workflow authors)

### Remote recipes vs local dev files

This tool separates the **Engine** (the binary) from the **Content** (YAML recipes).

- Standard mode (remote): run by name (e.g. `marketing/blog_post`) and the tool fetches the YAML from a remote catalog.
- Dev mode (local): run a local file path (e.g. `./my_test.yaml`) to test changes before publishing.

Remote catalog configuration:

- By default, this project fetches from:
  `https://raw.githubusercontent.com/ProggePal/palsGemFlows/main/workflows/`
- To use a different catalog, set `PALSGEMFLOWS_RECIPES_BASE_URL` to a GitHub Raw base URL that points at your catalog folder (for this repo, that’s `workflows/`).
  Example:
  `https://raw.githubusercontent.com/MYORG/automation-platform/main/recipes/`

Examples:

```bash
# Remote recipe (fetched at runtime)
./pals-gemflows run marketing/blog_post

# Local file (dev/testing)
./pals-gemflows run ./my_test.yaml
```

The runtime looks for workflows in `./workflows` by default.

You can point to a different folder:

```bash
./pals-gemflows run --workflows-dir /path/to/workflows scoping_application
```

## Settings

Environment variables:

- `GEMINI_API_KEY` (required for `gemini` steps)
- `PALSGEMFLOWS_RECIPES_BASE_URL` (remote recipe catalog base URL)
- `PALSGEMFLOWS_WORKFLOWS_DIR` (optional default workflows dir)
- `POSTHOG_API_KEY` (optional analytics)
- `POSTHOG_ENDPOINT` (optional; defaults to `https://us.i.posthog.com`)

Flags:

- `--workflows-dir PATH` (overrides the workflows folder)
- `--recipes-base-url URL` (overrides `PALSGEMFLOWS_RECIPES_BASE_URL` for remote fetch)

First run setup:

- If `GEMINI_API_KEY` is not set and you’re running interactively, the app prompts you to paste it.
- It can optionally write a local `.env` file next to the binary so you don’t have to paste again.

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

To build for a different OS/CPU (cross-compile), set `GOOS`/`GOARCH`:

```bash
GOOS=windows GOARCH=amd64 ./scripts/package.sh
GOOS=darwin GOARCH=arm64 ./scripts/package.sh
GOOS=linux GOARCH=amd64 ./scripts/package.sh
```

## Homebrew Tap (recommended distribution)

This lets teammates install/update with Homebrew:

```bash
brew install ProggePal/homebrew-palsgemflows/pals-gemflows
```

### Maintainers: how to publish a new version

1) Tag a release in GitHub (example tag: `v0.1.0`).
2) Build and upload the release assets:

```bash
VERSION=v0.1.0 ./scripts/build_release_assets.sh
```

Upload the generated `dist/pals-gemflows_v0.1.0_*.zip` files to the GitHub Release.

3) Generate the Homebrew formula for the tap:

```bash
VERSION=v0.1.0 ./scripts/generate_homebrew_formula.sh > pals-gemflows.rb
```

4) In the tap repo, commit it as `Formula/pals-gemflows.rb`.

Tip (from Homebrew docs): use `brew tap-new` to create the tap locally, then push it to GitHub. Homebrew recommends naming the repository starting with `homebrew-` so `brew tap` works nicely.
