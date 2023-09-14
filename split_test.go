package split_test

import (
	"bytes"
	_ "embed"
	"flag"
	"strings"
	"testing"

	"github.com/tenntenn/golden"
	"github.com/ynm3n/go-split"
)

var flagUpdate bool

func init() {
	flag.BoolVar(&flagUpdate, "update", false, "update golden files")
}

//go:embed testdata/input.txt
var src string

func TestSplit_Run(t *testing.T) {
	t.Parallel()
	const (
		prefix   = "x"
		testdata = "testdata"
		inputtxt = "input.txt"
	)
	const flagL, flagN, flagB = split.FlagL, split.FlagN, split.FlagB
	type flagMap = map[split.FlagType]string

	tests := map[string]struct {
		flags    flagMap
		wantExit int
	}{
		"NoFlag":  {flagMap{}, 0},
		"FlagL":   {flagMap{flagL: "10"}, 0},
		"FlagN-N": {flagMap{flagN: "5"}, 0},
		// "FlagN-LN":       {flagMap{flagN: "5"}, 0},
		// "FlagN-RN":       {flagMap{flagN: "5"}, 0},
		// "FlagN-KN":       {flagMap{flagN: "5"}, 0},
		"FlagB":       {flagMap{flagB: "30000"}, 0},
		"TooManyFlag": {flagMap{flagL: "10", flagB: "1b"}, 1},
	}

	for name, tt := range tests {
		name, tt := name, tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			gotErr := new(bytes.Buffer)
			s := &split.Split{
				Input:        strings.NewReader(src),
				ErrOutput:    gotErr,
				OutputDir:    dir,
				OutputPrefix: prefix,
			}
			config := &split.Config{
				Flag: tt.flags,
			}
			if gotExit := s.Run(config); gotExit != tt.wantExit {
				t.Errorf("test for %v: want = %v, but got = %v", config, tt.wantExit, gotExit)
			} else if gotExit == 0 && gotErr.Len() > 0 {
				t.Errorf("test for %v: exit code 0, but errors are outputted %v", config, gotErr)
			}

			got := golden.Txtar(t, dir)
			if flagUpdate {
				golden.Update(t, testdata, name, got)
				return
			}
			if diff := golden.Diff(t, testdata, name, got); diff != "" {
				t.Errorf("golden diff: %v", diff)
			}
		})
	}
}
