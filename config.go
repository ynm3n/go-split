package split

import (
	"errors"
	"flag"
	"fmt"
)

var (
	ErrTooManyFlags           = errors.New("cannot split in more than one way")
	ErrInvalidFlagValue       = errors.New("invalid flag value")
	ErrInvalidFlag            = errors.New("invalid flag")
	ErrIndeterminableFileSize = errors.New("cannot determine file size")
	ErrUnimplemented          = errors.New("未実装です…")
)

var ErrTodo = errors.New("todo")

const (
	usageL = "put NUMBER lines/records per output file"
	usageN = "generate CHUNKS output files"
	usageB = "put SIZE bytes per output file"
)

type FlagType int

const (
	FlagL FlagType = iota
	FlagN
	FlagB
)

type Config struct {
	Flag map[FlagType]string
}

// ひとまず全部文字列で受け取る
// この関数ではバリデーションチェックはしない 読み込むだけ
// flag.Parse 関数によってパニックが起こりうる
func parse() (*Config, []string) {
	config := &Config{
		Flag: make(map[FlagType]string),
	}
	var l, n, b string
	flag.StringVar(&l, "l", "", usageL)
	flag.StringVar(&n, "n", "", usageN)
	flag.StringVar(&b, "b", "", usageB)
	flag.Parse()

	fs := []struct {
		Type   FlagType
		String string
	}{
		{FlagL, l}, {FlagN, n}, {FlagB, b},
	}
	for _, f := range fs {
		if len(f.String) > 0 {
			config.Flag[f.Type] = f.String
		}
	}

	args := flag.Args()
	return config, args
}

func (config *Config) validate() error {
	switch len(config.Flag) {
	case 0:
		return fmt.Errorf("config validation: %w", ErrNoFlag)
	case 1:
		for ft, f := range config.Flag {
			if ft < 0 || ft > 2 { // フラグ追加時にバグりそう より良い方法を探すべきかも
				return fmt.Errorf("config validation: %w", ErrInvalidFlag)
			} else if len(f) == 0 {
				return fmt.Errorf("config validation: %w, %q", ErrInvalidFlagValue, f)
			}
		}
	default:
		return fmt.Errorf("config validation: %w", ErrTooManyFlags)
	}
	return nil
}
