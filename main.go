package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	ErrTooManyOptions         = errors.New("cannot split in more than one way")
	ErrInvalidFlag            = errors.New("flag must be positive value")
	ErrIndeterminableFileSize = errors.New("cannot determine file size")
	ErrUnimplemented          = errors.New("未実装です…")
)

func main() {
	var flagL, flagN, flagB string
	flag.StringVar(&flagL, "l", "", "put NUMBER lines/records per output file")
	flag.StringVar(&flagN, "n", "", "separate output files into NUMBER bytes; CHUNKS is not supported")
	flag.StringVar(&flagB, "b", "", "put NUMBER bytes per output file")
	flag.Parse()
	args := flag.Args()
	if len(args) > 2 {
		panic(fmt.Errorf("extra operand %s", args[2:]))
	}

	cntEmptyFlag := 0
	for _, f := range []string{flagL, flagN, flagB} {
		if len(f) > 0 {
			cntEmptyFlag++
		}
	}

	switch cntEmptyFlag {
	case 0: // デフォルト動作は "-l 1000"
		err := splitByLine(args, 1000)
		if err != nil {
			panic(err)
		}
	case 1:
		var err error
		if len(flagL) > 0 {
			err = callSplit(args, splitByLine, flagL)
		} else if len(flagN) > 0 {
			err = callSplit(args, splitIntoN, flagN)
		} else {
			err = callSplit(args, splitByByte, flagB)
		}
		if err != nil {
			panic(err)
		}
	default:
		panic(ErrTooManyOptions)
	}
}

func callSplit(args []string, splitFunc func([]string, int) error, flagX string) error {
	x, err := strconv.Atoi(flagX)
	if err != nil {
		return err
	} else if x <= 0 {
		return ErrInvalidFlag
	}
	err = splitFunc(args, x)
	if err != nil {
		return err
	}
	return nil
}

func splitByLine(args []string, l int) error {
	prefix := "x"
	if len(args) == 2 {
		prefix = args[1]
	}

	switch len(args) {
	case 0:
		// 本来は対話型インターフェースになる
		return ErrUnimplemented
	default:
		src, err := os.Open(args[0])
		defer src.Close()
		if err != nil {
			return err
		}
		r := bufio.NewReader(src)
		for suffix := []rune("aa"); ; nextSuffix(&suffix) {
			err = writeLine(r, prefix+string(suffix), l)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
	}
	return nil
}

func writeLine(r *bufio.Reader, name string, cnt int) error {
	dst, err := os.Create(name)
	defer dst.Close()
	if err != nil {
		return err
	}
	w := bufio.NewWriter(dst)
	defer w.Flush()
	for i := 0; i < cnt; i++ {
		bs, err := r.ReadBytes('\n')
		if err != nil {
			return err
		}
		_, err = w.Write(bs)
		if err != nil {
			return err
		}
	}
	return nil
}

func splitByByte(args []string, b int) error {
	prefix := "x"
	if len(args) == 2 {
		prefix = args[1]
	}

	switch len(args) {
	case 0:
		// 本来は対話型インターフェースになる
		return ErrUnimplemented
	default:
		src, err := os.Open(args[0])
		defer src.Close()
		if err != nil {
			return err
		}
		r := bufio.NewReader(src)
		for suffix := []rune("aa"); ; nextSuffix(&suffix) {
			err = writeByte(r, prefix+string(suffix), b)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
	}
	return nil
}

func writeByte(r *bufio.Reader, name string, cnt int) error {
	dst, err := os.Create(name)
	defer dst.Close()
	if err != nil {
		return err
	}
	w := bufio.NewWriter(dst)
	defer w.Flush()
	for i := 0; i < cnt; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		err = w.WriteByte(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func splitIntoN(args []string, n int) error {
	if len(args) == 0 {
		return ErrIndeterminableFileSize
	}
	prefix := "x"
	if len(args) == 2 {
		prefix = args[1]
	}

	src, err := os.Open(args[0])
	defer src.Close()
	if err != nil {
		return err
	}

	fi, err := src.Stat()
	if err != nil {
		return err
	}
	fileSize := int(fi.Size())

	r := bufio.NewReader(src)
	suffix := []rune("aa")
	for i := 0; i < n; i++ {
		byteSize := fileSize / n
		if i == n-1 {
			byteSize += fileSize % n
		}
		err = writeByte(r, prefix+string(suffix), byteSize)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		nextSuffix(&suffix)
	}
	return nil
}

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
