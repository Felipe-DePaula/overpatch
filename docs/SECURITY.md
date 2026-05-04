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

### Git as safety net

By default, `apply` refuses on a dirty working tree. After apply, `git reset --hard` is always a one-line recovery.

### Audit trail

Every run writes input, diff, and result to `.overpatch/runs/<timestamp>/`. Even successful runs are reviewable.

## Reporting vulnerabilities

Until the project has a stable maintainer email, please open a private GitHub Security Advisory.
