package split_test

import (
	"testing"

	"github.com/ynm3n/go-split"
)

func TestRun(t *testing.T) {
	prefix := "x"
	tests := []struct {
		name   string
		config *split.Config
		want   error
	}{
		{
			name: "no error (Flag: L)",
			config: &split.Config{
				SelectedFlag: split.FlagL,
				L:            1000,
				SrcName:      "testdata/input.txt",
				Prefix:       prefix,
			},
			want: nil,
		},
		// { 工事中
		// 	name: "no error (Flag: N)",
		// 	config: &split.Config{
		// 		SelectedFlag: split.FlagN,
		// 		N:            Chunk{},
		// 		SrcName:      "testdata/input.txt",
		// 		Prefix:       prefix,
		// 	},
		// 	want: nil,
		// },
		{
			name: "no error (Flag: B)",
			config: &split.Config{
				SelectedFlag: split.FlagB,
				B:            10000,
				SrcName:      "testdata/input.txt",
				Prefix:       prefix,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tt.config.Prefix = dir + "/" + tt.config.Prefix
			if err := split.Run(tt.config); err != tt.want {
				t.Errorf("test for %v: want = %v, but got = %v", tt.config, tt.want, err)
			}
		})
	}
}
