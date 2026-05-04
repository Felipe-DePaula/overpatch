# Pipeline

> See also: [`docs/PRODUCER_CONTRACT.md`](PRODUCER_CONTRACT.md) for producer obligations.

## Overview

Overpatch mediates between a *producer* (who proposes changes) and a local codebase (which receives them). The full pipeline has seven stages:

```
User request
  ↓
Context discovery
  (read relevant files, understand project structure)
  ↓
Producer
  (human, LLM, agent, or provider)
  ↓
Overpatch JSON document
  ↓
overpatch validate
  ↓
overpatch plan / inspect
  ↓
overpatch apply --yes
```

Each stage has a clear responsibility. No stage does the work of another.

### Stage breakdown

**1. User request**
A natural-language instruction: "Disable the login route", "Add rate limiting to the API handler", "Rename the `Config` struct to `Settings`".

**2. Context discovery**
Whoever produces the document must first understand the project. This means reading the relevant files, locating the code to change, and confirming the exact text that anchors each operation. In the future, `overpatch dump` will automate this; today it is a manual or tool-assisted step.

**3. Producer**
A human, LLM, agent, or future provider gateway that consumes the context and emits an Overpatch JSON document. The producer bears full responsibility for having read the right files and for the correctness of every anchor. See [`docs/PRODUCER_CONTRACT.md`](PRODUCER_CONTRACT.md).

**4. Overpatch JSON document**
The formal, schema-validated description of the changes. A document is either `success` (with a non-empty `operations` array), `no_changes`, or `failed`. The document is the handoff point — everything before it is the producer's domain; everything after it is the executor's domain.

**5. Validate**
`overpatch validate` checks the document against the schema, verifies business rules (unique IDs, non-empty operations for `success`, etc.), and checks path safety (no traversal, no sensitive paths). This phase reads no target files. It is cheap and safe to run first.

**6. Inspect / Plan**
`overpatch inspect` shows the parsed document. `overpatch plan` loads the target files, applies every operation against in-memory copies, and prints a unified diff. Nothing is written to disk. If any operation fails (anchor not found, `expected_occurrences` mismatch, file missing), the plan aborts with a structured error.

**7. Apply**
`overpatch apply --yes` runs the full three-phase commit: validate → stage → commit. Changes are written to disk only if staging succeeds for every operation.

## Current responsibility boundary (v0.1)

In v0.1, the Overpatch CLI starts at the JSON document.

The CLI:
- validates, plans, and applies explicit operations
- does **not** build context packs
- does **not** call any LLM or provider
- does **not** infer which files should change
- does **not** interpret natural language
- does **not** execute globs, directory paths, or recursive operations

Stages 1–3 (user request, context discovery, producer) happen entirely outside the CLI, via human judgment, manual LLM use, or external tooling.

## Producer responsibility

The producer is responsible for everything that happens before the JSON document. See [`docs/PRODUCER_CONTRACT.md`](PRODUCER_CONTRACT.md) for the full contract.

Short summary: the producer must read the right files, use exact file paths, include accurate `expected_occurrences`, and return `failed` when context is insufficient.

## Executor responsibility

The executor (validator + planner + writer) treats the incoming document as untrusted. It:

- validates schema and business rules before touching any file
- checks path safety (no traversal, no sensitive paths)
- stages all changes in memory before writing anything
- generates and shows a diff before committing
- writes to disk only on an explicit `apply --yes` command
- aborts the entire batch if any operation fails during staging

The executor does not discover files, does not guess anchors, and does not retry on its own.

## Future pipeline work

The following stages are planned but not yet implemented:

| Stage | Planned command | Milestone |
| --- | --- | --- |
| Context builder | `overpatch dump` | v0.5 |
| Provider gateway | `overpatch plan --prompt "..." --provider=<x>` | v0.6 |
| Smart context | `overpatch dump --smart "<prompt>"` | v0.7 |
| Retry loop on errors | automatic, using structured error feedback | v0.6 |

Until these are available, the producer role is filled by a human or an external LLM session.
