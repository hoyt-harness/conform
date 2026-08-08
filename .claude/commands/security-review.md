---
description: Security review for conform changes
---

Review the following diff for security issues relevant to a CLI formatting dispatcher:

- Command injection: check that formatter arguments are passed as separate argv elements (exec.Command slice form), never via shell string interpolation
- Path traversal: verify that file paths passed to formatters are not manipulated or escaped in a way that could escape the intended directory
- Temp file safety: confirm temp files in check mode are created with os.CreateTemp (random suffix), written with correct permissions, and removed with defer regardless of outcome
- Subprocess trust: confirm that formatter binaries are resolved via exec.LookPath (PATH lookup only, no user-supplied binary paths)
- Error message leakage: check that formatter stderr captured in ERROR messages does not expose sensitive system paths or environment details beyond what is necessary

Flag any findings with severity (HIGH/MEDIUM/LOW) and the specific line or pattern of concern.
