package split

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Split struct {
	Input        io.Reader
	ErrOutput    io.Writer
	OutputDir    string
	OutputPrefix string
}

const (
	ExitSuccess = 0
	ExitError   = 1
)

var (
	ErrTooManyArgs   = errors.New("too many args")
	ErrNoFlag        = errors.New("no flag set")
	ErrInvalidNumber = errors.New("invalid number: Numerical result out of range")
)

func Main() int {
	s := &Split{
		ErrOutput: os.Stderr,
		OutputDir: ".",
	}

	config, args := parse()
	if len(args) > 2 { // parse 関数の中で処理すべきかもしれない
		fmt.Fprintln(s.ErrOutput, "Error:", ErrTooManyArgs)
		return ExitError
	}

	if len(args) == 0 {
		s.Input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintln(s.ErrOutput, "Error:", err)
			return ExitError
		}
		defer f.Close()
		s.Input = f
	}
	if len(args) <= 1 {
		s.OutputPrefix = "x"
	} else {
		s.OutputPrefix = args[1]
	}
	return s.Run(config)
}

func (s *Split) Run(config *Config) int {
	if err := config.validate(); err != nil {
		if errors.Is(err, ErrNoFlag) {
			config.Flag[FlagL] = "1000"
		} else {
			fmt.Fprintln(s.ErrOutput, "Error:", err)
			return ExitError
		}
	}

	var err error
	if l, exist := config.Flag[FlagL]; exist {
		err = s.withL(l)
	} else if n, exist := config.Flag[FlagN]; exist {
		err = s.withN(n)
	} else {
		b := config.Flag[FlagB]
		err = s.withB(b)
	}
	if err != nil {
		fmt.Fprintln(s.ErrOutput, "Error:", err)
		return ExitError
	}
	return ExitSuccess
}

func (s *Split) withL(l string) error {
	const maxBuf = 100000 // 行の最大長 要検討

	num, err := strconv.Atoi(l)
	if err != nil {
		return fmt.Errorf("withL: %w", err)
	} else if num <= 0 {
		return fmt.Errorf("withL: %w, %q", ErrInvalidNumber, l)
	}

	sc := bufio.NewScanner(s.Input)
	sc.Buffer(make([]byte, 4096), maxBuf)
	continueLoop := true
	for suffix := []rune(originSuffix); continueLoop; nextSuffix(&suffix) {
		buf := new(bytes.Buffer)
		for i := 0; i < num; i++ {
			if sc.Scan() {
				b := append(sc.Bytes(), '\n') // 多分遅い
				if _, err := buf.Write(b); err != nil {
					return fmt.Errorf("withL: %w", err)
				}
			} else {
				continueLoop = false
				break
			}
		}
		if err := sc.Err(); err != nil {
			return fmt.Errorf("withL: %w", err)
		}

		dst := s.OutputPrefix + string(suffix)
		if err := s.writeToFile(buf, dst, buf.Len()); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("withL: %w", err)
		}
	}
	return nil
}

// s.Input が標準入力などの場合: io.ReadAll の場所でブロッキングが発生する？
// 本家の split コマンドだと indeterminableなんちゃら と言われて異常終了する
// どの挙動を採用するかは要検討
// chunk の K 要素とラウンドロビン分割は一旦スルーしてる。後で対処する
func (s *Split) withN(n string) error {
	chunk, err := parseChunk(n)
	if err != nil {
		return fmt.Errorf("withN: %w", err)
	} else if chunk.k >= 0 {
		panic(ErrUnimplemented)
	}
	switch chunk.chunkType { // 後で処理をまとめる
	case byByte:
		b, err := io.ReadAll(s.Input)
		if err != nil {
			return fmt.Errorf("withN split by byte: %w", err)
		}
		buf := bytes.NewBuffer(b)
		oneFileSize := buf.Len() / chunk.n
		suffix := []rune(originSuffix)
		for i := 0; i < chunk.n; i++ {
			dst := s.OutputPrefix + string(suffix)
			size := oneFileSize
			if i == chunk.n-1 {
				size = buf.Len() // 最後のファイルにあまりを全部書き込む
			}
			if err := s.writeToFile(buf, dst, size); err != nil {
				if !errors.Is(err, io.EOF) { // EOFは処理継続
					return fmt.Errorf("withN split by byte: %w", err)
				}
			}
			nextSuffix(&suffix)
		}
	case byLine:
		// ReadAll を使わずに読み込みながら改行を数える方が綺麗かもしれない？ なんとも言えない 要検討
		b, err := io.ReadAll(s.Input)
		if err != nil {
			return fmt.Errorf("withN split by line: %w", err)
		}
		bs := bytes.SplitAfter(b, []byte("\n"))
		oneFileLine := len(bs) / chunk.n
		suffix := []rune(originSuffix)
		for i := 0; i < chunk.n; i++ {
			dst := s.OutputPrefix + string(suffix)
			line := oneFileLine
			if i == chunk.n-1 {
				line = len(bs) // 最後のファイルにあまりを全部書き込む
			}
			buf := new(bytes.Buffer)
			for j := 0; j < line; j++ {
				if _, err := buf.Write(bs[0]); err != nil {
					return fmt.Errorf("withN split by line: %w", err)
				}
				bs = bs[1:]
			}
			if err := s.writeToFile(buf, dst, buf.Len()); err != nil {
				if !errors.Is(err, io.EOF) { // EOFは処理継続
					return fmt.Errorf("withN split by line: %w", err)
				}
			}
			nextSuffix(&suffix)
		}
	case byLineRoundRobin:
		panic(ErrUnimplemented)
	default:
		panic("invalid chunk") // 事前にバリデーションしてるので不要かも？要検討
	}
	return nil
}

func (s *Split) withB(b string) error {
	num, err := parseSize(b)
	if err != nil {
		return fmt.Errorf("withB: %w", err)
	}

	for suffix := []rune(originSuffix); ; nextSuffix(&suffix) {
		dst := s.OutputPrefix + string(suffix)
		if err := s.writeToFile(s.Input, dst, num); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("withB: %w", err)
		}
	}
	return nil
}

func (s *Split) writeToFile(r io.Reader, dst string, n int) error {
	path := filepath.Join(s.OutputDir, dst)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("writeToFile: %w", err)
	}
	defer f.Close() // Close のエラーも処理しましょう
	if _, err := io.CopyN(f, r, int64(n)); err != nil {
		return fmt.Errorf("writeToFile: %w", err)
	}
	return nil
}

const originSuffix = "aa"

// もう少し使いやすくしたい
// クロージャーを使ってsuffix生成機みたいにする？
func nextSuffix(s *[]rune) {
	t := *s
	for i := len(t) - 1; i >= 0; i-- {
		if t[i] < 'z' {
			t[i]++
			return
		}
		t[i] = 'a'
	}

	// all z
	*s = []rune(strings.Repeat("a", len(t)+1))
}
