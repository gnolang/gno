package v1

import (
	"strconv"

	"gno.land/p/demo/avl"
)

var myTree avl.Tree

type MyStruct struct {
	FieldA string
	FieldB int
}

func (s *MyStruct) String() string {
	if s == nil {
		return "nil"
	}
	return "v1:" + s.FieldA + ":" + strconv.Itoa(s.FieldB)
}

func Get(key string) *MyStruct {
	ret, ok := myTree.Get(key)
	if !ok {
		return nil
	}
	return ret.(*MyStruct)
}

func init() {
	myTree.Set("a", &MyStruct{FieldA: "a", FieldB: 1})
	myTree.Set("b", &MyStruct{FieldA: "b", FieldB: 2})
}
