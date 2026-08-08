# Data Model: Format Dispatch

**Phase 1 output** | **Date**: 2026-08-08

`conform` has no persistent state. The "entities" here are the runtime structures
that flow through a single invocation.

---

## FileType

Represents a recognized source file category.

| Field | Type | Description |
|---|---|---|
| `name` | string | Human-readable label (e.g., `"python"`, `"go"`, `"shell"`) |
| `extensions` | []string | File extensions that map to this type (e.g., `[".py"]`) |
| `shebangs` | []string | Interpreter tokens matched in shebang lines (e.g., `["python3", "python"]`) |

**Values** (initial set):
`python`, `go`, `shell`, `powershell`, `json`, `yaml`, `markdown`, `css`,
`javascript`, `typescript`, `c`

**Special value**: `unknown` — returned when no extension or shebang matches.

---

## Formatter

Represents an external formatting tool associated with a FileType.

| Field | Type | Description |
|---|---|---|
| `binary` | string | Executable name looked up via PATH (e.g., `"ruff"`, `"gofmt"`) |
| `args` | []string | Arguments passed to the binary, with `{file}` as a placeholder |
| `fallback` | *Formatter | Optional fallback if this formatter is not found (e.g., `gofmt` fallback for `goimports`) |

**State transitions**:
- `binary` found via PATH → formatter is available
- `binary` not found, `fallback` exists → use fallback
- `binary` not found, no `fallback` → emit `ERROR`

---

## FormatResult

The outcome of processing one file. Produced for every input path, regardless of
success or failure.

| Field | Type | Description |
|---|---|---|
| `status` | Status | One of `CHANGED`, `OK`, `SKIP`, `ERROR` |
| `path` | string | The input file path as provided by the caller |
| `message` | string | Human-readable detail (e.g., `"3 lines reformatted"`, `"shfmt not installed"`) |

**Status values**:

| Status | Meaning | Exit contribution |
|---|---|---|
| `CHANGED` | File was reformatted | Exit 0 (normal); Exit 1 in `--check` mode |
| `OK` | File was already correctly formatted | Exit 0 always |
| `SKIP` | File type not recognized | Exit 0 always |
| `ERROR` | Formatter absent, crashed, or file unreadable | Exit 0 always |

---

## InvocationConfig

Parsed from CLI arguments at startup. Immutable for the lifetime of the process.

| Field | Type | Description |
|---|---|---|
| `paths` | []string | One or more file paths to process |
| `checkMode` | bool | If true, no files are modified; exit 1 if any would change |
