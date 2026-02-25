# Nitid CLI MVP

This page explains the current MVP behavior for `ntd`. Use it as the source of
truth for commands, arguments, and expected results.

## Why this MVP exists

The MVP focuses on fast note capture and simple note lifecycle management. You
can initialize a vault, capture notes, triage them into domains, manage tags,
and archive old notes without any external services.

## Commands you can run now

Use these commands to work with notes from the terminal.

- `ntd init [path]` creates the vault structure and `.nitid/config.toml`.
- `ntd capture [text] [--title "..."] [--domain <id>] [--tags t1,t2] [--kind note|adr|snippet|daily]` creates a note.
- `ntd ls [--domain <id>] [--tag <tag>] [--status inbox|active|archived] [--kind note|adr|snippet|daily]` lists notes.
- `ntd ls --long` lists notes with full file paths and full IDs.
- `ntd move <id|@ref> --domain <domain_id>` moves a note from inbox or another domain into a domain.
- `ntd tag <id|@ref> add|rm <tag>` adds or removes one tag.
- `ntd archive <id|@ref>` moves a note to archive.
- `ntd show <id|@ref>` prints note metadata and body in the terminal.
- `ntd edit <id|@ref>` opens a note in your terminal editor.

## Exit codes

The CLI keeps exit behavior simple so scripts are easy to write.

- `0` means success.
- `1` means invalid arguments, validation errors, or file system errors.

## Common workflows

These examples show the most common day-to-day flow.

```bash
ntd init .
ntd capture "Investigate flaky test in CI"
ntd ls --status inbox
ntd move @1 --domain engineering
ntd tag @1 add go
ntd show @1
ntd edit @1
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
