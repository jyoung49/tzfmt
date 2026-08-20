// Command tzfmt normalises UTC offsets given as command-line arguments,
// one per line of output. It's a thin wrapper around the tzfmt package,
// useful for checking a value from a shell pipeline without writing Go.
package main

import (
	"fmt"
	"os"

	"github.com/jyoung49/tzfmt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: tzfmt <offset> [offset...]")
		os.Exit(2)
	}

	exitCode := 0
	for _, arg := range os.Args[1:] {
		normalized, err := tzfmt.Normalize(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			exitCode = 1
			continue
		}
		fmt.Println(normalized)
	}
	os.Exit(exitCode)
}
