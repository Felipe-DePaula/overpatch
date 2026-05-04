# AGENTS.md

Guidance for AI coding agents working on this repository.

## What this project is

Overpatch is a toolkit that lets AI propose code changes as structured JSON, which a deterministic executor then validates and applies. The project is **about** safe AI-driven editing — so the project itself must be exemplary in that regard.

## Core principles

1. **Atomicity is non-negotiable.** Any change applied to a user's project must be all-or-nothing. No partial application, ever.
2. **Validate before stage. Stage before commit.** Three phases. Never collapse them.
3. **Anchors over line numbers.** Operations match by content, with line numbers as hints only.
4. **Fail loudly with structured errors.** Every failure must produce an error with a code, an operation ID, and enough context for an LLM to retry.
5. **The executor never trusts the LLM.** Schema validation, path safety, occurrence counts — all enforced.

## Repository boundary

Agents must operate strictly inside the repository root.

Before reading or editing files, agents should confirm the repository root with:

```bash
git rev-parse --show-toplevel
git status --short
```

Under no circumstances should agents inspect, list, read, or modify files outside the repository root unless the user explicitly authorizes it for that exact task.

This includes, but is not limited to:

- user home directories
- AppData
- WindowsApps
- Downloads
- Documents
- Desktop
- OneDrive
- parent directories
- sibling directories

If required tooling such as Go is not available in `PATH`, report the missing tool and stop. Do not search outside the repository to locate it.

## When editing this repo

- **Do not** introduce dependencies casually. Each new dependency goes through review. Prefer the standard library.
- **Do not** add `TODO` or `FIXME` comments. Either implement, or open an issue.
- **Do not** create files speculatively ("we might need this later"). Add when needed.
- **Do** keep `internal/` packages cohesive. If a package starts growing helpers, that's a signal to split.
- **Do** write a test alongside any new operation type or safety check.
- **Do** update `docs/PROTOCOL.md` whenever the schema changes, and bump `schema_version` if backward-incompatible.

## When generating Overpatch JSON (the protocol itself)

If you are an LLM being asked to produce an Overpatch document for a user's project (not editing this repo), follow `docs/PROTOCOL.md` exactly. In particular:

- `schema_version` must be `"overpatch/v1"`.
- Every operation must have an `id`, `action`, and `path`.
- For text-matching operations, always include `expected_occurrences` if you know the count from context.
- Never invent files that aren't in the dump.
- If the request is ambiguous or impossible, return `status: "failed"` with a clear `reason`.

## File layout reminders

- `cmd/overpatch/` — CLI entrypoint only. No business logic.
- `internal/` — all implementation. Not importable by other modules.
- `schemas/` — canonical JSON Schema files. Runtime embedding is planned, not implemented yet.
- `examples/` — working examples used in docs and tests.
- `experiments/` — research artifacts. Not shipped. Not stable.
- `docs/` — design documents. Update them when behavior changes.

## Commit messages

Conventional Commits. `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`. Scope optional but encouraged: `feat(executor): add staging phase`.

## Decision log

Significant architectural decisions live in `docs/ARCHITECTURE.md` under "Decisions". When changing course, append — don't rewrite history.
