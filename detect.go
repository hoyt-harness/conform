package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// extTable maps lowercase file extensions to their FileType.
var extTable = map[string]FileType{
	".py":   Python,
	".go":   Go,
	".sh":   Shell,
	".bash": Shell,
	".ps1":  PowerShell,
	".psm1": PowerShell,
	".json": JSON,
	".yaml": YAML,
	".yml":  YAML,
	".md":   Markdown,
	".css":  CSS,
	".js":   JavaScript,
	".jsx":  JavaScript,
	".ts":   TypeScript,
	".tsx":  TypeScript,
	".c":    C,
	".h":    C,
}

// shebangTable maps interpreter tokens (the word after the last slash, or
// after "env ") to their FileType.
var shebangTable = map[string]FileType{
	"python":     Python,
	"python3":    Python,
	"sh":         Shell,
	"bash":       Shell,
	"pwsh":       PowerShell,
	"powershell": PowerShell,
}

// Detect returns the FileType for the file at path.
// Extension takes precedence; for extensionless or unrecognized extensions
// the first line is read to detect a shebang.
func Detect(path string) FileType {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != "" {
		if ft, ok := extTable[ext]; ok {
			return ft
		}
		// Known extension, but not in our table — no point reading shebang.
		return Unknown
	}
	return detectShebang(path)
}

// detectShebang reads the first line of path and matches it against known
// shebang patterns. Returns Unknown if the file cannot be read or no match.
func detectShebang(path string) FileType {
	f, err := os.Open(path)
	if err != nil {
		return Unknown
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return Unknown
	}
	line := scanner.Text()

	if !strings.HasPrefix(line, "#!") {
		return Unknown
	}

	// Extract the interpreter token: either the basename of the interpreter
	// path, or the word following "env ".
	rest := strings.TrimSpace(line[2:])
	var token string
	if strings.HasPrefix(rest, "/usr/bin/env ") || strings.HasPrefix(rest, "/bin/env ") {
		parts := strings.Fields(rest)
		if len(parts) >= 2 {
			token = parts[1]
		}
	} else {
		// Direct path like #!/bin/bash or #!/usr/bin/python3
		parts := strings.Fields(rest)
		if len(parts) > 0 {
			token = filepath.Base(parts[0])
		}
	}

	if ft, ok := shebangTable[strings.ToLower(token)]; ok {
		return ft
	}
	return Unknown
}
