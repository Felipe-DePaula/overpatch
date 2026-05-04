# Roadmap

Legend:

- ✅ Done
- 🟡 In progress
- ⬜ Pending

## v0.1: Executor MVP

The first boss. Goal: a working CLI that validates, previews, and applies an Overpatch JSON document to a local directory.

- 🟡 Project skeleton: Go module, Cobra CLI, Windows manifest, and basic repository structure are done. Makefile and lint configuration are pending.
- 🟡 Schema structs and validation: structs, parser, structural validation, action-specific validation, and lexical path safety are done. JSON Schema embedding is pending.
- 🟡 CLI commands: `validate`, `inspect`, and `version` are done. `plan` and `apply` are pending.
- 🟡 Actions: action-specific input validation is done for `replace_text`, `replace_lines`, `insert_before_lines`, `insert_after_lines`, `create`, and `delete`. Runtime execution is pending.
- ⬜ Three-phase commit: validate → stage → commit.
- ✅ Path safety: lexical validation for traversal, absolute paths, and sensitive path blocklist is done. Physical root/symlink validation is pending for `plan`/`apply`.
- ⬜ Unified diff output.
- 🟡 Tests: schema and safety unit tests are done. Integration test against a fixture project is pending.
- ⬜ First release: `overpatch-{linux,darwin,windows}-{amd64,arm64}` binaries.

## v0.2: Audit and polish

- ⬜ `.overpatch/runs/<timestamp>/` log directory.
- ⬜ Persist input JSON, generated diff, execution report, and exit code for every run.
- ⬜ Structured error output with `--output=json`.
- ⬜ Color-aware terminal output and `--no-color` flag.
- ⬜ Better diff rendering, including optional side-by-side view.
- ⬜ `overpatch report <run-id>` to inspect a previous run summary.

## v0.3: Git integration

- ⬜ Detect whether the current directory is inside a Git repository.
- ⬜ Refuse to apply on a dirty working tree by default.
- ⬜ `--force-dirty` escape hatch.
- ⬜ Suggest `git diff` after apply.
- ⬜ `overpatch rollback` using Git when available.
- ⬜ Preserve enough run metadata to explain what was applied.

## v0.4: Schema v1.1

- ⬜ `expected_match_count` as a clearer successor to `expected_occurrences` for line-based operations.
- ⬜ `hint_line` for disambiguation when content alone is not unique.
- ⬜ Optional `pre_hash` file integrity check.
- ⬜ Optional `rationale` per operation, surfaced in audit logs.
- ⬜ Backward compatibility with v1 documents where reasonable.

## v0.5: Context Builder

- ⬜ `overpatch dump` generates a context pack.
- ⬜ Respect `.gitignore`.
- ⬜ Configurable blocklist such as `node_modules`, `dist`, `build`, and secrets.
- ⬜ Per-file size cap and total context pack size cap.
- ⬜ Binary detection using safe heuristics.
- ⬜ Secret detection integration or built-in lightweight checks.
- ⬜ Optional line numbering in dumps for future `hint_line` support.

## v0.6: Provider Gateway

- ⬜ Provider interface with adapters:
  - `manual`: write context to file and expect JSON back.
  - `anthropic`: Claude API.
  - `openai`: OpenAI API.
  - `ollama`: local HTTP provider.
- ⬜ `overpatch plan --prompt "..." --provider=<provider>`.
- ⬜ Retry loop using structured validation/staging errors as feedback to the model.
- ⬜ Provider configuration file or environment variable support.

## v0.7: Smart context

- ⬜ Embeddings-based file ranking, for example with Ollama and `nomic-embed-text`.
- ⬜ Tree-sitter signatures for files not fully included in the context pack.
- ⬜ `overpatch dump --smart "<prompt>"` to select a relevant subset.
- ⬜ Context budget reporting.

## v1.0: Stable toolkit

- ⬜ Schema v1 frozen.
- ⬜ Stable executor behavior for all v1 actions.
- ⬜ Cross-platform release binaries.
- ⬜ Distribution through GitHub Releases.
- ⬜ Optional Homebrew tap and Scoop bucket.
- ⬜ Full documentation.
- ⬜ Real-world example projects under `examples/projects/`.
