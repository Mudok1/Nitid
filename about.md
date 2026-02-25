Project: Nitid (CLI: ntd)

Nitid is a “second brain” for developers, built around a fast, delightful TUI (Terminal UI) experience. It helps capture ideas, technical decisions, snippets, tasks, and learnings—then turns them into connected, actionable knowledge. The goal is to minimize friction when writing notes and maximize clarity when retrieving them: transform scattered notes into a system that genuinely helps you think and ship.

What we’re building with Nitid:
- Rapid capture: save thoughts in seconds without breaking flow.
- TUI-first experience: navigate, edit, and search entirely in the terminal; simple and efficient UI.
- Flexible organization: tags, links/backlinks, minimal structure (git-friendly).
- Strong retrieval: text + semantic search, summaries, and task-oriented “recalls.”
- AI integration: help organize the chaos (auto-classify, tag suggestions, duplicate detection, related-note linking, summaries/ADRs, TODO extraction, structure suggestions).
- Developer-focused: natural support for code, architecture decisions, debugging notes, ADRs, and living documentation.
- Local-first & open-source: privacy, portability, and user control.

Principles:
- Simplicity over complexity (Unix-like: small actions done well).
- Human-readable data (e.g., Markdown) that’s easy to version.
- Predictable outputs: don’t invent features; if something doesn’t exist yet, propose it as design.
- Use AI to automate repetitive work without hiding data—the user should understand and edit everything.

Non-goals:
- Not a social network or a closed SaaS.
- Not a heavy GUI-first app: the TUI is the core; optional UIs can come later.

Agent role:
Help design and implement Nitid while staying aligned with these goals. When proposing changes, prioritize low coupling, perceived speed in the TUI, engineering best practices, tests, clear docs, and an excellent terminal experience. If details are missing, make minimal assumptions and state them explicitly.