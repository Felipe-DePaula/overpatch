# Overpatch

**AI-assisted deterministic patching toolkit for codebases.**

Overpatch turns natural-language change requests into structured, validated, auditable patch operations. The AI proposes. JSON formalizes. Overpatch validates, previews, and applies.

---

## Why

Letting an LLM edit your files directly is fast and unsafe. Letting it describe the edit in a strict, validated format and applying that description deterministically is fast and safe.

Overpatch is the second path.

## How it works

```
┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  Context     │ → │   Provider   │ → │   Protocol   │ → │   Executor   │
│  (dump)      │   │   (LLM)      │   │   (JSON)     │   │   (apply)    │
└──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘
```

1. **Context** — Overpatch builds a context pack of the project (filtered, capped, secret-aware).
2. **Provider** — Some LLM (cloud API, local Ollama, or a manual browser flow) reads the context and returns a JSON document describing the changes.
3. **Protocol** — The JSON conforms to the Overpatch schema: a list of typed operations with explicit anchors and expected match counts.
4. **Executor** — Overpatch validates the JSON, stages every change in memory, generates a unified diff, and only then commits to disk — atomically.

If anything fails during validation or staging — schema invalid, anchor not found, path unsafe, occurrence count mismatched — nothing is written. A failure during the commit phase (disk full, permissions) may leave the working tree partially updated; Git is the recommended recovery path.

## Status

**v0.1 in development.** Schema is not stable yet. The CLI executor is the first deliverable; context builder and provider integrations come next.

See [`docs/ROADMAP.md`](docs/ROADMAP.md).

## The pieces

| Module                | Role                                                       | Status      |
| --------------------- | ---------------------------------------------------------- | ----------- |
| Overpatch Protocol    | JSON schema for patch operations                           | v1 draft    |
| Overpatch Executor    | CLI that validates, plans, and applies operations          | in progress |
| Overpatch Context     | Dump generator with filters, secret detection, size caps   | planned     |
| Overpatch Provider    | Pluggable bridge to LLMs (API, local, manual)              | planned     |
| Overpatch Guard       | Safety layer (path traversal, sensitive files, git checks) | planned     |
| Overpatch Runs        | Audit log of every plan/apply in `.overpatch/runs/`        | planned     |

## Quick start

Not yet. The executor is being built. Once available:

```bash
overpatch validate ops.json     # schema + safety check
overpatch plan ops.json         # show diff without writing
overpatch apply ops.json        # apply atomically
```

## Development

To build and test Overpatch locally, Go must be installed and available in your `PATH`.

See [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) for setup and build notes.

## Documentation

- [`docs/VISION.md`](docs/VISION.md) — what Overpatch is and is not
- [`docs/PIPELINE.md`](docs/PIPELINE.md) — end-to-end flow and responsibility boundaries
- [`docs/PRODUCER_CONTRACT.md`](docs/PRODUCER_CONTRACT.md) — obligations for humans/LLMs/agents producing Overpatch JSON
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — pipeline and module boundaries
- [`docs/PROTOCOL.md`](docs/PROTOCOL.md) — JSON schema specification
- [`docs/ROADMAP.md`](docs/ROADMAP.md) — versioned milestones
- [`docs/SECURITY.md`](docs/SECURITY.md) — threat model and safeguards
- [`docs/GLOSSARY.md`](docs/GLOSSARY.md) — terminology
- [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) — local development and build notes

## Experiments

[`experiments/browser-provider/`](experiments/browser-provider/) — bookmarklets that inject the planning prompt into a chat UI and extract the JSON response. Used as a manual provider during early development.

## License

MIT. See [`LICENSE`](LICENSE).
