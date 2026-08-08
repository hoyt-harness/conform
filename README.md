# conform

A formatting dispatcher for AI coding agents. After every file write, `conform`
detects the file type, routes to the right formatter, rewrites the file in place,
and reports one status line per file.

```
CHANGED src/main.py      reformatted
OK      internal/run.go  no changes
SKIP    .editorconfig    unknown type
ERROR   hooks/pre-push   shfmt not installed
```

## Why

AI agents write code quickly. Without automatic formatting, you discover violations
at commit time — after the fact, as friction. `conform` wires into your agent's
post-write hook so formatting happens immediately after each file is written.
The output line tells you what changed; you review the diff and move on. Code
is clean before you ever think about committing.

## Installation

Requires Go 1.22+.

```bash
git clone https://github.com/hoyt-harness/conform
cd conform
go build -o conform .
cp conform ~/.local/bin/    # or any directory on your PATH
```

`conform` is a single static binary with no runtime dependencies beyond the
external formatters it dispatches to.

## Usage

```
conform [--check] <file> [<file>...]
```

Format one or more files:

```bash
conform main.go
conform src/app.py hooks/pre-push config.json
```

Check mode (no files modified, exit 1 if any would change):

```bash
conform --check main.go
```

## Supported File Types

| Type | Extensions | Formatter required |
|---|---|---|
| Python | `.py` | `ruff` |
| Go | `.go` | `goimports` (falls back to `gofmt`) |
| Shell | `.sh`, `.bash`, extensionless with shebang | `shfmt` |
| PowerShell | `.ps1`, `.psm1` | `pwsh` + PSScriptAnalyzer |
| JSON | `.json` | `prettier` |
| YAML | `.yaml`, `.yml` | `prettier` |
| Markdown | `.md` | `prettier` |
| CSS | `.css` | `prettier` |
| JavaScript | `.js`, `.jsx` | `prettier` |
| TypeScript | `.ts`, `.tsx` | `prettier` |
| C | `.c`, `.h` | `clang-format` |

Formatters are discovered via PATH at runtime. A missing formatter produces an
`ERROR` status line for that file and exits 0 — other files are still processed.

## Exit Codes

| Code | Condition |
|---|---|
| `0` | Normal completion (including SKIP and ERROR lines) |
| `1` | `--check` mode and at least one file would be reformatted |
| `2` | Invalid arguments |

## Claude Code Hook

Add to `~/.claude/settings.json` to run `conform` after every file write:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.file_path' | { read -r f; conform \"$f\"; } 2>/dev/null || true"
          }
        ]
      }
    ]
  }
}
```

## CI Usage

```yaml
- name: Check formatting
  run: conform --check $(git diff --name-only HEAD~1)
```

## License

GNU General Public License v3.0 — see [COPYING.md](COPYING.md).
