package gnolang

import (
	"fmt"
	"strings"
	"testing"
)

// Calibration for the OpCPUSlopeTypeDep*, OpCPUSlopeEmbedDepthStep and
// OpCPUSlopeMapKeyNode constants: each sub-benchmark reports ns per unit of
// the work it bills, on a grid from narrow to wide shapes.

// benchPrescanFile is nDecls struct declarations with refs pointer fields
// each, pointing at other group members (edges) or at int (leaves).
func benchPrescanFile(b *testing.B, nDecls, refs int, toGroup bool) *FileNode {
	b.Helper()
	var src strings.Builder
	src.WriteString("package main\n")
	for i := range nDecls {
		fmt.Fprintf(&src, "type T%d struct {\n", i)
		for j := range refs {
			if toGroup {
				fmt.Fprintf(&src, "\tf%d *T%d\n", j, (i+j+1)%nDecls)
			} else {
				fmt.Fprintf(&src, "\tf%d *int\n", j)
			}
		}
		src.WriteString("}\n")
	}
	return NewMachine("main", nil).MustParseFile("main.gno", src.String())
}

func benchPrescan(b *testing.B, nDecls, refs int, toGroup bool) {
	b.Helper()
	fn := benchPrescanFile(b, nDecls, refs, toGroup)
	sites := appendTypeDeclSites(nil, fn, fn.Decls)
	var nodes, edges int64
	b.ResetTimer()
	for range b.N {
		g := newTypeDeclGraph(sites)
		if g.invalidCycle() != nil {
			b.Fatal("unexpected cycle")
		}
		nodes, edges = g.nodes, g.edges
	}
	b.ReportMetric(float64(nodes), "nodes")
	b.ReportMetric(float64(edges), "edges")
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(nodes+edges), "ns/unit")
}

func BenchmarkTypeDeclGroup_Prescan(b *testing.B) {
	for _, s := range []struct{ n, refs int }{{8, 1}, {64, 4}, {256, 16}, {1024, 1}, {1024, 16}} {
		b.Run(fmt.Sprintf("n%d_refs%d_leaf", s.n, s.refs), func(b *testing.B) { benchPrescan(b, s.n, s.refs, false) })
		b.Run(fmt.Sprintf("n%d_refs%d_group", s.n, s.refs), func(b *testing.B) { benchPrescan(b, s.n, s.refs, true) })
	}
}

// benchEmbedTree builds a struct embedding width distinct declared structs,
// each embedding width more, depth levels deep.
func benchEmbedTree(depth, width int) Type {
	var mk func(d int) *StructType
	mk = func(d int) *StructType {
		st := &StructType{}
		if d == 0 {
			st.Fields = []FieldType{{Name: "x", Type: IntType}}
			return st
		}
		for i := range width {
			dt := &DeclaredType{Name: Name(fmt.Sprintf("E%d_%d", d, i)), Base: mk(d - 1)}
			st.Fields = append(st.Fields, FieldType{Name: dt.Name, Type: dt, Embedded: true})
		}
		return st
	}
	return &DeclaredType{Name: "Root", Base: mk(depth)}
}

func benchEmbedDepth(b *testing.B, depth, width int) {
	b.Helper()
	t := benchEmbedTree(depth, width)
	var work int64
	b.ResetTimer()
	for range b.N {
		work = 0
		embedDepth(t, map[Type]struct{}{}, &work)
	}
	b.ReportMetric(float64(work), "work")
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(work), "ns/unit")
}

func BenchmarkTypeDeclGroup_EmbedDepth(b *testing.B) {
	for _, s := range []struct{ d, w int }{{1, 8}, {1, 128}, {2, 64}, {4, 8}, {8, 1}, {8, 2}} {
		b.Run(fmt.Sprintf("d%d_w%d", s.d, s.w), func(b *testing.B) { benchEmbedDepth(b, s.d, s.w) })
	}
}

// benchMapKeyType is a struct of n fields, each a slice of a map keyed by a
// small struct, so every node kind of the walk is exercised.
func benchMapKeyType(n int) Type {
	key := &StructType{Fields: []FieldType{{Name: "a", Type: IntType}, {Name: "b", Type: StringType}}}
	st := &StructType{}
	for i := range n {
		st.Fields = append(st.Fields, FieldType{
			Name: Name(fmt.Sprintf("f%d", i)),
			Type: &SliceType{Elt: &MapType{Key: key, Value: IntType}},
		})
	}
	return st
}

func benchMapKey(b *testing.B, n int) {
	b.Helper()
	t := benchMapKeyType(n)
	var work int64
	b.ResetTimer()
	for range b.N {
		work = 0
		if uncomparableMapKey(t, &work) != nil {
			b.Fatal("unexpected key")
		}
	}
	b.ReportMetric(float64(work), "work")
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(work), "ns/unit")
}

func BenchmarkTypeDeclGroup_MapKey(b *testing.B) {
	for _, n := range []int{1, 8, 128} {
		b.Run(fmt.Sprintf("fields%d", n), func(b *testing.B) { benchMapKey(b, n) })
	}
}
