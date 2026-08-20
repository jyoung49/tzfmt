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

There's also a small CLI in `cmd/tzfmt` for checking a value from a
shell prompt:

```
$ go run ./cmd/tzfmt +0530 EST
+05:30
tzfmt: "EST" is an ambiguous timezone abbreviation; use an explicit offset like -05:00
```

## Status

Early. The offset normaliser and its test suite are the whole thing
right now — see the roadmap for what's next.
