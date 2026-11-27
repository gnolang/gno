package diff

import (
	"strings"
	"testing"
)

func TestUnsupported(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "plain",
			src:  "package main\nfunc main() { println(1) }\n",
		},
		{
			name: "goto is not a goroutine",
			src:  "package main\nfunc main() {\n\tgoto L\nL:\n\tprintln(1)\n}\n",
		},
		{
			name: "keyword inside a string",
			src:  "package main\nfunc main() {\n\tprintln(\"chan select go\")\n\t// go func() {}\n}\n",
		},
		{
			name: "channel",
			src:  "package main\nfunc main() { ch := make(chan int); _ = ch }\n",
			want: "channels",
		},
		{
			name: "goroutine",
			src:  "package main\nfunc main() { go println(1) }\n",
			want: "goroutines",
		},
		{
			name: "import",
			src:  "package main\nimport \"fmt\"\nfunc main() { fmt.Println(1) }\n",
			want: "imports",
		},
		{
			name: "generic",
			src:  "package main\nfunc Id[T any](v T) T { return v }\nfunc main() { println(Id(1)) }\n",
			want: "type parameters",
		},
		{
			name: "integer range",
			src:  "package main\nfunc main() { for i := range 3 { println(i) } }\n",
			want: "integer range",
		},
		{
			name: "min builtin",
			src:  "package main\nfunc main() { println(min(1, 2)) }\n",
			want: "min builtin",
		},
		{
			name: "user min",
			src:  "package main\nfunc min(a, b int) int { if a < b { return a }; return b }\nfunc main() { println(min(1, 2)) }\n",
		},
		{
			name: "method selector named min",
			src:  "package main\ntype S struct{}\nfunc (S) min(a int) int { return a }\nfunc main() { var s S; println(s.min(1)) }\n",
		},
		{
			name: "goroutine after quote rune literal",
			src:  "package main\nfunc main() { c := '\"'; _ = c; go println(1) }\n",
			want: "goroutines",
		},
		{
			name: "channel after backtick rune literal",
			src:  "package main\nfunc main() { c := '`'; _ = c; ch := make(chan int); _ = ch }\n",
			want: "channels",
		},
		{
			name: "imaginary literal",
			src:  "package main\nfunc main() { println(1i) }\n",
			want: "complex",
		},
		{
			name: "complex builtin",
			src:  "package main\nfunc main() { println(complex(1, 2)) }\n",
			want: "complex",
		},
		{
			name: "channel arrow",
			src:  "package main\nfunc main() { <-c }\n",
			want: "channels",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Unsupported(tc.src)
			if tc.want == "" {
				if got != "" {
					t.Fatalf("Unsupported() = %q, want empty", got)
				}
				return
			}
			if got == "" || !strings.Contains(got, tc.want) {
				t.Fatalf("Unsupported() = %q, want it to mention %q", got, tc.want)
			}
		})
	}
}
