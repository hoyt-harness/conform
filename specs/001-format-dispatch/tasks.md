# Tasks: Format Dispatch

**Input**: Design documents from `specs/001-format-dispatch/`

**Prerequisites**: plan.md ✓ spec.md ✓ research.md ✓ data-model.md ✓ contracts/cli.md ✓

**Test mandate**: Constitution Principle III is NON-NEGOTIABLE. Test files MUST be
written and confirmed failing before any implementation file is written. Tasks are
ordered to enforce this.

**Source paths**: All source files at repo root (flat package per plan.md structure
decision). Module path: `github.com/hoyt-harness/conform`.

---

## Phase 1: Setup

**Purpose**: Repository scaffolding, Go module, PCS hardening, branch.

- [ ] T001 Create git branch `001-format-dispatch` and set as working branch
- [ ] T002 Initialize Go module: `go mod init github.com/hoyt-harness/conform` in repo root, producing `go.mod`
- [ ] T003 [P] Create `.gitignore` — exclude `conform` and `conform.exe` binaries, `*.test`, `coverage.out`
- [ ] T004 [P] Create `hooks/ci-check.sh` — run `go build ./...`, `go vet ./...`, `go test ./...`; blocking on failure
- [ ] T005 [P] Create `hooks/pre-push` — invoke `hooks/ci-check.sh`; set executable bit
- [ ] T006 [P] Create `.gitattributes` — `hooks/* text eol=lf` to prevent CRLF corruption on Windows
- [ ] T007 [P] Create `.github/workflows/ci.yml` — Ubuntu runner, Go 1.22+, runs `hooks/ci-check.sh`
- [ ] T008 [P] Copy `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md` from PCS repo-template (`D:\Engineering\PositronikalCodingStandards\repo-template\`)
- [ ] T009 [P] Copy `.github/PULL_REQUEST_TEMPLATE.md` from PCS repo-template
- [ ] T010 [P] Create `LICENSE` (GPL-3.0, matching other Positronikal original projects — confirm with Hoyt if a different license is preferred for a general-audience tool)

**Checkpoint**: Repo structure is in place. Module compiles (nothing to compile yet). Proceed to foundational phase.

---

## Phase 2: Foundational — Types and Failing Tests

**Purpose**: Define shared types; write all test cases before any implementation.
All tests in this phase MUST fail (red phase) before Phase 3 begins.

**⚠️ CRITICAL**: Do not write implementation logic until T017 confirms red phase.

- [ ] T011 Create `types.go` — define `FileType` (string typedef + constants for all 11 supported types plus `Unknown`), `FormatResult` struct (`Status`, `Path`, `Message`), `Formatter` struct (`Binary`, `Args []string`, `Fallback *Formatter`), `InvocationConfig` struct (`Paths []string`, `CheckMode bool`), `Status` type with constants `StatusChanged`, `StatusOK`, `StatusSkip`, `StatusError`
- [ ] T012 Create `detect_test.go` — failing tests covering:
  - Extension → FileType mapping for all 11 supported types (one case per extension)
  - Shebang detection for `python`, `python3`, `sh`, `bash`, `pwsh`, `powershell`
  - Extensionless file with unrecognized shebang → `Unknown`
  - No extension, no shebang → `Unknown`
  - `.yaml` and `.yml` both map to YAML type
  - `.jsx` and `.tsx` map to JS/TS types respectively
- [ ] T013 Create `dispatch_test.go` — failing tests covering:
  - CHANGED: file content differs after formatter runs (mock formatter that changes content)
  - OK: file content identical after formatter runs (mock formatter that is a no-op)
  - ERROR: formatter binary not on PATH → FormatResult with StatusError
  - SKIP: Unknown FileType → FormatResult with StatusSkip, file unchanged
  - `--check` / CHANGED: file would change → StatusChanged returned, original file UNCHANGED
  - `--check` / OK: file already correct → StatusOK returned
  - goimports fallback: goimports absent but gofmt present → gofmt used
  - goimports absent AND gofmt absent → StatusError
  - Multi-file: three files with different outcomes in one call → three FormatResults in order
- [ ] T014 [P] Create minimal stub `detect.go` — package declaration + `Detect(path string) FileType` signature only (returns `Unknown` always), so the package compiles
- [ ] T015 [P] Create minimal stub `dispatch.go` — package declaration + `Format(path string, cfg InvocationConfig) FormatResult` signature only (returns zero value always)
- [ ] T016 [P] Create minimal stub `main.go` — `package main` + `func main() {}` only
- [ ] T017 Run `go test ./...` — confirm ALL tests FAIL (red phase). Do not proceed to Phase 3 until this is confirmed.

**Checkpoint**: Red phase confirmed. All types defined. All test cases encoded. Implementation begins.

---

## Phase 3: US1 + US3 + US4 — Core Detection and Dispatch

**Goal**: All four status codes (CHANGED, OK, SKIP, ERROR) working end-to-end.
After this phase, the binary formats files correctly, skips unknowns, and handles
missing formatters gracefully — satisfying User Stories 1, 3, and 4.

**Independent Test**: `go test ./...` passes; manual Scenarios 1–5 from quickstart.md all produce correct output.

### Implementation

- [ ] T018 [US1] [US4] Implement `detect.go` — `Detect(path string) FileType`:
  - Build extension table as `map[string]FileType` covering all extensions in contracts/cli.md
  - If extension matches → return mapped FileType
  - If extension unrecognized or absent → open file, read first line, match shebang pattern `#!.*/(python3?|sh|bash|pwsh|powershell)(\s|$)` and env-shebang form
  - If shebang matches → return corresponding FileType
  - Otherwise → return `Unknown`
- [ ] T019 [US1] [US3] Implement `dispatch.go`:
  - Build formatter registry as `map[FileType]Formatter` per the table in contracts/cli.md and research.md decisions (including goimports→gofmt fallback chain, ruff double-pass with `--select I`, PowerShell `-NoProfile`)
  - Implement `resolveFormatter(ft FileType) (*Formatter, error)` — walks fallback chain via `exec.LookPath`; returns error if entire chain absent
  - Implement `readDigest(path string) ([]byte, error)` — SHA-256 of file contents
  - Implement `Format(path string, cfg InvocationConfig) FormatResult`:
    - `Detect(path)` → if Unknown → return `FormatResult{Status: StatusSkip, ...}`
    - `resolveFormatter` → if error → return `FormatResult{Status: StatusError, Message: "<binary> not installed"}`
    - In check mode: copy file to `<path>.conform-check`, run formatter on copy, compare digest with original, remove copy; if different → StatusChanged (original unchanged); if same → StatusOK
    - In normal mode: capture pre-digest, run formatter in-place, capture post-digest; if different → StatusChanged with message "N lines reformatted" (derive N from line count delta); if same → StatusOK
    - Formatter subprocess non-zero exit → StatusError with stderr as message
- [ ] T020 [US1] Implement `main.go` — full entry point:
  - `flag.Bool("check", false, "report what would change without modifying files")`
  - Parse remaining args as file paths; error and exit 2 if none provided
  - Build `InvocationConfig{Paths: args, CheckMode: *check}`
  - Loop over paths: call `Format(path, cfg)`, print result line in contract format: `"%-7s %s  %s\n"` with status, path, message
  - If check mode and any StatusChanged → exit 1 after all files processed
- [ ] T021 [US1] [US3] [US4] Run `go test ./...` — all tests must pass (green phase)

**Checkpoint**: Binary builds and all four status codes work. Validate manually with quickstart.md Scenarios 1–5 before proceeding.

---

## Phase 4: US2 — CI Formatting Gate

**Goal**: `--check` flag prevents file modification and exits non-zero when any file would change.
`go test ./...` passes for all US2 scenarios (T013 check-mode tests).

**Independent Test**: quickstart.md Scenarios 6 and 7 both produce correct exit codes.

**Note**: The `--check` logic was implemented as part of `dispatch.go` in T019 (check mode path). This phase verifies the test suite covers it and validates end-to-end behavior.

### Implementation

- [ ] T022 [US2] Run `go test ./...` — confirm check-mode tests from T013 pass (they should pass already from Phase 3 implementation; this task confirms it explicitly)
- [ ] T023 [US2] Manual validation: run quickstart.md Scenario 6 (check mode, already formatted → exit 0) and Scenario 7 (check mode, would change → exit 1, file unchanged)
- [ ] T024 [US2] Confirm no files are modified in check mode by verifying file mtime and content are unchanged after `./conform --check <unformatted-file>`

**Checkpoint**: CI gate behavior confirmed. `--check` is reliable.

---

## Phase 5: Polish and Release

**Purpose**: Documentation, hook wiring, PCS compliance, release.

- [ ] T025 [P] Write `README.md` — installation (build from source, install to `~/.local/bin/`), supported file types table (from contracts/cli.md), PostToolUse hook wiring snippet (from quickstart.md), `--check` CI usage example; written for general audience (no Positronikal-specific references)
- [ ] T026 [P] Write `USING` — end-user guide: prerequisites (formatter installs per type), hook wiring, check mode, exit codes reference; mirrors USING pattern from other Positronikal repos
- [ ] T027 [P] Wire PostToolUse hook in `~/.claude/settings.json` — replace existing per-language Go hook with single catch-all `conform` hook per quickstart.md PostToolUse section; verify hook fires on next Edit/Write call in a live session
- [ ] T028 [P] Run positronikal-check compliance scan: `uv run --project D:/Engineering/PositronikalCodingStandards positronikal-check D:/Engineering/conform`; address any hard-fail findings
- [ ] T029 Run `hooks/ci-check.sh` locally — full clean pass required before tagging
- [ ] T030 Validate all 8 quickstart.md scenarios end-to-end
- [ ] T031 `git commit` all files on branch `001-format-dispatch` with message `feat: initial implementation of format dispatch binary`
- [ ] T032 Merge branch to `main`, tag `v0.1.0`, push with tags, create GitHub release `v0.1.0`
- [ ] T033 Add `conform` repo to `03-OPERATIONS/Engineering/CLAUDE.md` registry (update `pre-repo` → `active`)
- [ ] T034 Update `~/.claude/settings.json` allowlist to permit `Bash(conform:*)` without prompting

---

## Dependencies and Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 complete
- **Phase 3 (US1+US3+US4)**: Depends on Phase 2 complete (red phase confirmed)
- **Phase 4 (US2)**: Depends on Phase 3 complete — --check mode implemented in Phase 3
- **Phase 5 (Polish)**: Depends on Phase 4 complete

### Within Phase 3

```
T018 (detect.go) ──┐
                   ├──→ T021 (go test — green)
T019 (dispatch.go)─┤
                   │
T020 (main.go) ────┘
```

T018 and T019 can be written in parallel (different files). T020 depends on both being complete for the binary to build. T021 confirms all pass.

### Parallel Opportunities in Phase 1

T003–T010 are all independent (different files) and can be written in a single pass.

### Parallel Opportunities in Phase 2

T014, T015, T016 (stubs) can be written in parallel.
T012 and T013 (test files) can be written in parallel.
T011 (types.go) must precede T012, T013, T014, T015, T016.

---

## Implementation Strategy

### MVP (Phase 1–3 only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Types + failing tests (red phase confirmed at T017)
3. Complete Phase 3: Core implementation (green phase at T021)
4. **STOP and VALIDATE**: Build binary, run quickstart.md Scenarios 1–5
5. The MVP binary is usable for the PostToolUse hook immediately

### Full Delivery

1. MVP (above)
2. Phase 4: Verify CI gate behavior (Scenarios 6–7)
3. Phase 5: Polish, README, hook wiring, tag v0.1.0

---

## Notes

- `[P]` = different files, no shared state — can be written in any order within the phase
- Tests in Phase 2 encode the entire contract; they are the specification made executable
- The `--check` mode is implemented in Phase 3 (T019) but verified in Phase 4 — no separate implementation phase needed
- US3 (graceful degradation) and US4 (unknown type) are paths through the same code as US1; their test cases in T013 and T012 verify them independently
- Commit after Phase 2 (red phase) and after Phase 3 (green phase) to create a clean audit trail of TDD discipline
