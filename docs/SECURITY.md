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

### Git as safety net (v0.1)

A Git guard is implemented for `apply` in v0.1. Before writing any file to disk, `apply` checks:

1. `git` is available in `PATH`.
2. The current directory is inside a Git repository.
3. The working tree is clean (no staged, unstaged, or untracked changes).

If any check fails, `apply` prints a clear `apply: refused` message with a hint and exits with code 1. No files are written.

`validate`, `inspect`, and `plan` do not require Git and are unaffected by this guard.

Overpatch does not install Git, initialize repositories automatically, or create commits. Recovery from a partial apply (disk full, permissions error mid-batch) relies on manual `git checkout -- .` or `git restore .`.

`--force-dirty` (to bypass the clean-tree check) and `overpatch rollback` are still planned for a future version and are not implemented in v0.1.

### Audit trail (planned — v0.2)

Run logs are planned for v0.2 and are not written yet. In v0.1, no `.overpatch/runs/` directory is created and no run is persisted. Output goes to stdout/stderr only.

## Reporting vulnerabilities

Until the project has a stable maintainer email, please open a private GitHub Security Advisory.
