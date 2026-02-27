# Nitid command guide

This guide explains every current `ntd` command with practical examples.

## Related docs

Use these pages with this guide when you need deeper context.

- `docs/cli-mvp.md` for feature-level behavior.
- `docs/note-schema-v1.md` for metadata rules.
- `docs/domain-tag-conventions.md` for routing rules.
- `docs/development-workflow.md` for changelog and update policy.
- `docs/CHANGELOG.md` for recent project changes.

## Before you start

Run commands from your vault root.

```bash
cd /path/to/nitid
ntd help
```

If `ntd` is not in your `PATH`, build it with:

```bash
go build -o ~/.local/bin/ntd .
```

## Note selectors: ID and `@ref`

Most commands accept either:

- A full ID or unique ID prefix, for example `01KJ9PJ4`.
- A row reference from `ntd ls`, for example `@1`.

Use `@ref` for speed. It avoids shell issues with `#` comments.

## Command by command

### `ntd version`

Print the current CLI version.

```bash
ntd version
```

### `ntd init [path]`

Create the Nitid folder structure and default config.

```bash
ntd init .
```

### `ntd capture [text] [flags]`

Create a new note.

Flags:

- `--title`: set title manually.
- `--domain`: route directly to a domain.
- `--tags`: comma-separated tags.
- `--kind`: `note`, `adr`, `snippet`, or `daily`.

Examples:

```bash
ntd capture "Fix race condition in worker pool"
ntd capture --domain engineering --tags go,debug "Found root cause"
echo "quick capture from pipe" | ntd capture --tags scratchpad
```

### `ntd templates` and `ntd templates show <name>`

Inspect built-in templates.

```bash
ntd templates
ntd templates show adr
```

### `ntd new <template> [text] [flags]`

Create a note from a template.

Supported templates: `note`, `adr`, `meeting`, `bug`.

Flags:

- `--title`: set title manually.
- `--domain`: route directly to a domain.
- `--tags`: comma-separated tags.

```bash
ntd new adr --title "Use ULID for note IDs"
ntd new bug --domain engineering "panic in config parse"
```

### `ntd daily [--date YYYY-MM-DD] [--edit]`

Create or reuse a daily note for a date.

```bash
ntd daily
ntd daily --date 2026-02-25 --edit
```

### `ntd ls [flags]`

List notes in a readable table.

Flags:

- `--domain <id>`
- `--tag <tag>`
- `--status inbox|active|archived`
- `--kind note|adr|snippet|daily`
- `--long` for full IDs and paths
- `--sort updated|created|title|id`
- `--asc` for ascending sort order

Examples:

```bash
ntd ls
ntd ls --status inbox
ntd ls --domain engineering --tag go
ntd ls --sort title --asc
ntd ls --long
```

### `ntd find <query> [flags]`

Search notes by title, body, domain, and tags.

Flags:

- `--domain <id>`
- `--tag <tag>`
- `--status inbox|active|archived`
- `--kind note|adr|snippet|daily`
- `--limit N`

```bash
ntd find goroutine
ntd find flaky --status inbox --limit 10
```

### `ntd show <id|@ref>`

Print full metadata and body for one note.

```bash
ntd show @1
ntd show 01KJ9PJ4
ntd show @1 --raw
```

### `ntd edit <id|@ref>`

Open a note in your terminal editor.

Editor resolution order:

1. `$VISUAL`
2. `$EDITOR`
3. fallback to `nano` (or `vi` if `nano` is not available)

Examples:

```bash
ntd edit @1
EDITOR=nano ntd edit @2
```

### `ntd clean [--dry-run]`

Remove editor temporary files from `notes/`.

This command targets common leftovers such as `.swp`, `.swo`, and `~` files.

```bash
ntd clean --dry-run
ntd clean
```

### `ntd validate`

Validate all notes in `notes/`.

This command checks parsing, duplicate IDs, and mismatches between note
metadata and expected file location.

```bash
ntd validate
```

### `ntd doctor`

Run environment checks and quick vault health diagnostics.

This command checks your notes directory, editor availability, completion
command availability, and validation summary.

```bash
ntd doctor
```

### `ntd tui`

Open the interactive TUI with list, preview, and metadata panels.

Useful keys inside TUI:

- `j` / `k`: move selection.
- `/`: start a quick `find` command.
- `:`: open command mode.
- `e`: edit selected note body directly inside TUI.
- `Ctrl+S`: save while editing.
- `Esc`: cancel editing.
- `a`: archive selected note (with confirmation).
- `q`: quit TUI.

```bash
ntd tui
```

### `ntd move <id|@ref> --domain <domain_id>`

Move a note into a domain and set it active.

```bash
ntd move @1 --domain engineering
```

### `ntd tag <id|@ref> add|rm <tag>`

Add or remove a single tag.

```bash
ntd tag @1 add concurrency
ntd tag @1 rm scratchpad
```

### `ntd archive <id|@ref>`

Move a note to archive and set status to archived.

```bash
ntd archive @1
```

### `ntd delete <id|@ref> --yes`

Permanently delete a note file.

This command requires `--yes` (or `-y`) to reduce accidental deletions.

```bash
ntd delete @1 --yes
```

### `ntd completion bash`

Enable command and selector completion in Bash.

Temporary for current shell:

```bash
source <(ntd completion bash)
```

Persistent in Bash:

```bash
echo 'source <(ntd completion bash)' >> ~/.bashrc
```

## Typical daily flow

```bash
ntd capture "Investigate flaky test"
ntd ls --status inbox
ntd move @1 --domain engineering
ntd tag @1 add testing
ntd show @1
ntd edit @1
ntd tui
```

## Troubleshooting

- If `ntd show #1` fails, use `ntd show @1`.
- If command not found, rebuild and ensure `~/.local/bin` is in your `PATH`.
- If completion does not work, re-run `source <(ntd completion bash)`.
- If you see Vim swap files (`.swp`), run `ntd clean`.
