// Command tzfmt normalises UTC offsets given as command-line arguments,
// one per line of output. It's a thin wrapper around the tzfmt package,
// useful for checking a value from a shell pipeline without writing Go.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jyoung49/tzfmt"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run does the actual work of the command and returns the process exit
// code. It's split out from main so it can be exercised by tests without
// the os.Exit call at the end of a real invocation tearing down the test
// binary.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: tzfmt <offset> [offset...]")
		return 2
	}

	exitCode := 0
	for _, arg := range args {
		normalized, err := tzfmt.Normalize(arg)
		if err != nil {
			fmt.Fprintln(stderr, err)
			exitCode = 1
			continue
		}
		fmt.Fprintln(stdout, normalized)
	}
	return exitCode
}
