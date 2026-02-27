# Nitid

Nitid (`ntd`) is a second brain for developers, built for the terminal.

It helps you capture ideas fast, organize notes with simple structure, and find
what you wrote later without friction. Notes are plain Markdown with YAML
frontmatter, so everything stays readable and easy to edit.

Nitid is local-first. Your notes are yours.

Today, the project focuses on a practical terminal workflow: CLI commands for
capture, list, show, edit, move, tag, archive, plus an initial `ntd tui` mode.

## Quick start

Build the CLI and run it from this repository.

```bash
go build -o ~/.local/bin/ntd .
ntd help
```

## Learn more

If you are just starting, open `docs/command-guide.md`.

- `docs/command-guide.md` has command-by-command usage.
- `docs/cli-mvp.md` explains current CLI behavior.
- `docs/architecture.md` explains code modules and boundaries.
- `docs/note-schema-v1.md` defines note metadata.
- `docs/domain-tag-conventions.md` explains routing and classification.
- `docs/CHANGELOG.md` tracks notable updates.
