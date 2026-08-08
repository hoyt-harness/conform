# Research: Format Dispatch

**Phase 0 output** | **Date**: 2026-08-08

No NEEDS CLARIFICATION markers were present in the spec. This document records the
technical decisions made during planning, the alternatives considered, and the rationale
for each choice.

---

## Decision 1: CHANGED/OK detection

**Decision**: Read file bytes before invoking the formatter; read again after; compare
SHA-256 digests. If digests differ → `CHANGED`. If equal → `OK`.

**Rationale**: Most formatters in the dispatch table (`ruff format`, `gofmt`, `prettier`,
`shfmt`, etc.) write in-place and do not stream to stdout. Digest comparison is the only
consistent mechanism that works across all of them without special-casing each formatter.
It is also accurate: two files with identical content have identical digests regardless
of how the formatter was invoked.

**Alternatives considered**:
- *Use each formatter's own `--check`/`--diff` mode*: Inconsistent — `gofmt -l`, `ruff
  format --check`, `prettier --check`, and `shfmt --diff` all exist but produce
  different output formats and exit codes. Would require per-formatter logic and make
  CHANGED/OK detection depend on formatter flag stability. Rejected.
- *Compare file mtime before and after*: Unreliable — some formatters touch mtime even
  when content is unchanged. Rejected.
- *Diff stdout against original*: Not applicable to in-place formatters. Rejected.

---

## Decision 2: `--check` mode implementation

**Decision**: In check mode, read the file into a buffer, write the buffer to a temp
file in the same directory, invoke the formatter on the temp file, compare the temp
file content against the original buffer. If different → report `CHANGED` (without
writing to the original). Remove the temp file regardless. Exit non-zero if any file
would change.

**Rationale**: This is the only approach that works uniformly across in-place formatters
without requiring formatter-specific flags. The temp file lives in the same directory as
the original so relative path references inside formatters (e.g., a config file search
from the working directory) resolve the same way they would in normal mode.

**Alternatives considered**:
- *Use formatter `--check` flag per formatter*: Same fragmentation problem as Decision 1.
  Rejected for the same reason.
- *Write to stdout-mode flags where available*: `gofmt` (without `-w`) writes to stdout;
  others don't. Too inconsistent. Rejected.

---

## Decision 3: `goimports` vs `gofmt` relationship

**Decision**: Run `goimports -w` as the primary Go formatter. If `goimports` is not
found, fall back to `gofmt -w`. Both are checked via `exec.LookPath` at invocation time
(not at startup). If neither is found, emit `ERROR`.

**Rationale**: `goimports` is a strict superset of `gofmt` — it formats identically and
additionally manages import blocks. For Positronikal's Go projects, import organization
is required. The fallback preserves formatting behavior on systems where `goimports` is
not yet installed.

**Alternatives considered**:
- *Require `goimports`, no fallback*: Unnecessarily hard for users with only `gofmt`
  installed. Rejected.
- *Always run both*: Running `gofmt` after `goimports` is redundant (goimports already
  calls gofmt internally). Rejected.

---

## Decision 4: Shebang detection

**Decision**: For files with no recognized extension (or no extension at all), read the
first line and match the pattern `#!.*/(python3?|sh|bash|pwsh|powershell)(\s|$)` and
the common `#!/usr/bin/env (python3?|sh|bash|pwsh)` form. Map matched interpreter to
the same type as the corresponding extension.

**Rationale**: Git hooks and many shell scripts have no extension. The PostToolUse hook
fires on them as written (path is the file path from the Edit/Write tool). Shebang
detection is the standard UNIX approach and is reliable for the file types in scope.

**Scope**: Only shebangs for formatters in the dispatch table are recognized. Unknown
shebangs fall through to `SKIP`.

**Alternatives considered**:
- *Skip extensionless files entirely (always SKIP)*: Would miss shell hooks and scripts
  that are a common output of the AI agent. Rejected.
- *Use file magic bytes*: Overly complex for text source files; magic bytes distinguish
  binary formats, not scripting language dialects. Rejected.

---

## Decision 5: PowerShell `Invoke-Formatter` invocation

**Decision**: Invoke as `pwsh -NoProfile -Command "Invoke-Formatter -Path '<abs_path>'"`.
Require the `PSScriptAnalyzer` module (which provides `Invoke-Formatter`) to be
pre-installed by the user. If `pwsh` is not found, emit `ERROR`. If
`Invoke-Formatter` is unavailable (module not installed), the subprocess will exit
non-zero and the error message will appear in the `ERROR` line.

**Rationale**: `Invoke-Formatter` is the correct PSScriptAnalyzer formatter. `-NoProfile`
avoids loading user profile scripts that might interfere. The module dependency is
documented in the README; it cannot be bundled.

**Note**: On Windows in Git Bash, `pwsh` resolves to PowerShell 7 (`pwsh.exe`). On
Linux/macOS, `pwsh` resolves to the installed PowerShell Core binary. Both are
acceptable.

---

## Decision 6: `clang-format` vs `uncrustify` for C

**Decision**: Use `clang-format -i` as the sole C formatter. `uncrustify` is not
included in the initial dispatch table.

**Rationale**: `clang-format` is installed as part of LLVM/Clang toolchains, which are
more widely available than `uncrustify`. It ships by default with most C development
environments and is the formatter referenced in the CODING_BIBLE. `uncrustify` can be
added in a later minor version if a concrete need arises.

**Alternatives considered**:
- *Support both with fallback*: Would require two LookPath checks per file and
  decision logic about which to prefer. Adds complexity for a formatter type that
  is currently unused in active Positronikal repos. Rejected.

---

## Decision 7: `ruff check --fix` scope

**Decision**: Run `ruff format` first (formatting pass), then `ruff check --fix --select
I` (import sorting only). The `--select I` flag limits the auto-fix to the `isort` rule
group, avoiding the risk that `ruff check --fix` silently changes semantics by
auto-fixing logic errors.

**Rationale**: `ruff format` handles whitespace/indentation. `ruff check --fix` without
filtering could auto-fix issues like unused imports (F401), which is a semantic change,
not a formatting change. `--select I` (isort rules) is the safe, formatting-only subset.

**Alternatives considered**:
- *Run `ruff check --fix` without `--select I`*: Risks silently removing imports the AI
  agent just added. Rejected.
- *Skip `ruff check` entirely, format only*: Loses import sorting, which is required by
  PCS. Rejected.
