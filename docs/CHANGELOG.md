# Changelog

All notable changes to this project are documented in this file.

## How to update this file

When you change behavior, docs, or file conventions, add an entry in
`## Unreleased` in the same pull request or commit.

- Use `### Added`, `### Changed`, `### Fixed`, or `### Removed`.
- Write short bullets that describe user-facing impact.
- Keep items in present tense and focus on what changed.

## Unreleased

### Added
- `clean` command to remove editor temporary files from `notes/`.
- `version` command to display the CLI version string.
- `find` command to search notes by text and metadata.
- `validate` command to check parse integrity, duplicate IDs, and path consistency.
- `show --raw` mode to print raw note files.
- `templates` command to list and inspect built-in note templates.
- `new` command to create notes from templates (`note`, `adr`, `meeting`, `bug`).
- `daily` command to create or reuse daily notes by date.
- `doctor` command for environment and vault health checks.
- `tui` command with a first Bubble Tea interface and three-panel layout.
- `internal/core` service layer for shared note operations used by CLI and TUI.

### Changed
- `edit` command now falls back to `nano` before `vi` when no editor is set.
- Command docs updated with selector and cleanup troubleshooting guidance.
- `ls` now supports `--sort` and `--asc` for explicit ordering.
- Codebase reorganized into `internal/cli` and `internal/vault` modules.
- Tests split by module (`internal/cli`, `internal/vault`) plus root integration tests.
- CLI command handlers now route key note workflows through shared core services.

## [0.1.0] - 2026-02-25

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
- Development process guide at `docs/development-workflow.md`.
- Local `about.md` project context file for agents and sessions.

### Changed
- Documentation rewritten for clarity in `docs/`.
- Project overview moved from `about.md` to `README.md`.
- Changelog moved to `docs/CHANGELOG.md`.
- Simplified `README.md` language and presentation.
