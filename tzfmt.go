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
	"time"
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

// labeledOffsetSuffix matches a UTC/GMT label immediately followed by a
// signed offset at the end of a timestamp ("...UTC+5:30", "...gmt-3").
// The literal label makes this unambiguous regardless of what precedes it.
var labeledOffsetSuffix = regexp.MustCompile(`(?:UTC|GMT)\s*([+-]\d{1,2}(?::?\d{2})?)$`)

// namedOffsetSuffix matches a signed offset followed by a trailing zone
// abbreviation, the shape produced by Go's time.Time.String() and similar
// formatters ("...+0000 UTC", "...-0700 MST"). The numeric offset is
// authoritative, so the name is matched only to anchor the pattern and is
// discarded rather than passed to ParseOffset - "...+0530 IST" should
// resolve to +05:30, not fail on the ambiguity of IST.
var namedOffsetSuffix = regexp.MustCompile(`([+-]\d{1,2}(?::?\d{2})?)\s+[A-Z]{2,6}$`)

// zuluOffsetSuffix matches a bare "Z" glued directly onto the end of a
// timestamp with no separating whitespace, the ISO 8601 shape
// ("...10:30:00Z", "...10:30:00.123Z").
var zuluOffsetSuffix = regexp.MustCompile(`[0-9.](Z)$`)

// bareOffsetSuffix matches a signed offset glued directly onto the end of
// a timestamp with no label and no separating whitespace, the other ISO
// 8601 shape ("...10:30:00+05:30", "...10:30:00-0500"). Unlike the other
// suffix patterns this one has no label or name to anchor it, so ExtractOffset
// only tries it when the input contains a colon - otherwise a plain date
// like "2024-01-05" would have its "-05" day-of-month misread as an offset.
var bareOffsetSuffix = regexp.MustCompile(`([+-]\d{1,2}(?::?\d{2})?)$`)

// labelOnlySuffix matches a trailing UTC/GMT/abbreviation token with no
// offset digits attached at all ("...2024-01-15 10:30:00 UTC", "... EST").
var labelOnlySuffix = regexp.MustCompile(`(?:^|\s)([A-Z]{2,6})$`)

// ExtractOffset finds and parses the UTC offset embedded at the end of a
// full timestamp, so callers don't have to isolate the offset from the
// rest of the string themselves. It recognises the offset shapes ParseOffset
// does, plus the extra ambiguity that comes from being embedded: a signed
// offset can be glued directly onto the time with no separator
// ("10:30:00+05:30", "10:30:00Z"), and some formats - notably Go's
// time.Time.String() - append a zone abbreviation after an explicit numeric
// offset ("+0000 UTC", "-0700 MST"). In that last case the numeric offset
// wins; the trailing name is discarded rather than rejected as ambiguous,
// since the number already resolves it.
//
// Timestamps that have no time component at all, like a bare date, are not
// supported: without a colon to signal a time part, ExtractOffset can't
// tell a signed offset from the last few digits of the date, so it reports
// no offset found rather than guessing.
func ExtractOffset(timestamp string) (int, error) {
	upper := strings.ToUpper(timestamp)

	if m := labeledOffsetSuffix.FindStringSubmatch(upper); m != nil {
		return ParseOffset(m[1])
	}
	if m := namedOffsetSuffix.FindStringSubmatch(upper); m != nil {
		return ParseOffset(m[1])
	}
	if m := zuluOffsetSuffix.FindStringSubmatch(upper); m != nil {
		return ParseOffset(m[1])
	}
	if strings.Contains(upper, ":") {
		if m := bareOffsetSuffix.FindStringSubmatch(upper); m != nil {
			return ParseOffset(m[1])
		}
	}
	if m := labelOnlySuffix.FindStringSubmatch(upper); m != nil {
		return ParseOffset(m[1])
	}

	return 0, fmt.Errorf("tzfmt: no UTC offset found in %q", timestamp)
}

// ResolveZoneOffset looks up an IANA time zone name (e.g. "America/New_York")
// and returns its UTC offset, in minutes, at the given instant. This is an
// opt-in alternative to a raw offset for callers who have a proper zone name
// rather than a number: unlike a fixed offset, a named zone can resolve to a
// different value depending on at, because of daylight saving.
//
// "Local" is rejected even though the standard library accepts it, because
// it resolves to whatever zone the host machine happens to be configured
// with rather than a fixed, portable answer - the same ambiguity this
// package rejects timezone abbreviations for.
func ResolveZoneOffset(zone string, at time.Time) (int, error) {
	if strings.EqualFold(zone, "Local") {
		return 0, fmt.Errorf("tzfmt: %q is not a portable IANA zone name; it depends on host configuration", zone)
	}

	loc, err := time.LoadLocation(zone)
	if err != nil {
		return 0, fmt.Errorf("tzfmt: unknown IANA zone %q: %w", zone, err)
	}

	_, offsetSeconds := at.In(loc).Zone()
	return offsetSeconds / 60, nil
}

// ResolveZone resolves an IANA zone name to its canonical +HH:MM offset at
// the given instant. It is a convenience wrapper around ResolveZoneOffset
// and FormatOffset for the common case where the caller just wants the
// rendered string.
func ResolveZone(zone string, at time.Time) (string, error) {
	minutes, err := ResolveZoneOffset(zone, at)
	if err != nil {
		return "", err
	}
	return FormatOffset(minutes), nil
}
