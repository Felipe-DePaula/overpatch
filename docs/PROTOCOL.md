# Overpatch Protocol v1

> Documents are produced by a producer (human, LLM, agent, or provider). See [`docs/PRODUCER_CONTRACT.md`](PRODUCER_CONTRACT.md) for producer obligations.

## Document envelope

Every Overpatch document is a single JSON object:

```json
{
  "schema_version": "overpatch/v1",
  "status": "success",
  "reason": "",
  "summary": "Short human-readable description",
  "operations": [ ... ]
}
```

### Fields

| Field            | Type     | Required | Notes                                          |
| ---------------- | -------- | -------- | ---------------------------------------------- |
| `schema_version` | string   | yes      | Must be `"overpatch/v1"`                       |
| `status`         | enum     | yes      | `success`, `no_changes`, or `failed`           |
| `reason`         | string   | yes      | Empty for `success`; explanation otherwise     |
| `summary`        | string   | no       | One-line description, useful in logs           |
| `operations`     | array    | yes      | Non-empty iff `status == "success"`            |

### Status semantics

- `success` — the document describes one or more changes to apply. `operations` MUST be non-empty.
- `no_changes` — the AI analyzed the project and concluded no change is needed. `operations` MUST be empty. `reason` explains why.
- `failed` — the AI could not produce a valid plan. `operations` MUST be empty. `reason` explains why.

## Operation envelope

Every operation has these common fields:

| Field    | Type   | Required | Notes                                   |
| -------- | ------ | -------- | --------------------------------------- |
| `id`     | string | yes      | Unique within the document; must match `^op_[A-Za-z0-9_]+$` |
| `action` | enum   | yes      | One of the actions below                |
| `path`   | string | yes      | Relative path from project root         |

Operation `id` format:
- Must start with `op_` (literal prefix).
- Followed by one or more ASCII letters (A-Z, a-z), digits (0-9), or underscores (`_`).
- Examples: `op_001`, `op_replace_login`, `op_A1`, `op_abc_123`.
- Invalid examples: `001` (no prefix), `op-001` (hyphen), `op 001` (space), `op_` (empty suffix).

## Actions (v1)

### `replace_text`

Literal substring replacement.

```json
{
  "id": "op_001",
  "action": "replace_text",
  "path": "docs/README.md",
  "find": "Hello World",
  "replace": "Olá Mundo",
  "occurrence": "all",
  "expected_occurrences": 2
}
```

| Field                  | Type   | Required | Notes                                       |
| ---------------------- | ------ | -------- | ------------------------------------------- |
| `find`                 | string | yes      | Literal text to search for                  |
| `replace`              | string | yes      | Replacement text                            |
| `occurrence`           | enum   | yes      | `"all"` or `"first"`                        |
| `expected_occurrences` | int    | yes      | Total matches expected. Mismatch → error.   |

### `replace_lines`

Replace an exact contiguous block of lines.

```json
{
  "id": "op_002",
  "action": "replace_lines",
  "path": "src/auth.ts",
  "find_lines": [
    "router.post('/login', async (req, res) => {",
    "  return loginHandler(req, res);",
    "});"
  ],
  "replace_lines": [
    "router.post('/login', async (req, res) => {",
    "  res.status(503).json({ error: 'login_disabled' });",
    "});"
  ],
  "expected_occurrences": 1
}
```

The `find_lines` block must appear exactly `expected_occurrences` times. Lines are matched verbatim, including indentation.

### `insert_before_lines` / `insert_after_lines`

Insert lines before or after an anchor block.

```json
{
  "id": "op_003",
  "action": "insert_after_lines",
  "path": "src/main.ts",
  "find_lines": ["import express from 'express';"],
  "insert_lines": ["import { auditLog } from './audit';"],
  "expected_occurrences": 1
}
```

### `create`

Create a new file. Fails if the file exists.

```json
{
  "id": "op_004",
  "action": "create",
  "path": "src/audit.ts",
  "content": "export function auditLog(event: string) {\n  console.log(event);\n}\n"
}
```

### `delete`

Delete an existing file. Fails if the file does not exist.

```json
{
  "id": "op_005",
  "action": "delete",
  "path": "src/legacy.ts"
}
```

## Schema and validation

`schemas/overpatch.v1.schema.json` is the canonical JSON Schema specification for Overpatch v1 documents.

In the current v0.1 development state, Overpatch runtime validation is implemented in Go and does not embed or enforce the JSON Schema. Validation is performed by the Go validator in `internal/schema/validate.go`.

Embedding the JSON Schema at runtime and using it as the authoritative validator is planned for a future release but is not implemented yet.

## Validation rules

1. `schema_version` must be exactly `"overpatch/v1"`.
2. `status == "success"` requires `len(operations) >= 1`.
3. `status` in `{"no_changes", "failed"}` requires `len(operations) == 0` and `reason != ""`.
4. Every `id` must be unique within the document.
5. Every `path` must be relative, must not contain `..`, and must resolve inside the project root.
6. For `expected_occurrences`: actual count in the file must match exactly. No tolerance.
7. `create` fails if path exists. `delete` fails if path does not exist.

## Error feedback (planned for v1.1)

When validation or staging fails, Overpatch emits a structured error suitable for an LLM retry loop:

```json
{
  "error_code": "ANCHOR_NOT_FOUND",
  "operation_id": "op_002",
  "path": "src/auth.ts",
  "message": "find_lines block not found in file",
  "hint": "File has 47 lines; first 3 lines of find_lines did not match any region"
}
```
