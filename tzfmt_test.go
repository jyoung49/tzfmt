package tzfmt

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"bare Z", "Z", "+00:00"},
		{"lowercase z", "z", "+00:00"},
		{"UTC keyword", "UTC", "+00:00"},
		{"lowercase utc", "utc", "+00:00"},
		{"GMT keyword", "GMT", "+00:00"},
		{"already canonical", "+05:00", "+05:00"},
		{"negative zero collapses to positive zero", "-00:00", "+00:00"},
		{"concatenated four digits", "+0530", "+05:30"},
		{"negative concatenated four digits", "-0500", "-05:00"},
		{"single digit hour, no minutes", "+5", "+05:00"},
		{"single digit hour with sign, negative", "-5", "-05:00"},
		{"single digit hour with minutes", "+5:30", "+05:30"},
		{"UTC prefix with offset", "UTC+5:30", "+05:30"},
		{"UTC prefix with space before sign", "UTC +5:30", "+05:30"},
		{"gmt prefix lowercase, no colon", "gmt-3", "-03:00"},
		{"surrounding whitespace", "  +05:00  ", "+05:00"},
		{"max valid offset", "+14:00", "+14:00"},
		{"min valid offset", "-12:00", "-12:00"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Normalize(c.input)
			if err != nil {
				t.Fatalf("Normalize(%q) returned unexpected error: %v", c.input, err)
			}
			if got != c.want {
				t.Errorf("Normalize(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestNormalizeRejects(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"out of range positive", "+15:00"},
		{"out of range negative", "-13:00"},
		{"invalid minute value", "+05:99"},
		{"minute component too short", "+5:3"},
		{"double sign", "++5"},
		{"ambiguous abbreviation EST", "EST"},
		{"ambiguous abbreviation IST", "IST"},
		{"garbage after label", "UTCX"},
		{"just a number, no sign", "5"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, err := Normalize(c.input); err == nil {
				t.Errorf("Normalize(%q) = %q, want error", c.input, got)
			}
		})
	}
}

func TestFormatOffset(t *testing.T) {
	cases := []struct {
		minutes int
		want    string
	}{
		{0, "+00:00"},
		{330, "+05:30"},
		{-300, "-05:00"},
		{840, "+14:00"},
		{-720, "-12:00"},
	}

	for _, c := range cases {
		got := FormatOffset(c.minutes)
		if got != c.want {
			t.Errorf("FormatOffset(%d) = %q, want %q", c.minutes, got, c.want)
		}
	}
}
