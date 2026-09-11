package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// resolvePaths turns the command-line arguments into a flat list of regular
// files to scan. Each argument may be a plain file, a directory (walked
// recursively), or a glob pattern (expanded with filepath.Glob and then
// each match walked in turn). Anything that can't be resolved - a bad
// glob, a path that doesn't exist - is returned as an error rather than
// aborting the whole run, so one bad argument doesn't stop the others
// from being scanned.
func resolvePaths(patterns []string) (files []string, errs []error) {
	for _, pattern := range patterns {
		if !strings.ContainsAny(pattern, "*?[") {
			files, errs = collectPath(pattern, files, errs)
			continue
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", pattern, err))
			continue
		}
		if len(matches) == 0 {
			errs = append(errs, fmt.Errorf("%s: no matching files", pattern))
			continue
		}
		for _, m := range matches {
			files, errs = collectPath(m, files, errs)
		}
	}
	return files, errs
}

// collectPath adds path to files if it's a regular file, or walks it
// recursively and adds every regular file underneath if it's a directory.
// Hidden directories (leading dot, e.g. .git) are skipped since they're
// almost never what someone means to scan and .git in particular is full
// of binary pack data that just wastes time.
func collectPath(path string, files []string, errs []error) ([]string, []error) {
	info, err := os.Stat(path)
	if err != nil {
		return files, append(errs, err)
	}
	if !info.IsDir() {
		return append(files, path), errs
	}
	walkErr := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if d.IsDir() {
			if p != path && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() {
			files = append(files, p)
		}
		return nil
	})
	if walkErr != nil {
		errs = append(errs, walkErr)
	}
	return files, errs
}
