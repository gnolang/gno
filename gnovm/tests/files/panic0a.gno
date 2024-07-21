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
// panic wtf
// f<VPBlock(3,1)>([](const-type string){(const ("a" string)), (const ("b" string))}, S<VPBlock(3,0)>{(const ("c" string))}, map[(const-type string)] (const-type string){(const ("d" string)): (const ("gg" string)), (const ("test" string)): (const ("test" string))}, func func(s (const-type string))  (const-type string){ return s<VPBlock(1,0)> })
//     main/files/panic0a.gno:8
// main()
//     main/files/panic0a.gno:12

// Error:
// wtf
