package main

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		clean string
		want  Kind
	}{
		{"70000000", KindISSN},
		{"0134190440", KindISBN10},
		{"036000291452", KindUPCA},
		{"9780134190440", KindISBN13},
		{"9790260000438", KindISMN},
		{"", ""},
		{"123456789", ""},      // 9 digits: no scheme
		{"12345678901", ""},    // 11 digits: no scheme
		{"97800001234567", ""}, // 14 digits: no scheme
	}
	for _, c := range cases {
		if got := classify(c.clean); got != c.want {
			t.Errorf("classify(%q) = %q, want %q", c.clean, got, c.want)
		}
	}
}

// TestClassifyISMNPrefix checks that the ISBN-13/ISMN split really is
// driven by the 9790 prefix and nothing else, since both share the same
// 13-digit mod-10 checksum math.
func TestClassifyISMNPrefix(t *testing.T) {
	if got := classify("9780134190440"); got != KindISBN13 {
		t.Errorf("classify(9780134190440) = %q, want ISBN-13", got)
	}
	if got := classify("9790260000438"); got != KindISMN {
		t.Errorf("classify(9790260000438) = %q, want ISMN", got)
	}
}

func TestCheckDigit(t *testing.T) {
	cases := []struct {
		name  string
		kind  Kind
		clean string
		want  string
		valid bool
	}{
		{"ISSN valid", KindISSN, "03785955", "5", true},
		{"ISSN valid with X check digit", KindISSN, "7000000X", "X", true},
		{"ISSN mismatch", KindISSN, "03785954", "5", false},

		{"ISBN-10 valid", KindISBN10, "0134190440", "0", true},
		{"ISBN-10 valid with X check digit", KindISBN10, "080442957X", "X", true},
		{"ISBN-10 mismatch", KindISBN10, "0134190441", "0", false},

		{"ISBN-13 valid", KindISBN13, "9780134190440", "0", true},
		{"ISBN-13 mismatch", KindISBN13, "9780134190441", "0", false},

		{"UPC-A valid", KindUPCA, "036000291452", "2", true},
		{"UPC-A mismatch", KindUPCA, "036000291453", "2", false},

		{"ISMN valid", KindISMN, "9790260000438", "8", true},
		{"ISMN mismatch", KindISMN, "9790260000439", "8", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want, ok, err := checkDigit(c.kind, c.clean)
			if err != nil {
				t.Fatalf("checkDigit(%q, %q) returned error: %v", c.kind, c.clean, err)
			}
			if want != c.want {
				t.Errorf("checkDigit(%q, %q) want = %q, expected %q", c.kind, c.clean, want, c.want)
			}
			if ok != c.valid {
				t.Errorf("checkDigit(%q, %q) valid = %v, expected %v", c.kind, c.clean, ok, c.valid)
			}
		})
	}
}

func TestCheckDigitWrongLength(t *testing.T) {
	cases := []struct {
		kind  Kind
		clean string
	}{
		{KindISSN, "1234567"},
		{KindISBN10, "123456789"},
		{KindISBN13, "123456789012"},
		{KindUPCA, "12345678901"},
		{KindISMN, "979026000043"},
	}
	for _, c := range cases {
		if _, _, err := checkDigit(c.kind, c.clean); err == nil {
			t.Errorf("checkDigit(%q, %q) with wrong length: want error, got nil", c.kind, c.clean)
		}
	}
}

func TestCheckDigitNonDigit(t *testing.T) {
	cases := []struct {
		kind  Kind
		clean string
	}{
		{KindISSN, "0378595A"},
		{KindISBN10, "013419044A"},
		{KindISBN13, "978013419044A"},
		{KindUPCA, "03600029145A"},
		{KindISMN, "979026000043A"},
	}
	for _, c := range cases {
		if _, _, err := checkDigit(c.kind, c.clean); err == nil {
			t.Errorf("checkDigit(%q, %q) with non-digit body: want error, got nil", c.kind, c.clean)
		}
	}
}

func TestCheckDigitUnknownKind(t *testing.T) {
	if _, _, err := checkDigit(Kind("BOGUS"), "12345678"); err == nil {
		t.Error("checkDigit with unknown kind: want error, got nil")
	}
}

func TestClean(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"978-0-306-40615-7", "9780306406157"},
		{"0 13 419044 0", "0134190440"},
		{"080442957x", "080442957X"},
		{"9780134190440", "9780134190440"},
	}
	for _, c := range cases {
		if got := clean(c.raw); got != c.want {
			t.Errorf("clean(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}
