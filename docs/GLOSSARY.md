# Glossary

**Action** — A specific kind of operation: `replace_text`, `create`, etc.

**Anchor** — Text used to locate where an operation applies. Distinct from line numbers.

**Apply** — The third phase of execution: writing staged changes to disk.

**Commit (Overpatch sense)** — The disk-write phase. Distinct from a Git commit.

**Context pack / dump** — A serialized snapshot of the project sent to the LLM.

**Document** — A single Overpatch JSON file, containing one envelope and zero or more operations.

**Envelope** — The outer JSON object: `schema_version`, `status`, `summary`, `operations`.

**Executor** — The component that runs the three-phase commit.

**Operation** — A single change directive within a document.

**Plan** — The phase that stages all changes and produces a diff without writing.

**Provider** — An adapter that talks to an LLM (cloud API, local model, manual browser).

**Stage** — The in-memory representation of post-change file contents.

**Three-phase commit** — Validate → Stage → Commit.
