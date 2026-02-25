# Nitid Note Schema v1

Nitid stores each note as a Markdown file with YAML frontmatter. This schema is
designed to be human-readable, Git-friendly, and easy to validate from the CLI.

## Goals

This schema keeps capture simple now while staying stable for future indexing.

- Keep note files readable in plain text.
- Keep metadata predictable for CLI commands.
- Keep defaults clear so quick capture still produces valid notes.

## Required frontmatter fields

Every note must include the fields below.

- `id`: immutable ULID.
- `title`: short human-readable title.
- `created_at`: RFC3339 UTC timestamp.
- `updated_at`: RFC3339 UTC timestamp.
- `domain`: primary domain ID, or an empty string while in inbox.
- `tags`: list of lowercase kebab-case tags.
- `status`: `inbox`, `active`, or `archived`.
- `kind`: `note`, `adr`, `snippet`, or `daily`.
- `links`: list of note IDs or other references.

## File naming and paths

Nitid separates identity from readability in filenames.

- Filename format is `<ulid>--<slug>.md`.
- The slug can change over time, but the ULID never changes.
- Routing rules:
  - `status: inbox` stores the note in `notes/inbox/`.
  - `status: active` with a domain stores the note in `notes/domains/<domain_id>/`.
  - `kind: daily` stores the note in `notes/daily/YYYY/MM/`.
  - `status: archived` stores the note in `notes/archive/`.

## Validation rules

Validation happens on write so bad metadata does not spread.

- `id` must be a valid ULID.
- `domain` must match lowercase kebab-case: `^[a-z0-9]+(?:-[a-z0-9]+)*$`.
- Every tag must match lowercase kebab-case.
- `title` cannot be empty.
- `status` and `kind` must be one of the allowed values.

## Example

Use this example as a template for manual edits.

```yaml
---
id: "01JN8PX5WP8J67JAY2P2CVJH6D"
title: "Investigate worker pool leak"
created_at: "2026-02-25T10:20:31Z"
updated_at: "2026-02-25T10:20:31Z"
domain: "engineering"
tags: ["go", "debug", "concurrency"]
status: "active"
kind: "note"
links: []
---
```
