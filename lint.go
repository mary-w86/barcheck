package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// Finding describes one candidate barcode/ISBN found in a file and
// whether its checksum came out right.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Raw     string `json:"raw"`
	Kind    Kind   `json:"kind"`
	Valid   bool   `json:"valid"`
	Want    string `json:"want,omitempty"`
	Message string `json:"message"`
}

// candidateRe looks for digit runs that are plausibly a barcode or serial
// number - 8, 10, 12, or 13 digits - allowing hyphens or spaces as
// separators and a trailing X for ISBN-10 or ISSN. It's deliberately
// loose; classify() and the checksum functions do the real filtering by
// length and by whether the checksum makes sense.
var candidateRe = regexp.MustCompile(`\b[0-9](?:[0-9]|[- ](?=[0-9]))*[0-9Xx]\b`)

// clean strips separators and upper-cases any trailing X, turning a raw
// match like "978-0-306-40615-7" into "9780306406157".
func clean(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r == '-' || r == ' ' {
			continue
		}
		b.WriteRune(r)
	}
	s := b.String()
	return strings.ToUpper(s)
}

// Lint scans r line by line, treating it as the file named name, and
// returns one Finding per candidate code whose length matches a scheme
// we know how to check. Codes with checksum errors and codes that pass
// are both reported; the caller decides what to do with each.
func Lint(r io.Reader, name string) []Finding {
	var findings []Finding
	scanner := bufio.NewScanner(r)
	// Source lines containing long minified data can exceed bufio's
	// default 64KiB token size, so give ourselves plenty of headroom.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		for _, loc := range candidateRe.FindAllStringIndex(line, -1) {
			raw := line[loc[0]:loc[1]]
			code := clean(raw)
			kind := classify(code)
			if kind == "" {
				continue
			}
			want, ok, err := checkDigit(kind, code)
			if err != nil {
				continue
			}
			f := Finding{
				File:   name,
				Line:   lineNo,
				Column: loc[0] + 1,
				Raw:    raw,
				Kind:   kind,
				Valid:  ok,
			}
			if ok {
				f.Message = string(kind) + " checksum ok"
			} else {
				f.Want = want
				f.Message = string(kind) + " checksum mismatch: expected check digit " + want
			}
			findings = append(findings, f)
		}
	}
	return findings
}
