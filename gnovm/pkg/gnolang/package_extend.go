package gnolang

import (
	"fmt"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// ExtendPackage runs files new to the persisted stdlib at mpkg.Path into it, keeping every existing object with its ID and every
// existing name with its block index: realms may hold references into the
// package, by object ID, and its function values find their bodies by source
// location, so only new files, appended, leave them valid.
//
// mpkg is the package's whole new content: its stored files, unchanged, and
// the added ones. added names the added production .gno files, in the order
// to predefine them, after every file already stored; the order is recorded,
// so that every later boot predefines them the same way.
func (m *Machine) ExtendPackage(mpkg *std.MemPackage, added []string) {
	path := mpkg.Path
	pv := m.Store.GetPackage(path, false)
	if pv == nil {
		panic(fmt.Sprintf("extending package %q: not in the store", path))
	}
	if !IsStdlib(path) {
		panic(fmt.Sprintf("extending package %q: only stdlibs can be extended", path))
	}
	m.SetActivePackage(pv)
	pb := pv.GetBlock(m.Store)
	nValues, nFBlocks := len(pb.Values), len(pv.FBlocks)

	// Run the added files after the stored ones: their names take block
	// indices after every existing name (see runFileDecls).
	sub := &std.MemPackage{Name: mpkg.Name, Path: path, Type: mpkg.Type}
	for _, fname := range added {
		f := mpkg.GetFile(fname)
		if f == nil {
			panic(fmt.Sprintf("extending package %q: added file %q is not in the package", path, fname))
		}
		sub.Files = append(sub.Files, f)
	}
	fset := m.ParseMemPackageAsType(sub, sub.Type.(MemPackageType).AsRunnable())
	pv.deriveFBlocksMap(m.Store)
	updates := m.runFileDecls(false, fset.Files...)
	pv.deriveFBlocksMap(m.Store)

	// Save the new objects with IDs after the package's existing ones, and
	// the existing objects they hang from. A stdlib is saved under a throwaway
	// realm whose counter the store persists (FinalizeRealmTransaction), so
	// numbering continues from it.
	rlm := m.Store.GetPackageRealm(path)
	if rlm == nil {
		panic(fmt.Sprintf("extending package %q: no realm record to number new objects from", path))
	}
	rlm.MarkDirty(pv)
	rlm.MarkDirty(pb)
	for i := nFBlocks; i < len(pv.FBlocks); i++ {
		rlm.DidUpdate(m, pv, nil, pv.FBlocks[i].(Object))
	}
	for i := nValues; i < len(pb.Values); i++ {
		// As the realm walks a block's values (getChildObjects).
		for _, v := range getSelfOrChildObjects(pb.Values[i].V, nil) {
			if oo, ok := v.(Object); ok {
				rlm.DidUpdate(m, pb, nil, oo)
			}
		}
	}
	rlm.FinalizeRealmTransaction(m.Store)
	m.saveDeclaredTypes(path, pb.Values[nValues:])

	// Run the added files' init functions, then save what they changed.
	m.setRealm(rlm)
	m.runInitFromUpdates(pv, updates)
	rlm.FinalizeRealmTransaction(m.Store)
	m.setRealm(nil)

	m.Store.ReplaceMemPackage(mpkg, mpkg.Type.(MemPackageType))
	m.Store.AddPackageExtension(path, added)
}

// orderExtensionsLast reorders fset so that the files an extension added
// (ext, in the order they were added) come after every other file, the order
// in which ExtendPackage predefined them.
func orderExtensionsLast(fset *FileSet, ext []string) {
	slices.SortStableFunc(fset.Files, func(a, b *FileNode) int {
		ia, ib := slices.Index(ext, a.FileName), slices.Index(ext, b.FileName)
		switch {
		case ia < 0 && ib < 0:
			return 0
		case ia < 0:
			return -1
		case ib < 0:
			return 1
		default:
			return ia - ib
		}
	})
}
