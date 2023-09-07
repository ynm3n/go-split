package split

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"regexp"
	"strconv"
)

var (
	ErrTooManyOptions         = errors.New("cannot split in more than one way")
	ErrInvalidFlag            = errors.New("invalid flag value")
	ErrInvalidNumber          = errors.New("invalid number: Numerical result out of range")
	ErrIndeterminableFileSize = errors.New("cannot determine file size")
	ErrUnimplemented          = errors.New("未実装です…")
)

const (
	usageL = "put NUMBER lines/records per output file"
	usageN = "generate CHUNKS output files"
	usageB = "put SIZE bytes per output file"
)

type FlagType uint

const (
	FlagL FlagType = 1 << iota
	FlagN
	FlagB
)

type Config struct {
	SelectedFlag FlagType
	L            int // intに収まらない数値は未対応
	N            Chunk
	B            int // intに収まらない数値は未対応

	SrcName     string
	Prefix      string
	Interactive bool
}

// 工事中
type Chunk struct {
	N int
}

// ParseConfig returns an error if flags or arguments are invalid.
// If there are no cli flag, FlagL is set(default: 1000).
func ParseConfig() (*Config, error) {
	cf := &Config{
		Prefix: "x",
	}
	bit := FlagType(0)
	var tmpFlagL, tmpFlagN, tmpFlagB string
	flag.Func("l", usageL, setFlag(&bit, FlagL, &tmpFlagL))
	flag.Func("n", usageN, setFlag(&bit, FlagN, &tmpFlagN))
	flag.Func("b", usageB, setFlag(&bit, FlagB, &tmpFlagB))
	flag.Parse()
	switch cnt := bits.OnesCount(uint(bit)); cnt {
	case 0:
		cf.SelectedFlag = FlagL
		cf.L = 1000
	case 1:
		cf.SelectedFlag = bit
		var err error
		switch cf.SelectedFlag {
		case FlagL:
			cf.L, err = parseFlagL(tmpFlagL)
		case FlagN:
			cf.N, err = parseFlagN(tmpFlagN)
		case FlagB:
			cf.B, err = parseFlagB(tmpFlagB)
		}
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrTooManyOptions
	}

	args := flag.Args()
	switch len(args) {
	case 2:
		cf.Prefix = args[1]
		fallthrough
	case 1:
		cf.SrcName = args[0]
	case 0:
		if cf.SelectedFlag == FlagN {
			return nil, ErrIndeterminableFileSize
		}
		cf.Interactive = true
	default:
		return nil, fmt.Errorf("extra operand %s", args[2:])
	}
	return cf, nil
}

func setFlag(bit *FlagType, f FlagType, tmp *string) func(string) error {
	return func(s string) error {
		if *bit&f > 0 {
			return ErrTooManyOptions
		}
		*bit |= f
		*tmp = s
		return nil
	}
}

func parseFlagL(s string) (int, error) {
	num, err := parseNumber(s)
	return num, err
}

// 工事中
func parseFlagN(s string) (Chunk, error) {
	num, err := parseNumber(s)
	return Chunk{N: num}, err
}

// 処理内容がわかりにくい気がするのでもう少し分割したい
func parseFlagB(s string) (int, error) {
	if num, err := parseNumber(s); err == nil { // エラーなしの場合
		return num, nil
	} else if errors.Is(err, strconv.ErrRange) || errors.Is(err, ErrInvalidNumber) { // オーバーフローや0以下の数字
		return 0, err
	}

	p, err := regexp.Compile(`^[0-9]*([A-Z]{1,2}|b)$`)
	if err != nil {
		return 0, fmt.Errorf("正規表現を書き直しましょう %w", err)
	}
	if !p.MatchString(s) {
		return 0, ErrInvalidFlag
	}

	pre := s[:len(s)-1]
	suf := s[len(s)-1:]
	if twoChar := (len(s) > 1) && (s[len(s)-2] > '9'); twoChar {
		pre = s[:len(s)-2]
		suf = s[len(s)-2:]
	}
	if len(pre) == 0 {
		pre = "1"
	}

	unit, exist := sizeUnit[suf]
	if !exist {
		return 0, ErrInvalidFlag
	}
	num, err := strconv.Atoi(pre)
	if errors.Is(err, strconv.ErrRange) {
		return 0, err
	}
	if overflowWhenProduct(num, unit) {
		return 0, fmt.Errorf("overflow: %w", ErrInvalidFlag)
	}
	if !isValidNumber(num) {
		return 0, ErrInvalidNumber
	}
	return num, nil
}

const (
	Blocks = 512
	Kilo   = 1000
	Kibi   = 1024
)

var sizeUnit = map[string]int{
	"b": Blocks,

	"KB": Kilo,
	"MB": Kilo * Kilo,
	"GB": Kilo * Kilo * Kilo,
	"TB": Kilo * Kilo * Kilo * Kilo,

	"K": Kibi,
	"M": Kibi * Kibi,
	"G": Kibi * Kibi * Kibi,
	"T": Kibi * Kibi * Kibi * Kibi,
}

func parseNumber(s string) (int, error) {
	num, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	} else if !isValidNumber(num) {
		return 0, ErrInvalidNumber
	}
	return num, nil
}

func isValidNumber(a int) bool {
	return a > 0
}

func overflowWhenProduct(a, b int) bool {
	bg := big.NewInt(int64(a))
	bg.Mul(bg, big.NewInt(int64(b)))
	return bg.Cmp(big.NewInt(math.MaxInt)) > 0
}
