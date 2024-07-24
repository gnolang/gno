package main

type S struct {
	s string
}

func f(strs []string, s S, m map[string]string, t func(s string) string) {
	panic("wtf")
}

func main() {
	f([]string{"a", "b"},
		S{"c"},
		map[string]string{"d": "gg", "test": "test"},
		func(s string) string {
			return s
		},
	)
}

// Stacktrace:
// panic: wtf
// f<VPBlock(3,1)>([](const-type string)<len=2>,S<VPBlock(3,0)><len=1>,map[(const-type string)] (const-type string)<len=2>,func(s (const-type string))  (const-type string){ ... })
//     main/files/panic0a.gno:8
// main<VPBlock(1,2)>()
//     main/files/panic0a.gno:12

// Error:
// wtf
