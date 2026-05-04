# Architecture

## Pipeline

```
┌─────────────┐
│   User      │  "Disable the login route temporarily"
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Context    │  Builds a dump of the relevant project files
│  Builder    │  (filtered, capped, secret-scrubbed)
└──────┬──────┘
       │   dump.txt + prompt
       ▼
┌─────────────┐
│  Provider   │  Sends to an LLM (API, local, or manual browser)
│  Gateway    │  Receives the response
└──────┬──────┘
       │   raw response
       ▼
┌─────────────┐
│  Protocol   │  Parses and validates the AI_FINAL_OUTPUT envelope
│  Parser     │  Extracts the JSON document
└──────┬──────┘
       │   *schema.Document
       ▼
┌─────────────┐
│  Validator  │  Schema validation + business rules + safety
└──────┬──────┘
       │   validated document
       ▼
┌─────────────┐
│  Planner    │  Loads target files, stages all operations in memory
│             │  Generates unified diff
└──────┬──────┘
       │   stage map + diff
       ▼
┌─────────────┐
│  Executor   │  Confirms with user, then writes to disk
│             │  All-or-nothing
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Run Log    │  Persists input, diff, report, exit code
│             │  to .overpatch/runs/<timestamp>/
└─────────────┘
```

Each box is a Go package in `internal/`. Boundaries are enforced by the import graph.

## Module map

| Pipeline stage   | Go package           | Responsibility                              |
| ---------------- | -------------------- | ------------------------------------------- |
| Context Builder  | `internal/context`   | Walks the FS, filters, caps, dumps          |
| Provider Gateway | `internal/provider`  | Adapters: cloud, local, manual              |
| Protocol Parser  | `internal/schema`    | Structs, JSON Schema embed, parse           |
| Validator        | `internal/validator` | Schema rules + business rules               |
| Safety Layer     | `internal/safety`    | Path traversal, sensitive paths, git status |
| Operations       | `internal/ops`       | One file per action                         |
| Planner          | `internal/planner`   | Stage + diff                                |
| Executor         | `internal/executor`  | Commit phase                                |
| Diff             | `internal/diff`      | Unified diff rendering                      |
| Reporter         | `internal/report`    | Per-operation results                       |
| Run log          | `internal/runlog`    | `.overpatch/runs/` writer                   |
| CLI              | `internal/cli`       | Cobra commands, wiring, exit codes          |

## Three-phase commit

Every `apply` proceeds in three strictly ordered phases:

1. **Validate** — pure functions. Reads the JSON. Checks schema, paths, business rules. No filesystem reads of target files. Cheap. May reject the document.

2. **Stage** — reads target files. Applies every operation against in-memory copies. Builds a `map[path][]byte` of post-change contents and a set of paths to delete. If any operation fails (anchor not found, occurrence mismatch, etc.), the stage aborts and nothing has been written.

3. **Commit** — writes the staged map to disk. Each file write uses `os.WriteFile`, which is atomic at the file level on POSIX. If a write fails mid-batch (extremely rare: disk full, permissions), the run is marked failed and Git is the recovery path.

`plan` runs phases 1 and 2, prints the diff, and stops. `apply` runs all three.

## Atomicity guarantees

- **Within a single file:** atomic via `os.WriteFile` semantics.
- **Across multiple files:** best-effort. We write fast and sequentially after staging is complete. A crash between writes leaves the working tree partially updated — recoverable via Git (`git checkout -- .`).
- **Reinforcement:** Overpatch refuses to `apply` on a dirty working tree unless `--force-dirty` is passed. This makes Git the safety net.

A future version may stage to a temp directory and use `os.Rename` for stronger guarantees. Not worth the complexity for v0.1.

## Decisions

This section is append-only. New decisions go at the bottom with a date.

### 2026-05 — Anchors over line numbers

Operations match by content (`find`, `find_lines`), not by line numbers. Line numbers may appear as hints but are never authoritative. Rationale: LLMs miscount lines frequently; content matching is robust against the dump being slightly stale.

### 2026-05 — JSON for v1, hybrid format reserved for v2

The protocol is JSON for v1. Multiline strings with escaped newlines are clunky for LLMs but the toolchain (validation, parsing, schema) is mature. A future v2 may introduce a hybrid envelope-plus-textual-blocks format (à la OpenAI `apply_patch`) once the JSON path is solid.

### 2026-05 — Go for the executor

Single-binary distribution, cross-platform builds, mature CLI ecosystem (cobra), no runtime dependency for users. Python was considered for prototyping speed but rejected because the protocol design is already concrete enough to implement directly.
