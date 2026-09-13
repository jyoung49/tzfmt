package tzfmt_test

import (
	"fmt"
	"time"

	"github.com/jyoung49/tzfmt"
)

func ExampleNormalize() {
	for _, raw := range []string{"Z", "gmt-3", "+0530", "UTC +5:30", "-00:00"} {
		normalized, err := tzfmt.Normalize(raw)
		if err != nil {
			fmt.Println(raw, "->", err)
			continue
		}
		fmt.Println(raw, "->", normalized)
	}
	// Output:
	// Z -> +00:00
	// gmt-3 -> -03:00
	// +0530 -> +05:30
	// UTC +5:30 -> +05:30
	// -00:00 -> +00:00
}

func ExampleParseOffset() {
	minutes, err := tzfmt.ParseOffset("+05:30")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(minutes)
	// Output: 330
}

func ExampleExtractOffset() {
	// The numeric offset wins over the trailing zone abbreviation, so IST
	// doesn't get rejected as ambiguous when +0530 already says what it means.
	minutes, err := tzfmt.ExtractOffset("2009-11-10 23:00:00 +0530 IST")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(minutes)
	// Output: 330
}

func ExampleResolveZone() {
	offset, err := tzfmt.ResolveZone("America/New_York", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(offset)
	// Output: -05:00
}
