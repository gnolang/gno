package main

import "errors"

const (
	MYERS = "myers"
)

type Algorithm interface {
	Do() (srcDiff []LineDifferrence, dstDiff []LineDifferrence)
}

func AlgorithmFactory(src, dst []string, algoType string) (Algorithm, error) {
	switch algoType {
	case MYERS:
		return NewMyers(src, dst), nil
	default:
		return nil, errors.New("unknown algorithm type")
	}
}
