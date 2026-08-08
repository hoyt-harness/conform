# Feature Specification: Format Dispatch

**Feature Branch**: `001-format-dispatch`

**Created**: 2026-08-08

**Status**: Draft

**Input**: User description: "A single binary that fires after every AI agent file write,
detects the file type, routes to the appropriate formatter, rewrites the file in place,
and reports one status line per file. Includes a check mode for CI use."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Automatic Post-Write Formatting (Priority: P1)

A developer uses an AI coding agent to write or edit a source file. Without any
manual action, the file is immediately formatted to the project's language conventions
and the developer sees a one-line confirmation of what changed (or that nothing changed).
The developer reviews any diff shown and proceeds — the file is already clean.

**Why this priority**: This is the entire reason the tool exists. Every other story
derives value from this one working reliably across all supported file types.

**Independent Test**: Can be fully tested by writing any supported source file and
confirming the tool runs automatically and produces a status line within 2 seconds.

**Acceptance Scenarios**:

1. **Given** an AI agent writes a Python file with inconsistent indentation, **when** the
   tool runs on that file, **then** the file is reformatted to standard conventions and
   the output reads `CHANGED path/to/file.py  N lines reformatted`.

2. **Given** an AI agent writes a Go file that is already correctly formatted, **when** the
   tool runs on that file, **then** the file is unchanged and the output reads
   `OK     path/to/file.go  no changes`.

3. **Given** an AI agent writes a file in a supported type, **when** the tool runs,
   **then** the result appears in the developer's session output within 2 seconds and
   requires no manual invocation.

---

### User Story 2 — CI Formatting Gate (Priority: P2)

A CI pipeline verifies that all committed source files conform to formatting standards
before a build or merge is allowed. The tool is invoked in check mode, which reports
what would change without modifying any files. If any file would be reformatted, the
pipeline fails with a non-zero exit.

**Why this priority**: Without a CI gate, the post-write hook is the only enforcement
layer. The check mode closes the gap for files committed without running the hook,
and for reviewing formatting compliance independently of any AI session.

**Independent Test**: Can be fully tested by running the tool in check mode on a
correctly-formatted file (expect exit 0, no output changes) and on a
incorrectly-formatted file (expect exit 1, report of what would change).

**Acceptance Scenarios**:

1. **Given** a CI pipeline runs the tool in check mode on a correctly formatted codebase,
   **when** the check completes, **then** the tool exits 0 and reports no files that
   would change.

2. **Given** a CI pipeline runs the tool in check mode and one file has formatting
   violations, **when** the check completes, **then** the tool exits non-zero and reports
   which file would be changed, without modifying it.

---

### User Story 3 — Graceful Handling of Missing Formatters (Priority: P3)

A developer has not installed all formatters for every supported file type. When the
tool encounters a file whose formatter is not present, it reports the gap clearly and
continues processing all remaining files. The session is not interrupted and all other
file types are still formatted normally.

**Why this priority**: The tool fires on every file write. If one missing formatter caused
a hard failure, the developer's session would break whenever they write a file in a
language they haven't fully tooled up for. Graceful degradation is essential for
adoption across diverse environments.

**Independent Test**: Can be fully tested by removing one formatter from the system and
confirming that files of that type produce an `ERROR` status line while files of other
types are still processed correctly.

**Acceptance Scenarios**:

1. **Given** the formatter for shell scripts is not installed, **when** the tool runs on a
   shell script, **then** the output reads `ERROR  path/to/file.sh  <formatter> not
   installed` and the tool exits 0.

2. **Given** a session writes both a Python file and a shell script, and the shell
   formatter is not installed, **when** the tool processes both, **then** the Python file
   is formatted normally and the shell script produces an ERROR line — both are reported
   and neither blocks the other.

---

### User Story 4 — Unknown File Type Passthrough (Priority: P4)

The tool encounters a file type it does not recognize. It skips the file silently with
a `SKIP` status line and exits 0. The developer is informed but not blocked.

**Why this priority**: The tool fires on every write, including configuration files,
templates, and other non-source assets. Unknown types must not cause failures.

**Independent Test**: Can be tested by passing a file with an unrecognized extension
and confirming a `SKIP` line is produced and the file is unchanged.

**Acceptance Scenarios**:

1. **Given** a file with an extension the tool does not recognize, **when** the tool runs,
   **then** the output reads `SKIP   path/to/file.xyz  unknown type` and the file is
   unchanged.

2. **Given** an extensionless file with no recognizable content markers, **when** the tool
   runs, **then** the file is skipped with a `SKIP` status line.

---

### Edge Cases

- What happens when a file is passed that does not exist on disk?
- How does the tool handle a file whose formatter crashes mid-run?
- What happens when the same file is passed more than once in a single invocation?
- How does the tool handle an extensionless file that has a recognizable shebang line?
- What happens when a formatter produces output but also exits non-zero?
- How does the tool behave if it cannot write back to a file due to permissions?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST accept one or more file paths as positional arguments.
- **FR-002**: The tool MUST detect the type of each file from its extension; for files
  with no extension or an unrecognized extension, the tool MUST read the file's first
  line to detect a shebang marker and infer type from it.
- **FR-003**: The tool MUST apply the appropriate formatting tool to each recognized file
  type and rewrite the file in place with the result.
- **FR-004**: The tool MUST emit exactly one status line per input file to standard output,
  using one of four status codes: `CHANGED`, `OK`, `SKIP`, or `ERROR`.
- **FR-005**: The tool MUST exit 0 in all cases except an internal failure unrelated to
  any specific file (e.g., inability to parse its own arguments).
- **FR-006**: The tool MUST accept a `--check` flag. In check mode: no files are modified;
  the tool reports what would change; the tool exits non-zero if any file would be
  reformatted.
- **FR-007**: The tool MUST support the following file categories: Python source files,
  Go source files, POSIX shell scripts, PowerShell scripts, JSON documents, YAML
  documents, Markdown documents, CSS stylesheets, JavaScript and TypeScript source
  files, and C source and header files.
- **FR-008**: When a required formatter is not installed or not discoverable, the tool MUST
  emit an `ERROR` status line for the affected file and continue processing remaining
  files.
- **FR-009**: The tool MUST require no configuration file. All type-to-formatter
  associations are built in and require no per-project setup.
- **FR-010**: The tool MUST be installable as a single self-contained executable with no
  runtime dependencies beyond the external formatters it dispatches to.
- **FR-011**: Errors and diagnostic output unrelated to per-file status MUST be written to
  standard error, not standard output.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After any source file is written during an AI coding session, a formatting
  status line appears in the session output within 2 seconds of the write completing.
- **SC-002**: A file written in any supported language contains zero formatting violations
  immediately after the tool processes it, as confirmed by re-running the formatter
  independently.
- **SC-003**: A developer can install the tool on a new machine and have it produce correct
  formatting output for at least one file type without creating any configuration file.
- **SC-004**: When the tool runs in check mode on a codebase where all files are already
  correctly formatted, it exits 0 and reports zero files that would change.
- **SC-005**: When one formatter is absent from the environment, the tool processes all
  other supported file types correctly in the same invocation, producing no failures
  for those types.
- **SC-006**: The output format is stable: the four status codes (`CHANGED`, `OK`, `SKIP`,
  `ERROR`) and their line structure do not change between versions without a major
  version increment.

## Assumptions

- The tool operates in an environment where an AI coding agent is actively writing files;
  it is invoked automatically by the agent's post-write hook mechanism, not by the
  developer manually (though manual invocation is also supported).
- External formatters are installed separately by the developer and are discoverable
  via the system's executable search path; the tool does not manage formatter installation.
- Input files are text source files; the tool is not designed for binary files and
  behavior on binary inputs is undefined.
- The initial supported file type set covers all languages used in the Positronikal
  engineering environment; support for additional types is an additive, non-breaking
  change.
- The tool is designed for single-workstation use by an individual developer; no
  multi-user, networked, or concurrent-access requirements apply in v1.
- The post-write hook integration is the primary deployment target; CI check-mode use is
  secondary but equally supported.
