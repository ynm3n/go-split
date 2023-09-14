package split

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type chunkType int

const (
	byByte chunkType = iota
	byLine
	byLineRoundRobin
)

type chunk struct {
	n         int
	k         int // 0-indexed; if k is not set, then k == -1
	chunkType chunkType
}

var (
	ErrInvalidChunk = errors.New("invalid chunk")
)

func parseChunk(n string) (*chunk, error) {
	// 正規表現で負の数や0は弾ける
	// intにおさまらない数値は弾けないため、別の手段で確認
	matchN := regexp.MustCompile(`^[1-9][0-9]*$`).MatchString(n)
	matchLN := regexp.MustCompile(`^l/[1-9][0-9]*$`).MatchString(n)
	matchRN := regexp.MustCompile(`^r/[1-9][0-9]*$`).MatchString(n)
	matchKN := regexp.MustCompile(`^[1-9][0-9]*/[1-9][0-9]*$`).MatchString(n)
	matchLKN := regexp.MustCompile(`^l/[1-9][0-9]*/[1-9][0-9]*$`).MatchString(n)
	matchRKN := regexp.MustCompile(`^r/[1-9][0-9]*/[1-9][0-9]*$`).MatchString(n)

	switch { // もうちょっとまとめられそう
	case matchN:
		num, err := strconv.Atoi(n)
		if err != nil {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		chunk := &chunk{
			n:         num,
			k:         -1,
			chunkType: byByte,
		}
		return chunk, nil
	case matchLN, matchRN:
		i := strings.Index(n, "/")
		num, err := strconv.Atoi(n[i+1:])
		if err != nil {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		chunk := &chunk{
			n: num,
			k: -1,
		}
		if matchLN {
			chunk.chunkType = byLine
		} else {
			chunk.chunkType = byLineRoundRobin
		}
		return chunk, nil
	case matchKN:
		i := strings.Index(n, "/")
		numN, err := strconv.Atoi(n[i+1:])
		if err != nil {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		numK, err := strconv.Atoi(n[:i])
		if err != nil {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		if numK > numN {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		chunk := &chunk{
			n:         numN,
			k:         numK - 1,
			chunkType: byByte,
		}
		return chunk, nil
	case matchLKN, matchRKN:
		i, j := strings.Index(n, "/"), strings.LastIndex(n, "/")
		numN, err := strconv.Atoi(n[j+1:])
		if err != nil {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		numK, err := strconv.Atoi(n[i+1 : j])
		if err != nil {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		if numK > numN {
			return nil, fmt.Errorf("parse chunk: %w, %q", err, n)
		}
		chunk := &chunk{
			n: numN,
			k: numK - 1,
		}
		if matchLKN {
			chunk.chunkType = byLine
		} else {
			chunk.chunkType = byLineRoundRobin
		}
		return chunk, nil
	}
	return nil, fmt.Errorf("parse chunk: %w, %q", ErrInvalidChunk, n) // invalid syntax
}
