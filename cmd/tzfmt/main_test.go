package main

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "no arguments",
			args:       nil,
			wantCode:   2,
			wantStderr: "usage: tzfmt <offset> [offset...]\n",
		},
		{
			name:       "single valid offset",
			args:       []string{"+0530"},
			wantCode:   0,
			wantStdout: "+05:30\n",
		},
		{
			name:       "multiple valid offsets",
			args:       []string{"Z", "gmt-3"},
			wantCode:   0,
			wantStdout: "+00:00\n-03:00\n",
		},
		{
			name:       "single invalid offset",
			args:       []string{"EST"},
			wantCode:   1,
			wantStderr: "tzfmt: \"EST\" is an ambiguous timezone abbreviation; use an explicit offset like -05:00\n",
		},
		{
			name:       "mixed valid and invalid keeps going and reports failure",
			args:       []string{"+0530", "EST", "-0500"},
			wantCode:   1,
			wantStdout: "+05:30\n-05:00\n",
			wantStderr: "tzfmt: \"EST\" is an ambiguous timezone abbreviation; use an explicit offset like -05:00\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(c.args, &stdout, &stderr)

			if got != c.wantCode {
				t.Errorf("run(%v) exit code = %d, want %d", c.args, got, c.wantCode)
			}
			if stdout.String() != c.wantStdout {
				t.Errorf("run(%v) stdout = %q, want %q", c.args, stdout.String(), c.wantStdout)
			}
			if stderr.String() != c.wantStderr {
				t.Errorf("run(%v) stderr = %q, want %q", c.args, stderr.String(), c.wantStderr)
			}
		})
	}
}
