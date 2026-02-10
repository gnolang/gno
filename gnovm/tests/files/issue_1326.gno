// PKGPATH: gno.land/r/test
package test

import (
	"strconv"
)

func init() {
	New()
	println(Delta())
}

func main(cur realm) {
	println(Delta())
}

type Move struct {
	N1, N2, N3 byte
}

type S struct {
	Moves []Move
}

func (s S) clone() S {
	mv := s.Moves
	return S{Moves: mv}
}

func (olds S) change() S {
	s := olds.clone()

	counter++
	s.Moves = append([]Move{}, s.Moves...)
	s.Moves = append(s.Moves, Move{counter, counter, counter})
	return s
}

var (
	el      *S
	counter byte
)

func New() {
	el = new(S)
}

func Delta() string {
	n := el.change()
	*el = n
	return Values()
}

func Values() string {
	s := ""
	for _, val := range el.Moves {
		s += strconv.Itoa(int(val.N1)) + "," + strconv.Itoa(int(val.N2)) + "," + strconv.Itoa(int(val.N3)) + ";"
	}
	return s
}

// Output:
// 1,1,1;
// 1,1,1;2,2,2;
