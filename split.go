package split

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func Run(cf *Config) error {
	var src io.Reader
	if cf.Interactive {
		src = os.Stdin
	} else {
		f, err := os.Open(cf.SrcName)
		if err != nil {
			return err
		}
		defer f.Close()
		src = f
	}

	var err error
	switch cf.SelectedFlag {
	case FlagL:
		err = ByLines(src, cf.Prefix, cf.L)
	case FlagN:
		err = ByFlagN(src, cf.Prefix, cf.N)
	case FlagB:
		err = ByBytes(src, cf.Prefix, cf.B)
	}
	if err != nil {
		return err
	}
	return nil
}

// ByLines split src by n lines.
// n must be positive.
func ByLines(src io.Reader, prefix string, n int) error {
	r := bufio.NewReader(src)
	for suffix := []rune(defaultOriginSuffix); ; nextSuffix(&suffix) {
		name := prefix + string(suffix)
		if err := byLines(r, name, n); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
	}
	return nil
}

// ByBytes split src by n bytes.
// n must be positive.
func ByBytes(src io.Reader, prefix string, n int) error {
	r := bufio.NewReader(src)
	for suffix := []rune(defaultOriginSuffix); ; nextSuffix(&suffix) {
		name := prefix + string(suffix)
		if err := byBytes(r, name, n); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
	}
	return nil
}

// 工事中
func ByFlagN(src io.Reader, prefix string, c Chunk) error {
	return fmt.Errorf("ByFlagN: %w", ErrUnimplemented)
	// fi, err := src.Stat()
	// if err != nil {
	// 	return err
	// }
	// fileSize := int(fi.Size())

	// r := bufio.NewReader(src)
	// suffix := []rune("aa")
	// for i := 0; i < c.N; i++ {
	// 	byteSize := fileSize / c.N
	// 	if i == c.N-1 {
	// 		byteSize += fileSize % c.N
	// 	}
	// 	err = writeByte(r, prefix+string(suffix), byteSize)
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		return err
	// 	}
	// 	nextSuffix(&suffix)
	// }
}

func byLines(r *bufio.Reader, name string, n int) error {
	dst, err := os.Create(name)
	if err != nil {
		return err
	}
	defer dst.Close()
	w := bufio.NewWriter(dst)
	defer w.Flush()
	for i := 0; i < n; i++ {
		b, err := r.ReadBytes('\n')
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func byBytes(r *bufio.Reader, name string, n int) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	_, err = io.CopyN(w, r, int64(n))
	if err != nil {
		return err
	}
	return nil
}

const defaultOriginSuffix = "aa"

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
