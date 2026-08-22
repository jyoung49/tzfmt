package tzfmt

import (
	"testing"
	"time"
)

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

func TestExtractOffset(t *testing.T) {
	cases := []struct {
		name      string
		timestamp string
		want      int
	}{
		{"ISO with Z", "2024-01-15T10:30:00Z", 0},
		{"ISO with fractional seconds and Z", "2024-01-15T10:30:00.123456Z", 0},
		{"ISO with colon offset", "2024-01-15T10:30:00+05:30", 330},
		{"ISO with negative colon offset", "2024-01-15T10:30:00-05:00", -300},
		{"ISO with concatenated offset", "2024-01-15T10:30:00+0530", 330},
		{"space-separated UTC label only", "2024-01-15 10:30:00 UTC", 0},
		{"space-separated UTC label with offset", "2024-01-15 10:30:00 UTC+5:30", 330},
		{"space-separated UTC label with space before sign", "2024-01-15 10:30:00 UTC +5:30", 330},
		{"lowercase gmt label glued to offset", "2024-01-15 10:30:00 gmt-3", -180},
		{"go time.String() UTC", "2009-11-10 23:00:00 +0000 UTC", 0},
		{"go time.String() with positive offset and name", "2024-06-01 08:15:00 +0530 IST", 330},
		{"go time.String() with negative offset and name", "2024-06-01 08:15:00 -0700 MST", -420},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ExtractOffset(c.timestamp)
			if err != nil {
				t.Fatalf("ExtractOffset(%q) returned unexpected error: %v", c.timestamp, err)
			}
			if got != c.want {
				t.Errorf("ExtractOffset(%q) = %d, want %d", c.timestamp, got, c.want)
			}
		})
	}
}

func TestExtractOffsetRejects(t *testing.T) {
	cases := []struct {
		name      string
		timestamp string
	}{
		{"no offset at all", "2024-01-15T10:30:00"},
		{"bare date, no time component", "2024-01-05"},
		{"trailing ambiguous abbreviation with no numeric offset", "2024-01-15 10:30:00 EST"},
		{"empty string", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, err := ExtractOffset(c.timestamp); err == nil {
				t.Errorf("ExtractOffset(%q) = %d, want error", c.timestamp, got)
			}
		})
	}
}

func TestResolveZone(t *testing.T) {
	cases := []struct {
		name string
		zone string
		at   time.Time
		want string
	}{
		{"fixed offset, no DST", "Asia/Kolkata", time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC), "+05:30"},
		{"standard time", "America/New_York", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), "-05:00"},
		{"daylight time", "America/New_York", time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC), "-04:00"},
		{"standard time, UK", "Europe/London", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), "+00:00"},
		{"daylight time, UK", "Europe/London", time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC), "+01:00"},
		{"UTC zone name itself", "UTC", time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC), "+00:00"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ResolveZone(c.zone, c.at)
			if err != nil {
				t.Fatalf("ResolveZone(%q, %v) returned unexpected error: %v", c.zone, c.at, err)
			}
			if got != c.want {
				t.Errorf("ResolveZone(%q, %v) = %q, want %q", c.zone, c.at, got, c.want)
			}
		})
	}
}

func TestResolveZoneRejects(t *testing.T) {
	cases := []struct {
		name string
		zone string
	}{
		{"unknown zone", "Not/AZone"},
		{"empty string", ""},
		{"host-dependent Local", "Local"},
		{"lowercase local, still host-dependent", "local"},
	}

	at := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, err := ResolveZoneOffset(c.zone, at); err == nil {
				t.Errorf("ResolveZoneOffset(%q, ...) = %d, want error", c.zone, got)
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
