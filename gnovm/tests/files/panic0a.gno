// Test panic with function call with all kind of expressions
package main

type S struct {
	s string
}

func f(it1 int, it2, it3 int, pit *int, b bool, strs []string, s S, m map[string]string, t func(s string) string) {
	panic("wtf")
}

func main() {
	vit := 1
	lit := []int{1}
	var (
		pit *int = &vit
		v   any
	)
	b := true
	v = 1

	f(
		v.(int),
		lit[0],
		*pit,
		&vit,
		!b,
		[]string{"a", "b"},
		S{s: "c"},
		map[string]string{"d": "gg", "test": "test"},
		func(s string) string {
			return s
		},
	)
}

// Stacktrace:
// panic: wtf
// f<VPBlock(3,1)>(v<VPBlock(1,3)>.((const-type int)),lit<VPBlock(1,1)>[0],*pit<VPBlock(1,2)>,&vit<~VPBlock(1,0)>,!b<VPBlock(1,4)>,(const-type []string)<len=2>,(const-type main.S)<len=1>,(const-type map[string]string)<len=2>,func(s (const-type string)) .res.0 (const-type string){ ... })
//     main/panic0a.gno:9
// main<VPBlock(1,2)>()
//     main/panic0a.gno:22

// Error:
// wtf
