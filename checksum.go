package main

import "fmt"

// Kind identifies which checksum scheme a candidate code belongs to,
// decided purely by its digit count once separators are stripped.
type Kind string

const (
	KindISBN10 Kind = "ISBN-10"
	KindISBN13 Kind = "ISBN-13"
	KindUPCA   Kind = "UPC-A"
)

// classify returns the Kind that matches the length of a cleaned code,
// or "" if the length doesn't correspond to a scheme we check.
func classify(clean string) Kind {
	switch len(clean) {
	case 10:
		return KindISBN10
	case 12:
		return KindUPCA
	case 13:
		return KindISBN13
	default:
		return ""
	}
}

// checkDigit computes the expected check digit for clean (last character
// excluded from the calculation) and reports whether it matches what's
// actually there. It also returns the expected digit as a string so
// callers can build a useful message ("X" for ISBN-10 checksum 10).
func checkDigit(kind Kind, clean string) (want string, ok bool, err error) {
	switch kind {
	case KindISBN10:
		return checkISBN10(clean)
	case KindISBN13:
		return checkISBN13(clean)
	case KindUPCA:
		return checkUPCA(clean)
	default:
		return "", false, fmt.Errorf("unknown kind %q", kind)
	}
}

// checkISBN10 implements the ISO 2108 check: digits weighted 10 down to 1,
// summed, and the total must be divisible by 11. The check digit itself
// may be the letter X, standing in for the value 10.
func checkISBN10(clean string) (string, bool, error) {
	if len(clean) != 10 {
		return "", false, fmt.Errorf("ISBN-10 needs 10 characters, got %d", len(clean))
	}
	sum := 0
	for i := 0; i < 9; i++ {
		d := clean[i]
		if d < '0' || d > '9' {
			return "", false, fmt.Errorf("non-digit %q in ISBN-10 body", d)
		}
		sum += int(d-'0') * (10 - i)
	}
	rem := sum % 11
	wantVal := (11 - rem) % 11
	want := fmt.Sprintf("%d", wantVal)
	if wantVal == 10 {
		want = "X"
	}
	last := string(clean[9])
	return want, last == want, nil
}

// checkISBN13 and checkUPCA share the same GS1 mod-10 algorithm: digits
// alternate weights of 1 and 3, and the check digit brings the total up
// to the next multiple of 10. The two schemes are out of phase with each
// other, though: a UPC-A is really an EAN-13 with a leading zero, so its
// leftmost digit carries the weight that would otherwise land on that
// implied zero (3), not the weight EAN-13 gives its own leftmost digit (1).
func checkISBN13(clean string) (string, bool, error) {
	return checkMod10(clean, 13, 1)
}

func checkUPCA(clean string) (string, bool, error) {
	return checkMod10(clean, 12, 3)
}

func checkMod10(clean string, length, firstWeight int) (string, bool, error) {
	if len(clean) != length {
		return "", false, fmt.Errorf("expected %d digits, got %d", length, len(clean))
	}
	otherWeight := 4 - firstWeight
	sum := 0
	for i := 0; i < length-1; i++ {
		d := clean[i]
		if d < '0' || d > '9' {
			return "", false, fmt.Errorf("non-digit %q in body", d)
		}
		weight := firstWeight
		if i%2 == 1 {
			weight = otherWeight
		}
		sum += int(d-'0') * weight
	}
	wantVal := (10 - sum%10) % 10
	want := fmt.Sprintf("%d", wantVal)
	last := string(clean[length-1])
	return want, last == want, nil
}
