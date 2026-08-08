<!--
Sync Impact Report
Version change: (template) → 1.0.0
Added sections: Core Principles (I–V), Distribution, Governance
Removed sections: none (initial fill from template)
Follow-up TODOs: none
-->

# conform Constitution

## Core Principles

### I. Dispatcher, Not Formatter

`conform` routes files to external formatters; it never implements formatting itself. Every formatting
rule lives in a dedicated tool (ruff, gofmt, prettier, etc.). `conform` knows only how to detect a
file type, invoke the right formatter for that type, and report the result. Adding a new language
MUST mean adding an entry to the dispatch map, not writing formatting logic.

**Rationale**: Formatting rules change constantly and vary by ecosystem. Owning them here creates
maintenance burden and divergence from the authoritative formatter. The dispatcher model keeps
`conform` stable as formatters evolve.

### II. CLI Contract

The primary interface is `conform <filepath> [<filepath>...]`. One status line per file to stdout:

```
CHANGED path/to/file.py  3 lines reformatted
OK     path/to/other.go  no changes
SKIP   path/to/file.xyz  unknown type
ERROR  path/to/file.sh   shfmt not installed
```

Exit 0 in all normal cases (including SKIP and ERROR). Exit 1 on internal error only.
`--check` flag: dry-run mode, exit 1 if any file would change. Errors go to stderr.
The output contract is stable: downstream consumers (hooks, CI) may parse it.

**Rationale**: Claude Code's PostToolUse hook reads stdout. The contract must be machine-parseable,
stable, and non-zero-exit-on-absence so hook failures do not interrupt the AI session.

### III. Test-First (NON-NEGOTIABLE)

Tests are written before implementation. No implementation code is written until:
1. Tests are written that describe the expected behavior
2. Tests are confirmed to fail (red phase)
3. Implementation is written to make them pass (green phase)

Tests MUST cover: type detection (extension and shebang), formatter dispatch, output format
for all four status codes, `--check` exit code semantics, and graceful degradation when a
formatter is absent.

**Rationale**: The output contract and degradation behavior are the product. Tests that encode
them before implementation ensure the contract is correct by definition, not by accident.

### IV. Graceful Degradation

A missing formatter MUST NOT cause a non-zero exit or abort processing of other files.
`ERROR path/to/file.sh   shfmt not installed` is the correct output; terminating is not.
This applies equally to formatter crashes, permission errors, and unsupported file types (SKIP).
The tool reports and continues.

**Rationale**: The hook fires on every file write across many languages. If one formatter is absent,
all other files in the session must still be processed. Hard failures break the editing flow for
unrelated languages.

### V. Zero Setup

No configuration file. No installer subcommand. The formatter map is hard-coded by convention:
file extension → formatter. Installing `conform` means placing the binary on PATH. Installing a
formatter means placing it on PATH. No `.conform.toml`, no `[tool.conform]`, no `conform install`.

**Rationale**: Configuration adds a decision surface that defeats the "zero cognitive overhead"
goal. The whole point is that formatting just happens. Conventions are enforced by hard-coding;
if a project needs a different formatter for a type, it chooses a different tool.

## Distribution

`conform` ships as a single static Go binary with no runtime dependencies. Build from source:
`go build -o conform .`. Install to `~/.local/bin/conform` (already on PATH in the Positronikal
environment). Formatters are separate installs; `conform` discovers them via PATH at runtime.

The binary targets general use beyond the Positronikal environment. No Positronikal-specific
paths, branding, or configuration may be hard-coded into the binary or its documentation.

## Formatter Compatibility

The initial formatter map:

| Language / type       | Formatter                         | Detection              |
|-----------------------|-----------------------------------|------------------------|
| Python                | `ruff format` + `ruff check --fix`| `.py` / shebang        |
| Go                    | `gofmt` + `goimports`             | `.go`                  |
| Shell                 | `shfmt`                           | `.sh`, `.bash` / shebang |
| PowerShell            | `pwsh -Command Invoke-Formatter`  | `.ps1`, `.psm1`        |
| JSON / YAML / MD / CSS| `prettier`                        | extensions             |
| JavaScript/TypeScript | `prettier`                        | `.js`, `.ts`, `.jsx`, `.tsx` |
| C / C header          | `clang-format`                    | `.c`, `.h`             |

Adding a formatter: add a row to the map, add detection logic, add tests. No other changes.
Removing a formatter: mark SKIP for that type; never hard-fail on absence.

## Governance

This constitution is the primary design authority for `conform`. It supersedes preferences,
habits, and informal conventions. Amendments require:
1. A documented rationale (what changed and why)
2. A version increment following semantic versioning:
   - MAJOR: removal or incompatible redefinition of a principle
   - MINOR: new principle or materially expanded guidance
   - PATCH: clarification, wording, non-semantic refinement
3. Update to the Sync Impact Report comment at the top of this file

Compliance is verified at each spec-kit phase gate (spec → plan → tasks → implement).
Any implementation decision that conflicts with a principle MUST trigger a constitution amendment
or be rejected — not worked around.

**Version**: 1.0.0 | **Ratified**: 2026-08-08 | **Last Amended**: 2026-08-08
