# Roadmap

## v0.1 — Executor MVP

The first boss. Goal: a working CLI that applies a JSON document to a local directory.

- [~] Project skeleton (Go module, Makefile, lint config)
- [~] Schema structs and JSON Schema embedded
- [~] CLI commands: `validate`, `plan`, `apply`, `inspect`, `version`
- [ ] Actions: `replace_text`, `replace_lines`, `insert_before_lines`, `insert_after_lines`, `create`, `delete`
- [ ] Three-phase commit (validate → stage → commit)
- [ ] Path safety (traversal, absolute, blocklist)
- [ ] Unified diff output
- [~] Integration test against a fixture project
- [ ] First release: `overpatch-{linux,darwin,windows}-{amd64,arm64}` binaries

## v0.2 — Audit and polish

- [ ] `.overpatch/runs/<timestamp>/` log directory
- [ ] Structured error output (`--output=json`)
- [ ] Color-aware terminal output, `--no-color` flag
- [ ] Better diff rendering (side-by-side option)
- [ ] `overpatch report <ops.json>` post-apply summary

## v0.3 — Git integration

- [ ] Refuse to apply on dirty working tree (default)
- [ ] `--force-dirty` escape hatch
- [ ] `overpatch rollback` (one-step `git reset --hard ORIG_HEAD`)
- [ ] Suggest `git diff` after apply

## v0.4 — Schema v1.1

- [ ] `expected_match_count` (rename from `expected_occurrences` for line-based ops)
- [ ] `hint_line` for disambiguation when content alone isn't unique
- [ ] `pre_hash` optional file integrity check
- [ ] `rationale` per operation, surfaced in audit logs
- [ ] Backward-compat: still accepts v1 documents

## v0.5 — Context Builder

- [ ] `overpatch dump` generates context pack
- [ ] Respects `.gitignore`
- [ ] Configurable blocklist (`node_modules`, `dist`, etc.)
- [ ] Per-file size cap, total pack size cap
- [ ] Binary detection (null byte heuristic)
- [ ] Secret detection (gitleaks integration)
- [ ] Line numbering in dump (for hint_line support)

## v0.6 — Provider Gateway

- [ ] Provider interface with adapters:
  - `manual` — write context to file, expect JSON back
  - `anthropic` — Claude API
  - `openai` — GPT API
  - `ollama` — local HTTP
- [ ] `overpatch plan --prompt "..." --provider=anthropic`
- [ ] Retry loop with structured error feedback to the model

## v0.7 — Smart context

- [ ] Embeddings-based file ranking (Ollama + `nomic-embed-text`)
- [ ] Tree-sitter signatures for non-included files
- [ ] `overpatch dump --smart "<prompt>"` selects relevant subset

## v1.0 — Stable toolkit

- [ ] Schema v1 frozen
- [ ] Distribution: Homebrew tap, Scoop bucket, signed binaries
- [ ] Full documentation
- [ ] Real-world example projects in `examples/projects/`
