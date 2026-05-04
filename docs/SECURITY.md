# Security

## Threat model

Overpatch executes operations described by an LLM. The LLM is **untrusted**. Every operation must be validated as if it came from an attacker.

### What we defend against

- **Path traversal** — `../../../etc/passwd`, symlinks pointing outside the project, absolute paths.
- **Sensitive file overwrite** — `.git/`, `.env`, SSH keys, credentials.
- **Silent partial application** — operation #3 fails after #1 and #2 wrote successfully.
- **Stale dump attacks** — file changed between dump and apply; operation now matches the wrong content.
- **Confused deputy** — LLM convinced by prompt injection in source files to perform unauthorized changes.

### What we do not defend against

- **Malicious source code that the user asks to apply.** If the user explicitly approves a diff, that's their call. Overpatch shows the diff before commit.
- **Compromised provider.** If the LLM API is hijacked, that's a provider trust problem.
- **Local privilege escalation.** Overpatch runs as the user. It cannot do anything the user couldn't do manually.

## Safeguards

### Path validation

Every `path` field passes through:

1. Reject if absolute.
2. Reject if contains `..` after `filepath.Clean`.
3. Resolve to canonical path. Reject if outside project root.
4. Reject if matches the blocklist (`.git/`, `.env`, `.ssh/`, configurable).

### Occurrence counts

Every text-matching operation requires `expected_occurrences`. The executor counts actual occurrences. Mismatch → operation fails → entire run aborts.

### Atomicity

Operations are staged in memory before any disk write. If staging fails for any operation, no writes occur.

Current implementation notes:
- `create` and `modify` operations write via temp file + `os.Rename`, which is atomic at the individual-file level.
- `delete` is basic and has no backup. A deleted file cannot be recovered by Overpatch itself.
- Multi-file apply is not fully atomic. If a write fails after earlier files have already been written, the working tree may be left partially updated. Git is the recommended recovery path in that case, though Overpatch does not enforce its presence.

### Git as safety net (planned — v0.3)

Git integration is planned for v0.3 and is not implemented yet.

In v0.1, `apply` does not check Git status. There is no dirty-tree guard and no `--force-dirty` flag. If a write fails mid-batch, partial changes may remain on disk. Manual recovery via `git checkout -- .` is possible when the project is tracked by Git, but Overpatch does not enforce or verify this.

Until Git integration lands, users should run Overpatch inside a Git repository and review the `plan` output before running `apply`.

### Audit trail (planned — v0.2)

Run logs are planned for v0.2 and are not written yet. In v0.1, no `.overpatch/runs/` directory is created and no run is persisted. Output goes to stdout/stderr only.

## Reporting vulnerabilities

Until the project has a stable maintainer email, please open a private GitHub Security Advisory.
