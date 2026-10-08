package gnolang

import (
	"fmt"
	"strings"
	"testing"

	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// Gas metering of the type-declaration group walks (typecycle.go): the cycle
// pre-scan, the build-order dependency scan, and the settled-type checks
// (embed depth, map keys). Each fixture is shaped so one walk's charge
// dominates and the budget sits between the package's allocation gas and
// allocation plus that charge, so the test fails if the charge is removed.

// buildPrescanPkg: nDecls structs, each with refs pointer fields into the
// group, so the pre-scan sees nDecls sites, ~3 nodes per field and one edge
// per field.
func buildPrescanPkg(pkgName string, nDecls, refs int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkgName)
	for i := range nDecls {
		fmt.Fprintf(&b, "type T%d struct {\n", i)
		for j := range refs {
			fmt.Fprintf(&b, "\tf%d *T%d\n", j, (i+j+1)%nDecls)
		}
		b.WriteString("}\n")
	}
	b.WriteString("\nfunc main() {}\n")
	return b.String()
}

// buildEmbedDepthPkg: nRoots structs each embedding *Shared, where Shared
// embeds width leaf structs. Every root's depth check walks the whole
// shared subgraph again.
func buildEmbedDepthPkg(pkgName string, nRoots, width int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkgName)
	for i := range width {
		fmt.Fprintf(&b, "type L%d struct{ x int }\n", i)
	}
	b.WriteString("\ntype Shared struct {\n")
	for i := range width {
		fmt.Fprintf(&b, "\tL%d\n", i)
	}
	b.WriteString("}\n\n")
	for i := range nRoots {
		fmt.Fprintf(&b, "type R%d struct{ *Shared }\n", i)
	}
	b.WriteString("\nfunc main() {}\n")
	return b.String()
}

// buildMapKeyPkg: nDecls structs of fields map fields each.
func buildMapKeyPkg(pkgName string, nDecls, fields int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkgName)
	for i := range nDecls {
		fmt.Fprintf(&b, "type M%d struct {\n", i)
		for j := range fields {
			fmt.Fprintf(&b, "\tf%d map[string]int\n", j)
		}
		b.WriteString("}\n")
	}
	b.WriteString("\nfunc main() {}\n")
	return b.String()
}

// buildSitesPkg: nDecls trivial declarations, so the per-declaration
// charge dominates.
func buildSitesPkg(pkgName string, nDecls int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkgName)
	for i := range nDecls {
		fmt.Fprintf(&b, "type T%d int\n", i)
	}
	b.WriteString("\nfunc main() {}\n")
	return b.String()
}

// Budgets sit between the fixture's allocation gas (measured with the new
// slopes zeroed) and its total with them, so each test OOGs only while the
// charge it names is present (verified by mutation):
//
//	sites    2000 x `type Ti int`             978K alloc, +460K site charge
//	prescan  512 structs x 16 group pointers  1.16M alloc, +2.0M scan; edge
//	         charge alone is 450K, node charge alone 552K, so a 2.9M budget
//	         needs both
//	embed    128 roots embedding one 60-wide  98K alloc, +3.57M depth walk
//	         shared struct

func TestPreprocess_TypeDeclGroup_SiteCharged(t *testing.T) {
	panicked, val, _ := runPkgSource(t, "tdsites", 1_200_000, buildSitesPkg("tdsites", 2000))
	requireOOG(t, panicked, val)
}

func TestPreprocess_TypeDeclGroup_PrescanCharged(t *testing.T) {
	panicked, val, _ := runPkgSource(t, "tdprescan", 2_900_000, buildPrescanPkg("tdprescan", 512, 16))
	requireOOG(t, panicked, val)
}

func TestPreprocess_TypeDeclGroup_EmbedDepthCharged(t *testing.T) {
	panicked, val, _ := runPkgSource(t, "tdembed", 1_000_000, buildEmbedDepthPkg("tdembed", 128, 60))
	requireOOG(t, panicked, val)
}

// TestPreprocess_TypeDeclGroup_GasSufficient: the same shapes fit a
// generous budget, so the slopes do not reject legitimate packages.
func TestPreprocess_TypeDeclGroup_GasSufficient(t *testing.T) {
	for _, c := range []struct{ name, src string }{
		{"tdsites2", buildSitesPkg("tdsites2", 2000)},
		{"tdprescan2", buildPrescanPkg("tdprescan2", 512, 16)},
		{"tdembed2", buildEmbedDepthPkg("tdembed2", 128, 60)},
		{"tdmapkey2", buildMapKeyPkg("tdmapkey2", 200, 128)},
	} {
		panicked, val, _ := runPkgSource(t, c.name, 50_000_000, c.src)
		require.False(t, panicked, "%s: %v", c.name, val)
	}
}

// TestTypeDeclGroup_MapKeyCharged: endTypeDeclGroup bills the embed-depth
// walk per step and the map-key walk per node. The map-key work is a
// fraction of the scan charges on the same nodes, so a budget test cannot
// isolate it; this checks the exact charge on one built declaration.
func TestTypeDeclGroup_MapKeyCharged(t *testing.T) {
	m := NewMachine("main", nil)
	fn := m.MustParseFile("main.gno", "package main\n\ntype M struct {\n\ta map[string]int\n\tb [2]map[[3]string]*M\n}\n\nfunc main() {}\n")
	m.RunFiles(fn)
	td := fn.Decls[0].(*TypeDecl)
	t0 := fn.GetSlot(m.Store, td.Name, true).GetType()
	var embedWork, mapWork int64
	embedDepth(t0, map[Type]struct{}{}, &embedWork)
	require.Nil(t, uncomparableMapKey(baseOf(t0), &mapWork))
	require.Equal(t, int64(10), mapWork) // struct; a: map key value; b: array map key elem value ptr
	gm := stypes.NewGasMeter(1_000_000)
	checkBuiltTypeDecl(m.Store, gm, typeDeclSite{td, fn})
	require.Equal(t, OpCPUSlopeTypeDeclStep*embedWork+OpCPUSlopeTypeDeclNode*mapWork, gm.GasConsumed())
}
