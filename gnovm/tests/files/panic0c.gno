package main

type S struct {
	s string
}

func f(strs []string, s S, m map[string]string, t func(s string) string) {
	panic("wtf")
}

func main() {
	const strs = []string{"a", "b"}
	const s = S{"c"}
	const m = map[string]string{"d": "gg", "test": "test"}
	const t = func(s string) string {
		return s
	}
	f(strs, s, m, t)
}

// Stacktrace:
// panic wtf
// f<VPBlock(3,1)>([]string<len=2>,main.S,map[string]string<len=2>,func(s string)( string))
//     main/files/panic0c.gno:8
// main<VPBlock(1,2)>()
//     main/files/panic0c.gno:18

// Error:
// wtf
