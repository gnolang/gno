package gnolang

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

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
	m.RunMemPackage(&gnovm.MemPackage{
		Name: "p",
		Path: "p",
		Files: []*gnovm.MemFile{
			{Name: "a.gno", Body: program},
		},
	}, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		m.RunStatement(S(Call(Nx("main"))))
		pSink = buf.String()
	}

	if pSink == nil {
		b.Fatal("Benchmark did not run!")
	}
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
			"",
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
			m.RunMemPackage(&gnovm.MemPackage{
				Name: "p",
				Path: "p",
				Files: []*gnovm.MemFile{
					{Name: "a.gno", Body: program},
				},
			}, false)

			buf.Reset()
			m.RunStatement(S(Call(Nx("main"))))
			got := buf.String()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("Mismatched output: got - want +\n%s", diff)
			}
		})
	}
}
