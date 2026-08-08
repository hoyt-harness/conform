package main

import (
	"os"
	"path/filepath"
	"testing"
)

// unformattedGo is valid Go that gofmt would reformat (extra spaces in func signature).
const unformattedGo = "package main\n\nfunc main( ) {\n}\n"

// formattedGo is the gofmt-correct version.
const formattedGo = "package main\n\nfunc main() {\n}\n"

func writeTemp(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFormat_Changed(t *testing.T) {
	path := writeTemp(t, t.TempDir(), "bad.go", unformattedGo)
	result := Format(path, InvocationConfig{Paths: []string{path}})
	if result.Status != StatusChanged {
		t.Errorf("status = %q, want %q", result.Status, StatusChanged)
	}
	got, _ := os.ReadFile(path)
	if string(got) != formattedGo {
		t.Errorf("file not reformatted: got %q", string(got))
	}
}

func TestFormat_OK(t *testing.T) {
	path := writeTemp(t, t.TempDir(), "good.go", formattedGo)
	result := Format(path, InvocationConfig{Paths: []string{path}})
	if result.Status != StatusOK {
		t.Errorf("status = %q, want %q", result.Status, StatusOK)
	}
}

func TestFormat_Skip(t *testing.T) {
	path := writeTemp(t, t.TempDir(), "file.xyz", "some content\n")
	original, _ := os.ReadFile(path)

	result := Format(path, InvocationConfig{Paths: []string{path}})
	if result.Status != StatusSkip {
		t.Errorf("status = %q, want %q", result.Status, StatusSkip)
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(original) {
		t.Error("SKIP must not modify the file")
	}
}

func TestFormat_Error_MissingFormatter(t *testing.T) {
	path := writeTemp(t, t.TempDir(), "bad.go", unformattedGo)

	original := Formatters[Go]
	Formatters[Go] = Formatter{Binary: "nonexistent-formatter-xyzzy"}
	defer func() { Formatters[Go] = original }()

	result := Format(path, InvocationConfig{Paths: []string{path}})
	if result.Status != StatusError {
		t.Errorf("status = %q, want %q", result.Status, StatusError)
	}
}

func TestFormat_GoimportsFallbackToGofmt(t *testing.T) {
	// When goimports is absent, the fallback to gofmt must be used (not ERROR).
	path := writeTemp(t, t.TempDir(), "fallback.go", unformattedGo)

	original := Formatters[Go]
	Formatters[Go] = Formatter{
		Binary: "nonexistent-goimports-xyzzy",
		Args:   []string{"-w", "{file}"},
		Fallback: &Formatter{
			Binary: "gofmt",
			Args:   []string{"-w", "{file}"},
		},
	}
	defer func() { Formatters[Go] = original }()

	result := Format(path, InvocationConfig{Paths: []string{path}})
	if result.Status == StatusError {
		t.Error("must fall back to gofmt, not return ERROR, when goimports is absent")
	}
}

func TestFormat_CheckMode_WouldChange(t *testing.T) {
	path := writeTemp(t, t.TempDir(), "bad.go", unformattedGo)
	before, _ := os.ReadFile(path)

	result := Format(path, InvocationConfig{Paths: []string{path}, CheckMode: true})

	if result.Status != StatusChanged {
		t.Errorf("check mode on unformatted file: status = %q, want %q", result.Status, StatusChanged)
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Error("check mode must not modify the file")
	}
}

func TestFormat_CheckMode_AlreadyFormatted(t *testing.T) {
	path := writeTemp(t, t.TempDir(), "good.go", formattedGo)

	result := Format(path, InvocationConfig{Paths: []string{path}, CheckMode: true})

	if result.Status != StatusOK {
		t.Errorf("check mode on formatted file: status = %q, want %q", result.Status, StatusOK)
	}
	got, _ := os.ReadFile(path)
	if string(got) != formattedGo {
		t.Error("check mode must not modify already-formatted file")
	}
}

func TestFormat_MultipleFiles_OrderPreserved(t *testing.T) {
	dir := t.TempDir()
	bad := writeTemp(t, dir, "bad.go", unformattedGo)
	good := writeTemp(t, dir, "good.go", formattedGo)
	skip := writeTemp(t, dir, "file.xyz", "content\n")

	results := []FormatResult{
		Format(bad, InvocationConfig{Paths: []string{bad}}),
		Format(good, InvocationConfig{Paths: []string{good}}),
		Format(skip, InvocationConfig{Paths: []string{skip}}),
	}

	if results[0].Status != StatusChanged {
		t.Errorf("results[0] = %q, want %q", results[0].Status, StatusChanged)
	}
	if results[1].Status != StatusOK {
		t.Errorf("results[1] = %q, want %q", results[1].Status, StatusOK)
	}
	if results[2].Status != StatusSkip {
		t.Errorf("results[2] = %q, want %q", results[2].Status, StatusSkip)
	}
}
