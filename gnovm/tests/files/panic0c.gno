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
// f<VPBlock(3,1)>((const (slice[("a" string),("b" string)] []string)), (const (struct{("c" string)} main.S)), (const (map{("d" string):("gg" string),("test" string):("test" string)} map[string]string)), (const ( func(s string)( string))))
//     main/files/panic0c.gno:8
// main()
//     main/files/panic0c.gno:18

// Error:
// wtf
