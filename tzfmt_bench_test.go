package tzfmt

import "testing"

// BenchmarkParseOffset covers the keyword fast path (no regex) alongside
// the regex-matched shapes, so a regression in either is visible.
func BenchmarkParseOffset(b *testing.B) {
	cases := []struct {
		name  string
		input string
	}{
		{"keyword", "UTC"},
		{"concatenated", "+0530"},
		{"colon", "-05:00"},
		{"labeled", "UTC+5:30"},
		{"reject_abbreviation", "EST"},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ParseOffset(c.input)
			}
		})
	}
}

func BenchmarkNormalize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Normalize("gmt-3")
	}
}

// BenchmarkExtractOffset covers each of the suffix patterns ExtractOffset
// tries in turn, since a slow pattern near the front of that chain slows
// down every call regardless of which pattern ends up matching.
func BenchmarkExtractOffset(b *testing.B) {
	cases := []struct {
		name      string
		timestamp string
	}{
		{"zulu", "2024-01-15T10:30:00Z"},
		{"bare_colon_offset", "2024-01-15T10:30:00+05:30"},
		{"labeled_suffix", "2024-01-15 10:30:00 UTC+5:30"},
		{"named_suffix", "2009-11-10 23:00:00 +0000 UTC"},
		{"label_only", "2024-01-15 10:30:00 UTC"},
		{"no_match", "2024-01-15T10:30:00"},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ExtractOffset(c.timestamp)
			}
		})
	}
}
