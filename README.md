# barcheck

A linter for ISBN and barcode checksums. Point it at a file and it flags
every ISBN-10, ISBN-13, or UPC-A code whose check digit doesn't match the
rest of the number, with a file:line:column for each one.

The problem this solves: a metadata feed, a CSV export, a markdown catalog,
whatever - somewhere in it there's a 13-digit number that looks like an
ISBN but was typo'd, truncated, or pasted from the wrong row. Nothing
about the number itself looks wrong until you actually run the checksum.
This tool finds those numbers and runs the checksum for you.

## What it checks

- **ISBN-10** - ISO 2108, weighted sum mod 11, check digit can be `X`.
- **ISBN-13 / EAN-13** - GS1 mod-10, alternating weights of 1 and 3.
- **UPC-A** - the same mod-10 algorithm, one digit shorter and out of
  phase with EAN-13 (a UPC-A is an EAN-13 with an implied leading zero).

Any digit run in the input that's 10, 12, or 13 digits long (hyphens and
spaces as separators are allowed and ignored) is treated as a candidate
and checked. Shorter or longer runs are left alone.

## Usage

Build it:

```
go build -o barcheck .
```

Run it against one or more files:

```
$ cat catalog.txt
Title: The Go Programming Language
ISBN-13: 978-0-13-419044-1
ISBN-10: 0-13-419044-0

$ ./barcheck catalog.txt
catalog.txt:2:10: ISBN-13 checksum mismatch: expected check digit 0 ("978-0-13-419044-1")
catalog.txt:3:10: ISBN-10 checksum ok ("0-13-419044-0")
```

Exit status is non-zero if any file couldn't be read, or if any code
failed its checksum - so `barcheck` can be dropped into a pre-commit hook
or CI step against a data file.

### JSON output

```
$ ./barcheck --json catalog.txt
[
  {
    "file": "catalog.txt",
    "line": 2,
    "column": 10,
    "raw": "978-0-13-419044-1",
    "kind": "ISBN-13",
    "valid": false,
    "want": "0",
    "message": "ISBN-13 checksum mismatch: expected check digit 0"
  },
  {
    "file": "catalog.txt",
    "line": 3,
    "column": 10,
    "raw": "0-13-419044-0",
    "kind": "ISBN-10",
    "valid": true,
    "message": "ISBN-10 checksum ok"
  }
]
```

This is meant for feeding into other tooling - a CI step that fails the
build only on `"valid": false`, or a script that files a ticket per bad
code, without having to parse the text output.

### Flags

- `--json` - emit the findings above as a JSON array instead of text.
- `--quiet` - in text mode, only print codes that failed their checksum
  (valid codes are still counted for the exit status either way).

## What it doesn't do yet

No recursive directory walking, no ISSN or ISMN support, no way to
suppress a false positive inline. See the code for where those would go.
