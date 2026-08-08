package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	check := flag.Bool("check", false, "report what would change without modifying files")
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "conform: no files specified")
		os.Exit(2)
	}

	cfg := InvocationConfig{
		Paths:     paths,
		CheckMode: *check,
	}

	anyChanged := false
	for _, path := range cfg.Paths {
		result := Format(path, cfg)
		fmt.Printf("%-7s %s  %s\n", result.Status, result.Path, result.Message)
		if result.Status == StatusChanged {
			anyChanged = true
		}
	}

	if cfg.CheckMode && anyChanged {
		os.Exit(1)
	}
}
