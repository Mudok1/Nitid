# Nitid CLI MVP

This page explains the current MVP behavior for `ntd`. Use it as the source of
truth for commands, arguments, and expected results.

For change history, see `docs/CHANGELOG.md`.

## Why this MVP exists

The MVP focuses on fast note capture and simple note lifecycle management. You
can initialize a vault, capture notes, triage them into domains, manage tags,
and archive old notes without any external services.

## Commands you can run now

Use these commands to work with notes from the terminal.

- `ntd version` prints the current CLI version string.
- `ntd init [path]` creates the vault structure and `.nitid/config.toml`.
- `ntd capture [text] [--title "..."] [--domain <id>] [--tags t1,t2] [--kind note|adr|snippet|daily]` creates a note.
- `ntd new <template> [text] [--title "..."] [--domain <id>] [--tags t1,t2]` creates notes from built-in templates.
- `ntd daily [--date YYYY-MM-DD] [--edit]` creates or opens a daily note.
- `ntd templates` and `ntd templates show <name>` list and inspect available templates.
- `ntd ls [--domain <id>] [--tag <tag>] [--status inbox|active|archived] [--kind note|adr|snippet|daily] [--sort updated|created|title|id] [--asc]` lists notes.
- `ntd ls --long` lists notes with full file paths and full IDs.
- `ntd find <query> [--domain <id>] [--tag <tag>] [--status inbox|active|archived] [--kind note|adr|snippet|daily] [--limit N]` searches note text and metadata.
- `ntd move <id|@ref> --domain <domain_id>` moves a note from inbox or another domain into a domain.
- `ntd tag <id|@ref> add|rm <tag>` adds or removes one tag.
- `ntd archive <id|@ref>` moves a note to archive.
- `ntd delete <id|@ref> --yes` permanently deletes a note file.
- `ntd show <id|@ref>` prints note metadata and body in the terminal.
- `ntd show <id|@ref> --raw` prints the raw markdown file exactly as stored.
- `ntd edit <id|@ref>` opens a note in your terminal editor.
- `ntd clean [--dry-run]` removes editor temporary files from `notes/`.
- `ntd validate` checks notes for parse issues, duplicate IDs, and path mismatches.
- `ntd doctor` runs quick environment and vault health checks.
- `ntd tui` opens the interactive three-panel terminal interface.

Inside `ntd tui`, you can edit note bodies directly without leaving the TUI
(`e` to edit, `Ctrl+S` to save, `Esc` to cancel).

## Exit codes

The CLI keeps exit behavior simple so scripts are easy to write.

- `0` means success.
- `1` means invalid arguments, validation errors, or file system errors.

## Common workflows

These examples show the most common day-to-day flow.

```bash
ntd version
ntd init .
ntd capture "Investigate flaky test in CI"
ntd templates
ntd new adr --title "Use ULID for notes"
ntd daily --edit
ntd ls --status inbox
ntd ls --sort title --asc
ntd find flaky --limit 10
ntd move @1 --domain engineering
ntd tag @1 add go
ntd show @1
ntd show @1 --raw
ntd edit @1
ntd delete @1 --yes
ntd clean --dry-run
ntd validate
ntd doctor
ntd tui
ntd archive @1
```

You can also capture from stdin when that is faster.

```bash
echo "Quick note from pipe" | ntd capture --tags go,cli
```

## Use `ntd` directly

If you do not want to run `go run .` every time, build and install a local
binary.

```bash
go build -o ~/.local/bin/ntd .
ntd help
```

If `ntd` is not found, add `~/.local/bin` to your `PATH` in your shell config.

## Command completion

You can enable Bash completion so command names and note selectors are easier to
use.

```bash
source <(ntd completion bash)
```

After this, `tab` completion works for commands and for `@ref` or full note IDs
in commands like `ntd show`, `ntd move`, `ntd tag`, and `ntd archive`.

## Why `#1` can fail in shell

In Bash, an unquoted `#` starts a comment. That means `ntd show #1` can be
parsed as `ntd show`.

Use one of these options:

- Preferred: `ntd show @1`
- Quoted hash ref: `ntd show '#1'`
- Full ID or unique prefix: `ntd show 01KJ9PJ4`
