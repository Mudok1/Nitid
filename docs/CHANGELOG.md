# Changelog

All notable changes to this project are documented in this file.

## How to update this file

When you change behavior, docs, or file conventions, add an entry in
`## Unreleased` in the same pull request or commit.

- Use `### Added`, `### Changed`, `### Fixed`, or `### Removed`.
- Write short bullets that describe user-facing impact.
- Keep items in present tense and focus on what changed.

## Unreleased

No changes yet.

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
