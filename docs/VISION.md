# Vision

Overpatch is an **AI-assisted deterministic patching toolkit for codebases**.

## What Overpatch is

A pipeline that turns a natural-language change request into a verified, auditable, atomic modification of a local project.

The pipeline has clear stages, each a separable module:

- **Context** — assemble what the AI needs to see
- **Provider** — get a structured plan back from some AI
- **Protocol** — a strict JSON contract for that plan
- **Validation** — schema, safety, business rules
- **Planning** — stage changes and generate a diff
- **Apply** — commit atomically
- **Audit** — log every run for replay and review

## What Overpatch is not

- **Not an editor.** It does not replace your IDE. It complements it.
- **Not an autonomous agent.** It does not loop on its own. The human approves each apply.
- **Not a code generator.** Generation is delegated to whichever LLM provider you plug in.
- **Not a diff tool.** It produces diffs as a side effect, but its job is the safe application of structured intent.
- **Not coupled to a specific LLM.** The provider is pluggable: cloud API, local model, manual browser, mock.

## Design tenets

**Determinism over cleverness.** Two runs of the same JSON against the same project produce identical results. No hidden state, no time-dependent behavior, no network calls during apply.

**Explicit over implicit.** Every operation declares what it expects to find (`expected_occurrences`, `find_lines`, `path`). Mismatch is an error, not a guess.

**Atomic by default.** All operations succeed, or none are written. There is no `--partial` flag and there will not be one.

**Reversible (planned).** Overpatch is designed to integrate with Git as a safety net. Dirty working tree refusal and built-in rollback are planned for v0.3. Until then, reversibility depends on the user's own Git workflow — run Overpatch inside a tracked repository and review the diff before applying.

**Observable (planned).** Every run will leave a trail in `.overpatch/runs/<timestamp>/` with the input, the diff, the report, and the exit status. Run logs are planned for v0.2 and are not written yet.

**Pluggable at the edges, rigid in the middle.** The provider and the context builder can vary. The protocol and the executor cannot drift without a schema version bump.

## Non-goals

- Real-time collaboration.
- Cross-project transactions (each apply targets one repo).
- Editing binary files.
- Replacing `git`. Overpatch sits next to git, not over it.
