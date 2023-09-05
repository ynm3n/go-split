package split

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func Do(cf *Config) error {
	var src io.Reader
	if cf.Interactive {
		src = os.Stdin
	} else {
		f, err := os.Open(cf.SrcName)
		if err != nil { // 疑問: ファイルが存在しない問題はどこで対処すべき？
			return err
		}
		defer f.Close()
		src = f
	}

	var err error
	switch cf.SelectedFlag {
	case FlagL:
		err = splitByLine(src, cf.Prefix, cf.L)
	case FlagN:
		err = splitByChunk(src, cf.Prefix, cf.N)
	case FlagB:
		err = splitByByte(src, cf.Prefix, cf.B)
	}
	return err
}

func splitByLine(src io.Reader, prefix string, ln int) error {
	r := bufio.NewReader(src)
	for suffix := []rune("aa"); ; nextSuffix(&suffix) {
		err := writeLine(r, prefix+string(suffix), ln)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

// 工事中
func splitByChunk(src io.Reader, prefix string, c Chunk) error {
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
	return fmt.Errorf("splitByChunk: %w", ErrUnimplemented)
}

func splitByByte(src io.Reader, prefix string, bt int) error {
	r := bufio.NewReader(src)
	for suffix := []rune("aa"); ; nextSuffix(&suffix) {
		err := writeByte(r, prefix+string(suffix), bt)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
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
