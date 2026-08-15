// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"flag"
	"fmt"
	"os"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z".
var version = "dev"

func main() {
	ver := flag.Bool("version", false, "print version and exit")

	check := flag.Bool("check", false, "report what would change without modifying files")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		return
	}

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
