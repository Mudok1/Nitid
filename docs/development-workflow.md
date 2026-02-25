# Nitid development workflow

This page defines how changes are documented and shipped in Nitid. Use it to
keep code, docs, and changelog updates consistent.

## Why this workflow exists

Nitid is built incrementally, with an emphasis on predictable behavior and
clear docs. This workflow keeps changes understandable for contributors and for
future sessions.

## Required updates for every change

When you change behavior, command output, schema, or routing, update all
relevant docs in the same work item.

- Update feature docs in `docs/`.
- Update command examples when command behavior changes.
- Update `docs/CHANGELOG.md` under `## Unreleased`.

## Changelog rules

The changelog is user-facing and must stay current.

- Add entries in one of these sections: `Added`, `Changed`, `Fixed`,
  `Removed`.
- Write concise bullets that describe user impact.
- Keep entries in present tense.
- Do not defer changelog updates to a later commit.

## Documentation rules

Documentation must match code exactly.

- Keep command syntax aligned with `ntd help`.
- Keep examples runnable from the repository root.
- Prefer short, direct explanations over abstract wording.
- Add troubleshooting notes when behavior can surprise users.

## Release preparation checklist

Before pushing changes, run this checklist:

1. Run `go test ./...`.
2. Run `go vet ./...`.
3. Run `ntd help` and verify command docs still match.
4. Verify new files are intentionally tracked by Git.
5. Update `docs/CHANGELOG.md`.

## Notes and sample content policy

Repository tracking is intentionally limited for note content.

- Keep only starter content in `notes/domains/getting-started/` tracked.
- Keep personal and test notes local.
- Use `.gitignore` rules to prevent accidental note uploads.
