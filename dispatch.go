// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Formatters maps each FileType to its external formatting tool.
// Exported so tests can override individual entries.
var Formatters = map[FileType]Formatter{
	Python: {
		Binary: "ruff",
		Args:   []string{"format", "{file}"},
	},
	Go: {
		Binary: "goimports",
		Args:   []string{"-w", "{file}"},
		Fallback: &Formatter{
			Binary: "gofmt",
			Args:   []string{"-w", "{file}"},
		},
	},
	Shell: {
		Binary: "shfmt",
		Args:   []string{"-w", "{file}"},
	},
	PowerShell: {
		Binary: "pwsh",
		Args:   []string{"-NoProfile", "-Command", "Invoke-Formatter -Path '{file}'"},
	},
	JSON: {
		Binary: "prettier",
		Args:   []string{"--write", "{file}"},
	},
	YAML: {
		Binary: "prettier",
		Args:   []string{"--write", "{file}"},
	},
	Markdown: {
		Binary: "prettier",
		Args:   []string{"--prose-wrap", "always", "--write", "{file}"},
	},
	CSS: {
		Binary: "prettier",
		Args:   []string{"--write", "{file}"},
	},
	JavaScript: {
		Binary: "prettier",
		Args:   []string{"--write", "{file}"},
	},
	TypeScript: {
		Binary: "prettier",
		Args:   []string{"--write", "{file}"},
	},
	C: {
		Binary: "clang-format",
		Args:   []string{"-i", "{file}"},
	},
}

// ruffImportArgs are passed to ruff as a second Python pass for import sorting.
// --select I restricts fixes to isort rules only, avoiding semantic auto-fixes.
var ruffImportArgs = []string{"check", "--fix", "--select", "I", "{file}"}

// Format applies the appropriate formatter to the file at path and returns
// the outcome. In check mode the file is never modified.
func Format(path string, cfg InvocationConfig) FormatResult {
	ft := Detect(path)
	if ft == Unknown {
		return FormatResult{Status: StatusSkip, Path: path, Message: "unknown type"}
	}

	f, ok := Formatters[ft]
	if !ok {
		return FormatResult{Status: StatusSkip, Path: path, Message: "unknown type"}
	}

	resolved, err := resolveFormatter(f)
	if err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}

	if cfg.CheckMode {
		return checkFormat(path, *resolved, ft)
	}
	return applyFormat(path, *resolved, ft)
}

// resolveFormatter walks the fallback chain until a usable binary is found.
func resolveFormatter(f Formatter) (*Formatter, error) {
	cur := &f
	for cur != nil {
		if _, err := exec.LookPath(cur.Binary); err == nil {
			return cur, nil
		}
		cur = cur.Fallback
	}
	return nil, fmt.Errorf("%s not installed", f.Binary)
}

// applyFormat runs the formatter in-place and reports CHANGED or OK.
func applyFormat(path string, f Formatter, ft FileType) FormatResult {
	before, err := fileDigest(path)
	if err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}

	if err := runFormatter(f, path); err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}

	if ft == Python {
		ruff := Formatter{Binary: "ruff", Args: ruffImportArgs}
		if resolved, err := resolveFormatter(ruff); err == nil {
			_ = runFormatter(*resolved, path)
		}
	}

	after, err := fileDigest(path)
	if err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}

	if !bytes.Equal(before, after) {
		return FormatResult{Status: StatusChanged, Path: path, Message: "reformatted"}
	}
	return FormatResult{Status: StatusOK, Path: path, Message: "no changes"}
}

// checkFormat runs the formatter on a temp copy and reports what would change
// without modifying the original file.
func checkFormat(path string, f Formatter, ft FileType) FormatResult {
	original, err := os.ReadFile(path)
	if err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".conform-check-*")
	if err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(original); err != nil {
		tmp.Close()
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}
	tmp.Close()

	if err := runFormatter(f, tmpPath); err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}

	if ft == Python {
		ruff := Formatter{Binary: "ruff", Args: ruffImportArgs}
		if resolved, err := resolveFormatter(ruff); err == nil {
			_ = runFormatter(*resolved, tmpPath)
		}
	}

	modified, err := os.ReadFile(tmpPath)
	if err != nil {
		return FormatResult{Status: StatusError, Path: path, Message: err.Error()}
	}

	if !bytes.Equal(original, modified) {
		return FormatResult{Status: StatusChanged, Path: path, Message: "would reformat"}
	}
	return FormatResult{Status: StatusOK, Path: path, Message: "no changes"}
}

// runFormatter invokes f on targetPath, substituting {file} in Args.
func runFormatter(f Formatter, targetPath string) error {
	args := make([]string, len(f.Args))
	for i, a := range f.Args {
		args[i] = strings.ReplaceAll(a, "{file}", targetPath)
	}
	cmd := exec.Command(f.Binary, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// fileDigest returns the SHA-256 hash of the file at path.
func fileDigest(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
