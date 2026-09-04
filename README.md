# tzfmt

Normalises UTC offsets into one canonical shape.

Offsets arrive from wherever they arrive in whatever shape the source
felt like using: `Z`, `UTC`, `gmt-3`, `+0530`, `+5:30`, `-00:00`. If you
need to compare, sort, or store these, you first need them to look the
same. That's all this does.

## What it normalises

- Zulu / UTC / GMT keywords, any case, to `+00:00`
- Concatenated offsets (`+0530`) to colon-separated (`+05:30`)
- Unpadded hours (`+5`, `-5:30`) to zero-padded (`+05:00`, `-05:30`)
- A leading `UTC`/`GMT` label in front of a signed offset (`UTC+5:30`),
  with or without a space before the sign
- `-00:00` to `+00:00` — negative zero isn't a distinct offset
- Surrounding whitespace

## What it deliberately rejects

Timezone abbreviations like `EST` or `IST` are not accepted. The same
three letters mean different things in different places — `IST` is
UTC+5:30 in India and UTC+1 in Ireland — so guessing would just produce
a wrong answer with high confidence. `ParseOffset` returns an error
naming the ambiguity instead of picking one.

Offsets outside the range that actually exists in the world (-12:00 to
+14:00) are also rejected, on the theory that they're a typo rather
than a new timezone.

## Usage

```go
package main

import (
	"fmt"

	"github.com/jyoung49/tzfmt"
)

func main() {
	for _, raw := range []string{"Z", "gmt-3", "+0530", "UTC +5:30", "-00:00"} {
		normalized, err := tzfmt.Normalize(raw)
		if err != nil {
			fmt.Println(raw, "->", err)
			continue
		}
		fmt.Println(raw, "->", normalized)
	}
}
```

```
Z -> +00:00
gmt-3 -> -03:00
+0530 -> +05:30
UTC +5:30 -> +05:30
-00:00 -> +00:00
```

If you just want the offset as minutes (say, to feed a `time.FixedZone`
call), use `ParseOffset` directly:

```go
minutes, err := tzfmt.ParseOffset("+05:30")
// minutes == 330
loc := time.FixedZone("", minutes*60)
```

## Offsets embedded in timestamps

Real timestamps carry the offset as part of a larger string rather than
on its own, so pulling it out first is its own small headache. `ExtractOffset`
does that: it finds the offset at the end of a full timestamp and parses
it, whether it's glued directly onto the time (`2024-01-15T10:30:00+05:30`,
`2024-01-15T10:30:00Z`), set off with a `UTC`/`GMT` label
(`2024-01-15 10:30:00 UTC+5:30`), or has a trailing zone abbreviation
tacked on the way Go's `time.Time.String()` writes it out
(`2009-11-10 23:00:00 +0530 IST`). In that last case the numeric offset
wins over the abbreviation, so `IST` doesn't get rejected as ambiguous
when the `+0530` next to it already says what it means.

```go
minutes, err := tzfmt.ExtractOffset("2009-11-10 23:00:00 +0530 IST")
// minutes == 330
```

A timestamp with no time component - a bare date, say - has nothing to
anchor a bare signed offset to, so `ExtractOffset` reports no offset
found rather than misreading the last digits of the date as one.

## Resolving IANA zone names

Sometimes what you have isn't a raw offset at all but a proper zone name -
`America/New_York`, `Europe/London` - and you want the offset that zone is
actually at for a given instant, DST included. `ResolveZoneOffset` and
`ResolveZone` do that as an opt-in alternative to the raw-offset parsing
above:

```go
minutes, err := tzfmt.ResolveZoneOffset("America/New_York", time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC))
// minutes == -240 (EDT, daylight saving is in effect in July)

offset, err := tzfmt.ResolveZone("America/New_York", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
// offset == "-05:00" (EST, standard time in January)
```

`"Local"` is rejected even though the standard library resolves it: it
means whatever zone the host machine happens to be configured with, which
is exactly the kind of non-portable ambiguity this package exists to avoid.

There's also a small CLI in `cmd/tzfmt` for checking a value from a
shell prompt:

```
$ go run ./cmd/tzfmt +0530 EST
+05:30
tzfmt: "EST" is an ambiguous timezone abbreviation; use an explicit offset like -05:00
```

## Status

Early. Offset normalisation, extracting offsets from full timestamps, and
resolving IANA zone names are all in, each with a test suite. `Normalize`
and `ExtractOffset` also have fuzz targets (`go test -fuzz=FuzzNormalize`,
`go test -fuzz=FuzzExtractOffset`) that check idempotency and range
invariants against inputs the hand-written cases don't cover. There are
also benchmarks (`go test -bench=.`) for the regex-based parse path, since
`ParseOffset` and `ExtractOffset` are the kind of thing that ends up
called in a hot loop over a log file.
