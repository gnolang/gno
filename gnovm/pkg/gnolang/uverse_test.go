package gnolang

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
)

type uverseTestCases struct {
	name     string
	code     string
	expected string
}

func TestIssue1337PrintNilSliceAsUndefined(t *testing.T) {
	test := []uverseTestCases{
		{
			name: "print empty slice",
			code: `package test
			func main() {
				emptySlice1 := make([]int, 0)
				emptySlice2 := []int{}

				println(emptySlice1)
				println(emptySlice2)
			}`,
			expected: "slice[]\nslice[]\n",
		},
		{
			name: "nil slice",
			code: `package test
			func main() {
				println(nil)
			}`,
			expected: "undefined\n",
		},
		{
			name: "print empty string slice",
			code: `package test
			func main() {
				var a []string
				println(a)
			}`,
			expected: "(nil []string)\n",
		},
		{
			name: "print non-empty slice",
			code: `package test
			func main() {
				a := []string{"a", "b"}
				println(a)
			}`,
			expected: "slice[(\"a\" string),(\"b\" string)]\n",
		},
		{
			name: "print empty map",
			code: `package test
			func main() {
				var a map[string]string
				println(a)
			}`,
			expected: "(nil map[string]string)\n",
		},
		{
			name: "print non-empty map",
			code: `package test
			func main() {
				a := map[string]string{"a": "b"}
				println(a)
			}`,
			expected: "map{(\"a\" string):(\"b\" string)}\n",
		},
		{
			name: "print nil struct",
			code: `package test
			func main() {
				var a struct{}
				println(a)
			}`,
			expected: "struct{}\n",
		},
		{
			name: "print function",
			code: `package test
			func foo(a, b int) int {
				return a + b
			}
			func main() {
				println(foo(1, 3))
			}`,
			expected: "4\n",
		},
		{
			name: "print composite slice",
			code: `package test
			func main() {
				a, b, c, d := 1, 2, 3, 4
				x := []int{
					a: b,
					c: d,
				}
				println(x)
			}`,
			expected: "slice[(0 int),(2 int),(0 int),(4 int)]\n",
		},
		{
			name: "simple recover case",
			code: `package test

			func main() {
				defer func() { println("recover", recover()) }()
				println("simple panic")
			}`,
			expected: "simple panic\nrecover undefined\n",
		},
		{
			name: "nested recover",
			code: `package test

			func main() {
				defer func() { println("outer recover", recover()) }()
				defer func() { println("nested panic") }()
				println("simple panic")
			}`,
			expected: "simple panic\nnested panic\nouter recover undefined\n",
		},
		{
			name: "print non-nil function",
			code: `package test
			func f() int {
				return 1
			}

			func main() {
				g := f
				println(g)
			}`,
			expected: "f\n",
		},
		{
			name: "print primitive types",
			code: `package test
			func main() {
				println(1)
				println(1.1)
				println(true)
				println("hello")
			}`,
			expected: "1\n1.1\ntrue\nhello\n",
		},
	}

	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMachine("test", nil)
			n := m.MustParseFile("main.go", tc.code)
			m.RunFiles(n)
			m.RunMain()
			assertOutput(t, tc.code, tc.expected)
		})
	}
}

var pSink any = nil

func BenchmarkGnoPrintln(b *testing.B) {
	var buf bytes.Buffer
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	m := NewMachineWithOptions(MachineOptions{
		Output: &buf,
		Store:  store,
	})

	program := `package p
				func main() {
					for i := 0; i < 1000; i++ {
						println("abcdeffffffffffffffff1222 11111   11111")
					}
				}`
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "p",
		Path: "exmaple.com/r/p",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: program},
		},
	}, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		m.RunMain()
		pSink = buf.String()
	}

	if pSink == nil {
		b.Fatal("Benchmark did not run!")
	}
	pSink = nil
}

func TestGnoPrintAndPrintln(t *testing.T) {
	tests := []struct {
		name    string
		srcArgs string
		want    string
	}{
		{
			"print with no args",
			"print()",
			"",
		},
		{
			"print with 1 arg",
			`print("1")`,
			"1",
		},
		{
			"print with 2 args",
			`print("1", 2)`,
			"1 2",
		},
		{
			"print with 3 args",
			`print("1", 2, "*")`,
			"1 2 *",
		},
		{
			"print with own spaces",
			`print("1 ", 2, "*")`,
			"1  2 *",
		},
		{
			"println with no args",
			"println()",
			"\n",
		},
		{
			"print with 1 arg",
			`println("1")`,
			"1\n",
		},
		{
			"println with 2 args",
			`println("1", 2)`,
			"1 2\n",
		},
		{
			"println with 3 args",
			`println("1", 2, "*")`,
			"1 2 *\n",
		},
		{
			"println with own spaces",
			`println("1 ", 2, "*")`,
			"1  2 *\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			db := memdb.NewMemDB()
			baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
			iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
			store := NewStore(nil, baseStore, iavlStore)

			m := NewMachineWithOptions(MachineOptions{
				Output: &buf,
				Store:  store,
			})

			program := `package p
				func main() {` + tt.srcArgs + "\n}"
			m.RunMemPackage(&std.MemPackage{
				Type: MPUserProd,
				Name: "p",
				Path: "exmaple.com/r/p",
				Files: []*std.MemFile{
					{Name: "a.gno", Body: program},
				},
			}, false)

			buf.Reset()
			m.RunMain()
			got := buf.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
