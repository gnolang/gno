package avl

import (
	"sort"
	"strings"
	"testing"
)

func TestTraverseByOffset(t *testing.T) {
	const testStrings = `Alfa
Alfred
Alpha
Alphabet
Beta
Beth
Book
Browser`
	tt := []struct {
		name string
		desc bool
	}{
		{"ascending", false},
		{"descending", true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sl := strings.Split(testStrings, "\n")

			// sort a first time in the order opposite to how we'll be traversing
			// the tree, to ensure that we are not just iterating through with
			// insertion order.
			sort.Sort(sort.StringSlice(sl))
			if !tc.desc {
				reverseSlice(sl)
			}

			r := NewNode(sl[0], nil)
			for _, v := range sl[1:] {
				r, _ = r.Set(v, nil)
			}

			// then sort sl in the order we'll be traversing it, so that we can
			// compare the result with sl.
			reverseSlice(sl)

			var result []string
			for i := 0; i < len(sl); i++ {
				r.TraverseByOffset(i, 1, tc.desc, true, func(n *Node) bool {
					result = append(result, n.Key())
					return false
				})
			}

			if !slicesEqual(sl, result) {
				t.Errorf("want %v got %v", sl, result)
			}

			for l := 2; l <= len(sl); l++ {
				// "slices"
				for i := 0; i <= len(sl); i++ {
					max := i + l
					if max > len(sl) {
						max = len(sl)
					}
					exp := sl[i:max]
					actual := []string{}

					r.TraverseByOffset(i, l, tc.desc, true, func(tr *Node) bool {
						actual = append(actual, tr.Key())
						return false
					})
					// t.Log(exp, actual)
					if !slicesEqual(exp, actual) {
						t.Errorf("want %v got %v", exp, actual)
					}
				}
			}
		})
	}
}

func slicesEqual(w1, w2 []string) bool {
	if len(w1) != len(w2) {
		return false
	}
	for i := 0; i < len(w1); i++ {
		if w1[0] != w2[0] {
			return false
		}
	}
	return true
}

func reverseSlice(ss []string) {
	for i := 0; i < len(ss)/2; i++ {
		j := len(ss) - 1 - i
		ss[i], ss[j] = ss[j], ss[i]
	}
}
