// Command barcheck scans text files for ISBN and barcode-like numbers
// and reports any whose check digit doesn't match, so a bad ISBN in a
// catalog file or a typo'd EAN in a product feed gets caught before it
// ships instead of failing silently at the point of sale.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	jsonOut := flag.Bool("json", false, "emit findings as a JSON array instead of plain text")
	quiet := flag.Bool("quiet", false, "in text mode, only print codes that fail their checksum")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [--json] [--quiet] path [path ...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "each path may be a file, a directory (scanned recursively), or a glob pattern\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	patterns := flag.Args()
	if len(patterns) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	paths, resolveErrs := resolvePaths(patterns)
	hadError := false
	for _, err := range resolveErrs {
		fmt.Fprintf(os.Stderr, "barcheck: %v\n", err)
		hadError = true
	}

	var all []Finding
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "barcheck: %v\n", err)
			hadError = true
			continue
		}
		findings := Lint(f, path)
		f.Close()
		all = append(all, findings...)
	}

	failed := false
	if *jsonOut {
		if all == nil {
			all = []Finding{}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(all); err != nil {
			fmt.Fprintf(os.Stderr, "barcheck: %v\n", err)
			os.Exit(1)
		}
		for _, f := range all {
			if !f.Valid {
				failed = true
				break
			}
		}
	} else {
		for _, f := range all {
			if *quiet && f.Valid {
				continue
			}
			fmt.Printf("%s:%d:%d: %s (%q)\n", f.File, f.Line, f.Column, f.Message, f.Raw)
			if !f.Valid {
				failed = true
			}
		}
	}

	switch {
	case hadError:
		os.Exit(1)
	case failed:
		os.Exit(1)
	default:
		os.Exit(0)
	}
}
