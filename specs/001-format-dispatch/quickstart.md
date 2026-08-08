# Quickstart: conform Validation Guide

**Phase 1 output** | **Date**: 2026-08-08

This guide documents the runnable scenarios that prove `conform` works end-to-end.
Run these after building the binary and before tagging a release.

---

## Prerequisites

- Go 1.22+ installed and on PATH
- At least one formatter installed (see [CLI contract](contracts/cli.md) for the full list)
- For full validation: `ruff`, `gofmt`, `goimports`, `shfmt`, `prettier`, `pwsh`

---

## Build

```bash
cd D:/Engineering/conform
go build -o conform .
# Windows:
go build -o conform.exe .
```

Verify:
```bash
./conform --help     # or .\conform.exe --help on Windows
```

---

## Scenario 1: Python file is reformatted (CHANGED)

```bash
# Create a badly-formatted Python file
cat > /tmp/test_conform.py << 'EOF'
def   hello(  x,y ):
    return x+y
EOF

./conform /tmp/test_conform.py
```

**Expected output**:
```
CHANGED /tmp/test_conform.py  N lines reformatted
```

**Expected file content after**: properly formatted Python (consistent spacing,
import-sorted if imports were present).

---

## Scenario 2: Already-formatted file (OK)

```bash
cat > /tmp/test_ok.go << 'EOF'
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
EOF

./conform /tmp/test_ok.go
```

**Expected output**:
```
OK      /tmp/test_ok.go  no changes
```

---

## Scenario 3: Unknown file type (SKIP)

```bash
echo "not a recognized file type" > /tmp/test.xyz
./conform /tmp/test.xyz
```

**Expected output**:
```
SKIP    /tmp/test.xyz    unknown type
```

---

## Scenario 4: Formatter not installed (ERROR, exit 0)

```bash
# Temporarily rename shfmt (or test with a type whose formatter isn't installed)
cat > /tmp/test.sh << 'EOF'
#!/bin/bash
echo "hello"
EOF

# With shfmt absent:
./conform /tmp/test.sh
```

**Expected output**:
```
ERROR   /tmp/test.sh     shfmt not installed
```

**Expected exit code**: `0`

---

## Scenario 5: Multiple files, mixed results

```bash
./conform /tmp/test_conform.py /tmp/test_ok.go /tmp/test.xyz
```

**Expected output** (one line per file, in argument order):
```
CHANGED /tmp/test_conform.py  N lines reformatted
OK      /tmp/test_ok.go       no changes
SKIP    /tmp/test.xyz         unknown type
```

---

## Scenario 6: Check mode — already formatted (exit 0)

```bash
./conform --check /tmp/test_ok.go
```

**Expected output**:
```
OK      /tmp/test_ok.go  no changes
```

**Expected exit code**: `0`

**Expected file**: unchanged

---

## Scenario 7: Check mode — would change (exit 1)

```bash
# Re-create the unformatted Python file (Scenario 1 modified it)
cat > /tmp/test_check.py << 'EOF'
def   hello(  x,y ):
    return x+y
EOF

./conform --check /tmp/test_check.py
echo "Exit code: $?"
```

**Expected output**:
```
CHANGED /tmp/test_check.py  would reformat
```

**Expected exit code**: `1`

**Expected file**: unchanged (original unformatted content preserved)

---

## Scenario 8: Shebang detection (extensionless shell script)

```bash
cat > /tmp/myhook << 'EOF'
#!/bin/bash
echo "pre-push hook"
EOF

./conform /tmp/myhook
```

**Expected output**: `CHANGED` or `OK` depending on whether shfmt would reformat it.
**Must not produce**: `SKIP` (the shebang must be detected).

---

## Running the test suite

```bash
cd D:/Engineering/conform
go test ./...
```

**Expected**: all tests pass, no failures.

---

## PostToolUse hook validation

Add to `~/.claude/settings.json` (replacing existing per-language hooks):

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

**Validation**: In a Claude Code session, use the Edit tool to write a Python file with
inconsistent indentation. The `CHANGED` or `OK` line should appear in the session
output immediately after the Edit tool call completes.
