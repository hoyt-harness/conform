# CLI Contract: conform

**Phase 1 output** | **Date**: 2026-08-08

This document is the stable external contract for the `conform` binary. Downstream
consumers (hooks, CI scripts, documentation) depend on this contract. Changes to any
element below require a major version increment.

---

## Invocation

```
conform [--check] <file> [<file>...]
```

| Element | Description |
|---|---|
| `--check` | Dry-run mode: report what would change, modify nothing, exit non-zero if any file would be reformatted |
| `<file>` | One or more file paths to process; at least one is required |

---

## Standard Output

One line per input file, always. Format:

```
STATUS path/to/file  message
```

| Field | Width | Content |
|---|---|---|
| `STATUS` | 7 chars, left-aligned | One of: `CHANGED`, `OK     `, `SKIP   `, `ERROR  ` |
| `path/to/file` | variable | The exact path string passed as argument |
| `message` | variable | Human-readable detail; content may change across versions |

**Status codes**:

| Code | Meaning | Occurs when |
|---|---|---|
| `CHANGED` | File was reformatted | Formatter ran and modified file content |
| `OK` | No changes needed | Formatter ran and content was already correct |
| `SKIP` | Type not recognized | Extension and shebang both unrecognized |
| `ERROR` | Could not format | Formatter absent, crashed, or file unreadable |

**Example output**:
```
CHANGED src/main.py      4 lines reformatted
OK      internal/run.go  no changes
SKIP    .editorconfig    unknown type
ERROR   hooks/pre-push   shfmt not installed
```

---

## Standard Error

Diagnostic output not associated with a specific file (e.g., argument parsing errors,
internal failures). Per-file errors are reported on stdout as `ERROR` lines, not on
stderr.

---

## Exit Codes

| Code | Condition |
|---|---|
| `0` | Normal completion (including SKIP and ERROR lines) |
| `1` | `--check` mode and at least one file would be reformatted |
| `2` | Internal failure (invalid arguments, unreadable input before processing begins) |

---

## Supported File Types

| Type | Extensions | Shebang tokens |
|---|---|---|
| Python | `.py` | `python`, `python3` |
| Go | `.go` | — |
| Shell | `.sh`, `.bash` | `sh`, `bash` |
| PowerShell | `.ps1`, `.psm1` | `pwsh`, `powershell` |
| JSON | `.json` | — |
| YAML | `.yaml`, `.yml` | — |
| Markdown | `.md` | — |
| CSS | `.css` | — |
| JavaScript | `.js`, `.jsx` | — |
| TypeScript | `.ts`, `.tsx` | — |
| C | `.c`, `.h` | — |

Files not matching any row above produce a `SKIP` status line.

---

## Stability Guarantees

- The four status codes (`CHANGED`, `OK`, `SKIP`, `ERROR`) are stable across all
  minor and patch versions.
- The output line format (STATUS + path + message) is stable across all minor and
  patch versions; field widths are stable.
- The `message` field content is informational and may change in any version.
- Exit codes 0, 1, and 2 are stable across all minor and patch versions.
- Supported file type extensions may be added (new rows) in minor versions.
- Existing extension-to-type mappings are stable across all versions.
