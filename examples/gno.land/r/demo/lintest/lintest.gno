package lintest

import "gno.land/p/nt/avl"

var lintest *avl.Tree

type whatever struct{}

func (w whatever) Iterate(start, end string)

func init() {
	lintest = avl.NewTree()
	lintest.Set("key1", "value1")
	lintest.Set("key2", "value2")
	lintest.Set("key3", "value3")
}

func JoinValues() string {
	result := ""
	lintest.Iterate("", "", func(key string, value any) bool {
		result += value.(string) + ","
		return false
	})
	w := whatever{}
	w.Iterate("", "")
	return result
}

func JoinValuesReverse() string {
	result := ""
	//nolint
	lintest.ReverseIterate("", "", func(key string, value any) bool {
		result += value.(string) + ","
		return false
	})
	return result
}

func Render(path string) string {
	return "Rendering AVL Tree at path: " + path
}
