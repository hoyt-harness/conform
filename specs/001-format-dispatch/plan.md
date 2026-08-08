# Implementation Plan: Format Dispatch

**Branch**: `001-format-dispatch` | **Date**: 2026-08-08 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-format-dispatch/spec.md`

## Summary

`conform` is a single static CLI binary that dispatches source files to external
formatters by file type, rewrites files in place, and reports one status line per file.
It integrates with AI coding agents via a PostToolUse hook and exposes a `--check` flag
for CI pipeline use. The implementation is pure Go standard library with `os/exec` for
subprocess invocation — no external Go dependencies.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**: Standard library only (`os/exec`, `bufio`, `flag`, `path/filepath`,
`crypto/sha256`, `io`, `os`)

**Storage**: N/A — no persistent state; operates on files passed as arguments

**Testing**: `go test ./...` (standard Go test runner)

**Target Platform**: Linux, macOS, Windows (cross-compiled static binary; primary
development and initial deployment on Windows via Git Bash / MSYS2)

**Project Type**: CLI binary

**Performance Goals**: Total wall-clock time per file < 2 seconds (formatter execution
dominates; conform's own overhead target < 50ms excluding subprocess time)

**Constraints**: Single static binary with no runtime dependencies beyond external
formatters; zero configuration files; installable by placing binary on PATH

**Scale/Scope**: Typical invocation is 1 file (PostToolUse hook); batch invocation of
up to ~100 files in CI check mode; no concurrency requirements in v1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Gate | Status | Notes |
|---|---|---|---|
| I. Dispatcher, Not Formatter | `conform` MUST NOT implement formatting logic | ✓ PASS | All formatting delegated to external binaries via `os/exec` |
| II. CLI Contract | CHANGED/OK/SKIP/ERROR output; exit 0 always; `--check` exits non-zero on diff | ✓ PASS | Output contract fully specified in contracts/cli.md |
| III. Test-First | Tests written and failing before implementation code exists | ✓ PASS (gate) | Enforced in tasks.md — no source files created until tests exist |
| IV. Graceful Degradation | Missing/crashing formatter → ERROR line, exit 0 | ✓ PASS | exec.LookPath failure → ERROR; subprocess non-zero → ERROR |
| V. Zero Setup | No config file; hard-coded dispatch map; no installer | ✓ PASS | Dispatch table is a Go map literal in dispatch.go |
| Distribution | Single static binary; no Positronikal-specific content | ✓ PASS | `CGO_ENABLED=0 go build`; README written for general audience |

**Simplicity Gate (Article VII)**: 1 project, no future-proofing abstraction layers. ✓
**Anti-Abstraction Gate**: Using `os/exec` directly; no wrapper framework. ✓
**Test-First Gate**: No implementation before test files exist and fail. ✓ (enforced in tasks)

*Post-Phase-1 re-check:* All gates hold after design. The contracts and data model
introduce no new complexity beyond what the constitution permits.

## Project Structure

### Documentation (this feature)

```text
specs/001-format-dispatch/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output
├── quickstart.md        ← Phase 1 output
├── checklists/
│   └── requirements.md  ← spec validation checklist
├── contracts/
│   └── cli.md           ← Phase 1 output
└── tasks.md             ← Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
conform/
├── main.go          ← entry point, argument parsing, output loop
├── detect.go        ← file type detection (extension table + shebang fallback)
├── dispatch.go      ← formatter registry and subprocess invocation
├── detect_test.go   ← tests for type detection
├── dispatch_test.go ← tests for dispatch, CHANGED/OK/ERROR/SKIP output
├── go.mod
└── go.sum
```

**Structure Decision**: Single flat package. The binary has three clear responsibilities
(detect, dispatch, report) mapped to three source files. No subdirectories — the project
is small enough that package splitting would add import complexity with no modularity
benefit (Rule of Parsimony).

## Complexity Tracking

> No constitution violations — section not applicable.
