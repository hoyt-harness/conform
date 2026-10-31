# Code Review Notes — conform

## What failure looks like here
A formatter dispatch error where a file type is silently skipped or the wrong
formatter is invoked, producing incorrect output with exit 0.

## Key invariants the reviewer should check
- Every supported file extension maps to exactly one formatter in the dispatch table
- Missing formatters produce an ERROR status line and exit 0 (not a crash)
- `--check` mode never modifies files; exit 1 only when a file would change
- The PostToolUse hook command is correct and includes `|| true`
- No external dependencies — standard library only

## Paths to skip
- `var/` — health reports, not source

## Finding standards
- Every finding must cite a specific file and line number
- Cap nits at 3 per review pass; flag the class, not every instance
