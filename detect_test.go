// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectByExtension(t *testing.T) {
	cases := []struct {
		ext      string
		expected FileType
	}{
		{".py", Python},
		{".go", Go},
		{".sh", Shell},
		{".bash", Shell},
		{".ps1", PowerShell},
		{".psm1", PowerShell},
		{".json", JSON},
		{".yaml", YAML},
		{".yml", YAML},
		{".md", Markdown},
		{".css", CSS},
		{".js", JavaScript},
		{".jsx", JavaScript},
		{".ts", TypeScript},
		{".tsx", TypeScript},
		{".c", C},
		{".h", C},
		{".xyz", Unknown},
	}
	for _, tc := range cases {
		f, err := os.CreateTemp(t.TempDir(), "*"+tc.ext)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		got := Detect(f.Name())
		if got != tc.expected {
			t.Errorf("Detect(ext=%q) = %q, want %q", tc.ext, got, tc.expected)
		}
	}
}

func TestDetectByShebang(t *testing.T) {
	cases := []struct {
		name     string
		shebang  string
		expected FileType
	}{
		{"bash-direct", "#!/bin/bash\n", Shell},
		{"bash-env", "#!/usr/bin/env bash\n", Shell},
		{"sh-direct", "#!/bin/sh\n", Shell},
		{"sh-env", "#!/usr/bin/env sh\n", Shell},
		{"python3-direct", "#!/usr/bin/python3\n", Python},
		{"python3-env", "#!/usr/bin/env python3\n", Python},
		{"python-env", "#!/usr/bin/env python\n", Python},
		{"pwsh-direct", "#!/usr/bin/pwsh\n", PowerShell},
		{"pwsh-env", "#!/usr/bin/env pwsh\n", PowerShell},
		{"powershell-env", "#!/usr/bin/env powershell\n", PowerShell},
		{"not-shebang", "# just a comment\n", Unknown},
		{"empty", "", Unknown},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "myscript")
			if err := os.WriteFile(path, []byte(tc.shebang), 0o644); err != nil {
				t.Fatal(err)
			}
			got := Detect(path)
			if got != tc.expected {
				t.Errorf("Detect(shebang=%q) = %q, want %q", tc.shebang, got, tc.expected)
			}
		})
	}
}

func TestDetectExtensionTakesPrecedence(t *testing.T) {
	// A .py file with a bash shebang should still be detected as Python.
	path := filepath.Join(t.TempDir(), "confusing.py")
	if err := os.WriteFile(path, []byte("#!/bin/bash\nprint('hello')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Detect(path)
	if got != Python {
		t.Errorf("Detect(.py with bash shebang) = %q, want %q", got, Python)
	}
}
