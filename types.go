package main

// FileType identifies the language or format of a source file.
type FileType string

const (
	Unknown    FileType = ""
	Python     FileType = "python"
	Go         FileType = "go"
	Shell      FileType = "shell"
	PowerShell FileType = "powershell"
	JSON       FileType = "json"
	YAML       FileType = "yaml"
	Markdown   FileType = "markdown"
	CSS        FileType = "css"
	JavaScript FileType = "javascript"
	TypeScript FileType = "typescript"
	C          FileType = "c"
)

// Status is the outcome of processing one file.
type Status string

const (
	StatusChanged Status = "CHANGED"
	StatusOK      Status = "OK"
	StatusSkip    Status = "SKIP"
	StatusError   Status = "ERROR"
)

// FormatResult is the outcome of processing one file.
type FormatResult struct {
	Status  Status
	Path    string
	Message string
}

// Formatter describes an external formatting tool.
// Binary is the executable name, looked up via PATH at invocation time.
// Args are the arguments to pass; the literal string "{file}" is replaced
// with the absolute path of the target file before invocation.
// Fallback is tried when Binary is not found on PATH.
type Formatter struct {
	Binary   string
	Args     []string
	Fallback *Formatter
}

// InvocationConfig holds the parsed CLI arguments for one conform run.
type InvocationConfig struct {
	Paths     []string
	CheckMode bool
}
