// Command tzfmt normalises UTC offsets given as command-line arguments,
// one per line of output. With -extract or -zone it instead treats each
// argument as a full timestamp or an IANA zone name. It's a thin wrapper
// around the tzfmt package, useful for checking a value from a shell
// pipeline without writing Go.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

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
	fs := flag.NewFlagSet("tzfmt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	extract := fs.Bool("extract", false, "treat each argument as a full timestamp and extract its embedded UTC offset")
	zone := fs.Bool("zone", false, "treat each argument as an IANA zone name and resolve its current UTC offset")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: tzfmt [-extract | -zone] <value> [value...]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *extract && *zone {
		fmt.Fprintln(stderr, "tzfmt: -extract and -zone are mutually exclusive")
		return 2
	}

	values := fs.Args()
	if len(values) == 0 {
		fs.Usage()
		return 2
	}

	exitCode := 0
	for _, v := range values {
		var out string
		var err error
		switch {
		case *extract:
			var minutes int
			if minutes, err = tzfmt.ExtractOffset(v); err == nil {
				out = tzfmt.FormatOffset(minutes)
			}
		case *zone:
			out, err = tzfmt.ResolveZone(v, time.Now())
		default:
			out, err = tzfmt.Normalize(v)
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			exitCode = 1
			continue
		}
		fmt.Fprintln(stdout, out)
	}
	return exitCode
}
