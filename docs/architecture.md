# Nitid architecture

This page explains how Nitid is organized so you can understand and extend the
project in parts.

## Overview

Nitid follows a simple modular layout inside a single Go binary.

- **Command module (`internal/cli`)** handles CLI parsing, command routing, and
  user-facing output.
- **Core module (`internal/core`)** handles shared application workflows used by
  both CLI commands and TUI actions.
- **Vault module (`internal/vault`)** handles note storage, parsing,
  validation, and filesystem routing.
- **Entry module (`main.go`)** keeps startup and process exit logic minimal.

This split keeps responsibilities clear while preserving a small codebase.

## Module boundaries

### Command module

The command module translates user input into operations.

- Validates command arguments.
- Calls vault and feature functions.
- Prints user-facing output.

It does not own file format details.

### Vault module

The vault module is the data core.

- Defines the note model.
- Reads and writes markdown + YAML frontmatter.
- Enforces ID, status, domain, and tag rules.
- Resolves where notes belong in the vault tree.

It does not parse command flags or shell behavior.

### Core module

The core module is the shared use-case layer.

- Exposes note operations such as create, list, find, move, tag, archive, and
  edit.
- Keeps user-interface logic out of storage code.
- Gives CLI and TUI one consistent behavior path.

It does not render terminal views.

### Feature module

The feature module orchestrates multi-step workflows.

- Template rendering for `new`.
- Daily note date handling.
- Editor launching helpers.
- Validation and diagnostics aggregation for `validate` and `doctor`.

These helpers live in `internal/cli/features.go` and use `internal/vault`
interfaces and types.

## Data flow

For most commands, execution flow is:

1. Command module parses input.
2. Feature module (optional) derives workflow data.
3. Core module executes shared use-cases.
4. Vault module applies data rules and filesystem operations.
5. Command or TUI layer renders output.

## Why this structure

This architecture keeps Nitid easy to reason about.

- You can add commands without touching low-level file parsing.
- You can add UI surfaces (CLI and TUI) without duplicating workflows.
- You can evolve note schema logic without rewriting command handlers.
- You can test workflows and storage separately.
