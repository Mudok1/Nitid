# Nitid Domain and Tag Conventions

This page explains how notes are classified and routed in Nitid. The goal is to
keep decisions predictable so capture is fast and retrieval stays clean.

## Core model

Nitid uses a simple model with one primary location and optional facets.

- Domain is the primary home for a note.
- Tags are cross-cutting facets for search and filtering.
- Todos live in the note body as Markdown checkboxes.

## Domain conventions

Domains are stable buckets that should change rarely.

- Use lowercase kebab-case IDs.
- Keep IDs short and durable, for example `engineering`, `product`, `ops`, and `learning`.
- Keep exactly one domain per note, or no domain while the note is in inbox.

## Tag conventions

Tags help you find related notes across domains.

- Use lowercase kebab-case tags.
- Prefer specific tags such as `race-condition` over generic tags such as `bug`.
- Keep most notes between 2 and 6 tags.

## Routing rules

Routing controls where files live on disk and which status they get.

- New capture without a domain goes to `notes/inbox/` with `status: inbox`.
- Triaged note with a domain goes to `notes/domains/<domain_id>/` with `status: active`.
- Daily note goes to `notes/daily/YYYY/MM/` with `kind: daily`.
- Archived note goes to `notes/archive/` with `status: archived`.

## Domain vs tags examples

Use these examples as a quick reference when you are unsure where a note goes.

- Domain `engineering` with tags `go`, `cli`, `tui`, and `search`.
- Domain `product` with tags `roadmap`, `ux`, and `naming`.
