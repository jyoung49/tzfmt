package tzfmt

import "testing"

// FuzzNormalize checks two invariants that should hold for any input,
// not just the hand-picked cases in TestNormalize: a canonical result
// should be idempotent under re-normalisation, and it should always be
// within the offset range this package promises.
func FuzzNormalize(f *testing.F) {
	seeds := []string{
		"Z", "z", "UTC", "GMT",
		"+0530", "-0500", "+5:30", "-00:00", "gmt-3", "UTC+5:30",
		"EST", "IST",
		"", "   ", "++5", "UTCX", "+15:00", "-13:00",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		out, err := Normalize(input)
		if err != nil {
			return
		}

		again, err := Normalize(out)
		if err != nil {
			t.Fatalf("Normalize(%q) = %q, but re-normalising that failed: %v", input, out, err)
		}
		if again != out {
			t.Fatalf("Normalize not idempotent: Normalize(%q) = %q, Normalize(%q) = %q", input, out, out, again)
		}

		minutes, err := ParseOffset(out)
		if err != nil {
			t.Fatalf("ParseOffset failed on canonical output %q: %v", out, err)
		}
		if minutes < MinOffsetMinutes || minutes > MaxOffsetMinutes {
			t.Fatalf("Normalize(%q) produced out-of-range offset %d", input, minutes)
		}
	})
}

// FuzzExtractOffset checks that any offset it manages to pull out of a
// timestamp is within the range ParseOffset would itself accept - the
// two functions should never disagree about what counts as a valid
// offset.
func FuzzExtractOffset(f *testing.F) {
	seeds := []string{
		"2024-01-15T10:30:00Z",
		"2024-01-15T10:30:00.123456Z",
		"2024-01-15T10:30:00+05:30",
		"2024-01-15T10:30:00+0530",
		"2024-01-15 10:30:00 UTC",
		"2024-01-15 10:30:00 UTC+5:30",
		"2024-01-15 10:30:00 gmt-3",
		"2009-11-10 23:00:00 +0000 UTC",
		"2024-06-01 08:15:00 +0530 IST",
		"2024-01-05",
		"",
		"garbage",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		minutes, err := ExtractOffset(input)
		if err != nil {
			return
		}
		if minutes < MinOffsetMinutes || minutes > MaxOffsetMinutes {
			t.Fatalf("ExtractOffset(%q) produced out-of-range offset %d", input, minutes)
		}
	})
}
