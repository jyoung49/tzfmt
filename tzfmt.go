// Package tzfmt normalises the many ways people write a UTC offset into a
// single canonical form: signed, zero-padded, colon-separated (+05:30).
//
// Offsets show up in logs, CSV exports, and API responses in whatever shape
// whatever produced them happened to use: "Z", "UTC", "gmt-3", "+0530",
// "+5:30", "-00:00". Downstream code that just wants to compare or sort
// these strings has to deal with all of that. This package does the
// dealing-with once so callers get one shape back.
package tzfmt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MinOffsetMinutes and MaxOffsetMinutes are the real-world bounds on a UTC
// offset: Baker Island (-12:00) to the Line Islands (+14:00). Anything
// outside this range is almost certainly a typo or a unit mistake, so
// ParseOffset rejects it rather than silently accepting it.
const (
	MinOffsetMinutes = -12 * 60
	MaxOffsetMinutes = 14 * 60
)

// offsetPattern matches an optional UTC/GMT label followed by a signed
// hour, with an optional colon-separated or concatenated minute component.
// The label and the sign may have whitespace between them ("UTC +5:30").
var offsetPattern = regexp.MustCompile(`^(?:(UTC|GMT)\s*)?([+-])\s*(\d{1,2})(?::?(\d{2}))?$`)

// abbreviationPattern matches a bare run of letters that isn't UTC, GMT, or
// Z. These are timezone abbreviations like EST or IST, and they're
// intentionally not supported: the same three letters mean different
// offsets in different regions (IST is UTC+5:30 in India and UTC+1 in
// Ireland), so guessing would silently produce a wrong answer.
var abbreviationPattern = regexp.MustCompile(`^[A-Z]{2,6}$`)

// ParseOffset parses a UTC offset written in any of the forms described in
// the package doc and returns the offset as a signed number of minutes
// from UTC.
func ParseOffset(input string) (int, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, fmt.Errorf("tzfmt: empty offset")
	}

	upper := strings.ToUpper(s)
	if upper == "Z" || upper == "UTC" || upper == "GMT" {
		return 0, nil
	}

	if m := offsetPattern.FindStringSubmatch(upper); m != nil {
		sign := 1
		if m[2] == "-" {
			sign = -1
		}

		hours, err := strconv.Atoi(m[3])
		if err != nil {
			return 0, fmt.Errorf("tzfmt: invalid hour in offset %q", input)
		}

		minutes := 0
		if m[4] != "" {
			minutes, err = strconv.Atoi(m[4])
			if err != nil {
				return 0, fmt.Errorf("tzfmt: invalid minute in offset %q", input)
			}
		}

		if minutes > 59 {
			return 0, fmt.Errorf("tzfmt: minute component out of range in offset %q", input)
		}

		total := sign * (hours*60 + minutes)
		if total < MinOffsetMinutes || total > MaxOffsetMinutes {
			return 0, fmt.Errorf("tzfmt: offset %q is out of range (must be between -12:00 and +14:00)", input)
		}

		return total, nil
	}

	if abbreviationPattern.MatchString(upper) {
		return 0, fmt.Errorf("tzfmt: %q is an ambiguous timezone abbreviation; use an explicit offset like -05:00", input)
	}

	return 0, fmt.Errorf("tzfmt: unrecognised offset %q", input)
}

// FormatOffset renders a signed number of minutes from UTC in canonical
// form: a sign, two-digit hours, a colon, two-digit minutes. Zero is always
// rendered as "+00:00", regardless of the sign of the input, since a
// negative zero offset ("-00:00" from upstream data) has no meaning
// distinct from positive zero.
func FormatOffset(minutes int) string {
	sign := byte('+')
	if minutes < 0 {
		sign = '-'
		minutes = -minutes
	}
	return fmt.Sprintf("%c%02d:%02d", sign, minutes/60, minutes%60)
}

// Normalize parses a messy UTC offset and returns it in canonical form. It
// is a convenience wrapper around ParseOffset and FormatOffset for the
// common case where the caller just wants the cleaned-up string.
func Normalize(input string) (string, error) {
	minutes, err := ParseOffset(input)
	if err != nil {
		return "", err
	}
	return FormatOffset(minutes), nil
}
