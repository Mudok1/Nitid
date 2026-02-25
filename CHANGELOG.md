# Changelog

All notable changes to this project are documented in this file.

## Unreleased

### Added
- Initial Nitid CLI structure with `init` and `capture` commands.
- Note schema and routing support with ULID IDs and YAML frontmatter.
- MVP lifecycle commands: `ls`, `move`, `tag`, `archive`, and `show`.
- Added `edit` command to open notes in your terminal editor.
- Cleaner default `ls` table output and `--long` detailed mode.
- `@ref` selectors (for example `@1`) to avoid typing long IDs.
- Bash completion via `ntd completion bash`.
- Starter sample notes for `getting-started` and inbox flows.
- Tests for parsing, routing, and tag updates.
- Detailed command guide at `docs/command-guide.md`.

### Changed
- Documentation rewritten for clarity in `docs/`.
