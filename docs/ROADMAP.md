# Roadmap

Legend:

- ✅ Done
- 🟡 In progress
- ⬜ Pending

## v0.1: Executor MVP

The first boss. Goal: a working CLI that validates, previews, and applies an Overpatch JSON document to files in a local project.

- 🟡 Project skeleton: Go module, Cobra CLI, Windows manifest, and basic repository structure are done. Makefile and lint configuration are pending.
- ✅ CLI commands: `version`, `validate`, `inspect`, `plan`, and `apply` are done.
- ✅ `validate`: JSON parser, structural validation, action-specific validation, and lexical path safety are done.
- ⬜ JSON Schema embedding.
- ✅ `plan`: implemented for all basic v1 actions: `replace_text`, `replace_lines`, `insert_before_lines`, `insert_after_lines`, `create`, and `delete`. Staging is in memory and does not write to disk.
- 🟡 `apply`: implemented with required `--yes`. It applies `created`, `modified`, and `deleted` file changes. `created` and `modified` use temp file + rename. `delete` is still basic and has no backup yet. Run log, git guard, and rollback are pending.
- 🟡 Actions: action-specific validation and planning are done for all basic v1 actions. Runtime apply covers `created`, `modified`, and `deleted` file changes via `StageResult`. Additional safety and final integration work are pending.
- 🟡 Three-phase commit: validate and stage exist, and the initial commit/apply phase exists. Robustness work remains: delete backups, run log, git guard, and rollback.
- 🟡 Path safety: lexical validation for traversal, absolute paths, and sensitive path blocklist is done. Physical root/symlink validation is pending.
- 🟡 Unified diff output: simple unified diff rendering exists in `internal/diff`. More advanced rendering, color, and optional side-by-side output are pending.
- 🟡 Tests: unit tests for schema, safety, planner, diff, and executor are done. Integration test against a fixture project is pending.
- ⬜ First release: `overpatch-{linux,darwin,windows}-{amd64,arm64}` binaries.

### File and directory scope

In v0.1, operations target files, not directories. Directories are not first-class operation targets.

- `create` creates files and may create missing parent directories as operational support. The target of the operation remains a file.
- `delete` removes files and must fail if the path points to a directory.
- `replace_text`, `replace_lines`, `insert_before_lines`, and `insert_after_lines` modify existing files.
- Directory operations such as `create_dir`, `delete_dir`, `delete_tree`, `move_dir`, and recursive operations are out of scope for v0.1.
- Explicit directory support should be evaluated after v1.0, possibly in v1.1 or v2. If added, prefer safe and explicit operations such as `create_dir`, `delete_empty_dir`, `move_file`, and `rename_file`; avoid recursive `delete_dir` for safety.

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
