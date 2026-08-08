# Using conform

## Prerequisites

Install the formatters for the file types you work with. `conform` discovers
them via PATH — install them wherever your PATH resolves.

| File type | Formatter | Install |
|---|---|---|
| Python | `ruff` | `pip install ruff` or `uv tool install ruff` |
| Go | `goimports` | `go install golang.org/x/tools/cmd/goimports@latest` |
| Go (fallback) | `gofmt` | ships with Go |
| Shell | `shfmt` | `go install mvdan.cc/sh/v3/cmd/shfmt@latest` |
| PowerShell | `pwsh` + PSScriptAnalyzer | `Install-Module PSScriptAnalyzer` |
| JSON/YAML/MD/CSS/JS/TS | `prettier` | `npm install -g prettier` |
| C | `clang-format` | ships with LLVM/Clang toolchain |

You do not need to install all formatters. Files whose formatter is absent
produce an `ERROR` status line; all other files are processed normally.

## Basic Usage

Format a single file:

```bash
conform main.go
```

Format multiple files:

```bash
conform src/app.py internal/handler.go config.yaml
```

## Check Mode

Check mode reports what would change without modifying any files. Exit 1 if
any file would be reformatted.

```bash
conform --check main.go
echo $?   # 0 = already correct, 1 = would reformat
```

Useful for CI gates:

```bash
# Fail the build if any committed file is not formatted
conform --check $(git diff --name-only HEAD~1)
```

## Status Lines

Every file produces exactly one output line:

```
STATUS  path/to/file  message
```

| Status | Meaning |
|---|---|
| `CHANGED` | File was reformatted |
| `OK` | File was already correctly formatted |
| `SKIP` | File type not recognized |
| `ERROR` | Formatter absent, crashed, or file unreadable |

## PostToolUse Hook (Claude Code)

Wire `conform` into Claude Code so it runs automatically after every file write.
Add to `~/.claude/settings.json`:

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

The `|| true` ensures a missing or erroring formatter never interrupts the
agent session. Status lines appear in the session transcript immediately after
each Edit or Write tool call.

## Exit Codes

| Code | Condition |
|---|---|
| `0` | Normal completion |
| `1` | `--check` mode: at least one file would be reformatted |
| `2` | Invalid arguments (e.g., no files specified) |
