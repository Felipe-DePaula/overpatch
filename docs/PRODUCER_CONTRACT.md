# Producer Contract

> See also: [`docs/PIPELINE.md`](PIPELINE.md) for the end-to-end flow and responsibility boundaries.

## What is a producer?

A **producer** is any human, LLM, agent, provider, or tool that generates an Overpatch JSON document. The producer owns everything that happens before the document reaches the executor: reading context, understanding intent, selecting file paths, and writing the operations.

In v0.1, producers are typically humans or external LLM sessions (browser chat, API calls). Future versions will introduce a provider gateway that automates this role.

## Producer obligations

A producer MUST:

1. **Not invent files.** Every `path` in an operation must refer to a file that exists in the project (or, for `create`, a path that should be created). Paths are not checked against the filesystem during `validate` — the mismatch surfaces during `plan` or `apply`, causing a hard failure.

2. **Read the relevant file contents before proposing edits.** For any operation that matches text (`replace_text`, `replace_lines`, `insert_before_lines`, `insert_after_lines`), the producer must have the actual file content in context. Guessing or paraphrasing the anchor text leads to `ANCHOR_NOT_FOUND` errors.

3. **Use explicit, relative file paths.** Paths must be relative to the project root, must not contain `..`, and must resolve to a single specific file.

4. **Not use globs or directory-wide implicit operations.** There is no glob expansion, no recursive directory targeting, and no mass-edit by directory path. Every file that should be touched must appear as its own operation with an explicit `path`.

5. **Include accurate `expected_occurrences` for text- and line-matching operations.** Count how many times the anchor appears in the file. If the count cannot be confirmed from context, return `failed` — do not guess.

6. **Return `status: "failed"` with a clear `reason` when context is insufficient.** If the producer has not read the relevant files, cannot confidently identify anchors, or the request is ambiguous, the correct response is a `failed` document — not a best-guess operation.

7. **Return `status: "no_changes"` when no edit is needed.** If the requested change is already present in the code, return `no_changes` with a reason. Do not produce a no-op operation.

8. **Avoid sensitive paths.** The executor enforces a path blocklist, but the producer should not target files like `.env`, `*.pem`, or `.git/` even where the executor would allow them.

9. **Keep operations minimal and auditable.** Prefer one targeted operation over many speculative ones. Every operation in the document will be reviewed by a human before apply.

10. **Not rely on line numbers as authoritative anchors.** Line numbers shift as files change. Use content-based anchors (`find`, `find_lines`). Line numbers are hints only and are not supported as primary matching in v1 operations.

## Relationship to `expected_occurrences`

`expected_occurrences` is a guardrail, not a proof of correctness.

When the executor finds a different number of anchor matches than `expected_occurrences` declares, it aborts with an error. This catches:

- **Stale context** — the producer read a version of the file that no longer matches the filesystem.
- **Missing files** — the file was deleted or never existed.
- **Unexpected duplicates** — the anchor appears more times than the producer believed.
- **Drift** — the file was modified between context gathering and apply.

A matching `expected_occurrences` confirms the anchor count was correct at apply time. It does not confirm that the producer read the right version, that the surrounding context was understood correctly, or that the operation is semantically safe.

## File and directory scope

In v0.1, all operations target individual files.

- `create` creates a new file. It may create missing parent directories as a side effect, but the operation target is always the file path.
- `delete` removes a file. It must fail if the path resolves to a directory.
- `replace_text`, `replace_lines`, `insert_before_lines`, and `insert_after_lines` modify the contents of an existing file.

Directories are not first-class operation targets. There are no glob patterns, no recursive operations, and no `delete_tree` or `move_dir` actions in v1. A producer that targets `config/` as a path — expecting the executor to find all files under it — will receive a validation error.

## Ambiguity and failure

**If the user's request is ambiguous,** the producer should not guess. Ask the user for clarification. If the producer is an LLM in a non-interactive context, return `failed` with a reason that describes what is ambiguous.

**If required context is missing** (the producer has not seen the relevant files), return `failed`. Do not hallucinate anchors.

**If the change is already present,** return `no_changes`. A `replace_text` that replaces text with itself is a no-op; the executor will not catch it as an error — the producer must detect this before producing the document.

## Examples

### Good producer behavior

User request: "Add a `## Development` section link to the end of README.md."

Producer steps:
1. Read `README.md`.
2. Confirm the section does not already exist → not `no_changes`.
3. Identify a unique anchor near the end of the file.
4. Count occurrences: 1.
5. Produce:

```json
{
  "schema_version": "overpatch/v1",
  "status": "success",
  "summary": "Add Development section link to README",
  "operations": [
    {
      "id": "op_001",
      "action": "replace_text",
      "path": "README.md",
      "find": "MIT. See [`LICENSE`](LICENSE).",
      "replace": "MIT. See [`LICENSE`](LICENSE).\n\n## Development\n\nSee [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md).",
      "occurrence": "first",
      "expected_occurrences": 1
    }
  ]
}
```

### Bad producer behavior

User request: "Update all configs to use the new timeout value."

Bad producer response:

```json
{
  "schema_version": "overpatch/v1",
  "status": "success",
  "operations": [
    {
      "id": "op_001",
      "action": "replace_text",
      "path": "config/",
      "find": "timeout: 30",
      "replace": "timeout: 60",
      "occurrence": "all",
      "expected_occurrences": 3
    }
  ]
}
```

Why this is invalid:
- `path: "config/"` is a directory, not a file. The executor will reject it.
- There are no glob or recursive operations in v1.
- The producer should enumerate each config file individually, confirm the anchor text in each, and produce one operation per file.
- If the producer has not read the config files, it cannot know the anchor text or occurrence count — and must return `failed` instead.
