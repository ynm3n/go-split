package split

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
)

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

// 工事中… 全部正規表現で実装した方がわかりやすそう
// 処理内容がわかりにくい気がするのでもう少し分割したい
func parseSize(s string) (int, error) {
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
