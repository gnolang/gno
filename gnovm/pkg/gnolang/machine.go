package gnolang

import (
	"fmt"
	"io"
	"path"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

//----------------------------------------
// Machine

type Machine struct {
	// State
	Ops           []Op          // main operations
	Values        []TypedValue  // buffer of values to be operated on
	Exprs         []Expr        // pending expressions
	Stmts         []Stmt        // pending statements
	Blocks        []*Block      // block (scope) stack
	Frames        []Frame       // func call stack
	Package       *PackageValue // active package
	Realm         *Realm        // active realm
	Alloc         *Allocator    // memory allocations
	Exception     *Exception    // last exception
	NumResults    int           // number of results returned
	Cycles        int64         // number of "cpu" cycles
	GCCycle       int64         // number of "gc" cycles
	Stage         Stage         // pre for static eval, add for package init, run otherwise
	ReviveEnabled bool          // true if revive() enabled (only in testing mode for now)
	Lastline      int           // the line the VM is currently executing

	Debugger Debugger

	// Configuration
	Output   io.Writer
	Store    Store
	Context  any
	GasMeter store.GasMeter
}

// NewMachine initializes a new gno virtual machine, acting as a shorthand
// for [NewMachineWithOptions], setting the given options PkgPath and Store.
//
// The machine will run on the package at the given path, which will be
// retrieved through the given store. If it is not set, the machine has no
// active package, and one must be set prior to usage.
//
// Like for [NewMachineWithOptions], Machines initialized through this
// constructor must be finalized with [Machine.Release].
func NewMachine(pkgPath string, store Store) *Machine {
	return NewMachineWithOptions(
		MachineOptions{
			PkgPath: pkgPath,
			Store:   store,
		})
}

// MachineOptions is used to pass options to [NewMachineWithOptions].
type MachineOptions struct {
	// Active package of the given machine; must be set before execution.
	PkgPath       string
	Debug         bool
	Input         io.Reader // used for default debugger input only
	Output        io.Writer // default os.Stdout
	Store         Store     // default NewStore(Alloc, nil, nil)
	Context       any
	Alloc         *Allocator // or see MaxAllocBytes.
	MaxAllocBytes int64      // or 0 for no limit.
	GasMeter      store.GasMeter
	ReviveEnabled bool
	SkipPackage   bool // don't get/set package or realm.
}

const (
	startingOpsCap = 1024
	// sizeof(TypedValue) is 40 at time of writing; this ensures that the values
	// slice occupies 1000 bytes by default.
	startingValuesCap = 25
)

// the machine constructor gets spammed
// this causes a significant part of the runtime and memory
// to be occupied by *Machine
// hence, this pool
var machinePool = sync.Pool{
	New: func() any {
		return &Machine{
			Ops:    make([]Op, 0, startingOpsCap),
			Values: make([]TypedValue, 0, startingValuesCap),
		}
	},
}

// NewMachineWithOptions initializes a new gno virtual machine with the given
// options.
//
// Machines initialized through this constructor must be finalized with
// [Machine.Release].
func NewMachineWithOptions(opts MachineOptions) *Machine {
	vmGasMeter := opts.GasMeter

	output := opts.Output
	if output == nil {
		output = io.Discard
	}
	alloc := opts.Alloc
	if alloc == nil {
		alloc = NewAllocator(opts.MaxAllocBytes) // allocator is nil if MaxAllocBytes is zero
	}
	store := opts.Store
	if store == nil {
		// bare store, no stdlibs.
		store = NewStore(alloc, nil, nil)
	} else if store.GetAllocator() == nil {
		store.SetAllocator(alloc)
	}
	// Get machine from pool.
	mm := machinePool.Get().(*Machine)
	mm.Alloc = alloc
	if mm.Alloc != nil {
		mm.Alloc.SetGCFn(func() (int64, bool) { return mm.GarbageCollect() })
		mm.Alloc.SetGasMeter(vmGasMeter)
	}
	mm.Output = output
	mm.Store = store
	mm.Context = opts.Context
	mm.GasMeter = vmGasMeter
	mm.Debugger.enabled = opts.Debug
	mm.Debugger.in = opts.Input
	mm.Debugger.out = output
	mm.ReviveEnabled = opts.ReviveEnabled
	// Maybe get/set package and realm.
	if !opts.SkipPackage && opts.PkgPath != "" {
		pv := (*PackageValue)(nil)
		pv = store.GetPackage(opts.PkgPath, false)
		if pv == nil {
			pkgName := defaultPkgName(opts.PkgPath)
			pn := NewPackageNode(pkgName, opts.PkgPath, &FileSet{})
			pv = pn.NewPackage(mm.Alloc)
			store.SetBlockNode(pn)
			store.SetCachePackage(pv)
		}
		mm.Package = pv
		if pv != nil {
			mm.SetActivePackage(pv)
		}
	}
	return mm
}

// Release resets some of the values of *Machine and puts back m into the
// machine pool; for this reason, Release() should be called as a finalizer,
// and m should not be used after this call. Only Machines initialized with this
// package's constructors should be released.
func (m *Machine) Release() {
	// here we zero in the values for the next user
	ops, values := m.Ops[:0:startingOpsCap], m.Values[:0:startingValuesCap]
	clear(ops[:startingOpsCap])
	clear(values[:startingValuesCap])
	*m = Machine{Ops: ops, Values: values}

	machinePool.Put(m)
}

func (m *Machine) SetActivePackage(pv *PackageValue) {
	if err := m.CheckEmpty(); err != nil {
		panic(errors.Wrap(err, "set package when machine not empty"))
	}
	m.Package = pv
	m.Realm = pv.GetRealm()
	m.Blocks = []*Block{
		pv.GetBlock(m.Store),
	}
}

//----------------------------------------
// top level Run* methods.

// Upon restart, preprocess all MemPackage and save blocknodes.
// This is a temporary measure until we optimize/make-lazy.
//
// NOTE: package paths not beginning with gno.land will be allowed to override,
// to support cases of stdlibs processed through [RunMemPackagesWithOverrides].
func (m *Machine) PreprocessAllFilesAndSaveBlockNodes() {
	ch := m.Store.IterMemPackage()
	for mpkg := range ch {
		mpkg = MPFProd.FilterMemPackage(mpkg)
		fset := m.ParseMemPackage(mpkg)
		pn := NewPackageNode(Name(mpkg.Name), mpkg.Path, fset)
		m.Store.SetBlockNode(pn)
		PredefineFileSet(m.Store, pn, fset)
		for _, fn := range fset.Files {
			// Save Types to m.Store (while preprocessing).
			fn = Preprocess(m.Store, pn, fn).(*FileNode)
			// Save BlockNodes to m.Store.
			SaveBlockNodes(m.Store, fn)
		}
		// Normally, the fileset would be added onto the
		// package node only after runFiles(), but we cannot
		// run files upon restart (only preprocess them).
		// So, add them here instead.
		// TODO: is this right?
		if pn.FileSet == nil {
			pn.FileSet = fset
		}
		// pn.FileSet != nil happens for non-realm file tests.
		// TODO ensure the files are the same.
	}
}

//----------------------------------------
// top level Run* methods.

// Sorts the package, then sets the package if doesn't exist, runs files, saves
// mpkg and corresponding package node, package value, and types to store. Save
// is set to false for tests where package values may be native.
// If save is true, mpkg must be of type.IsStorable().
// NOTE: Production systems must separately check mpkg.type if save, typically
// you will want to ensure that it is MPUserAll, not MPUserProd or MPUserTest.
// NOTE: Does not validate the mpkg. Caller must validate the mpkg before
// calling.
func (m *Machine) RunMemPackage(mpkg *std.MemPackage, save bool) (*PackageNode, *PackageValue) {
	if bm.OpsEnabled || bm.StorageEnabled || bm.NativeEnabled {
		bm.InitMeasure()
	}
	if bm.StorageEnabled {
		defer bm.FinishStore()
	}
	return m.runMemPackage(mpkg, save, false)
}

// RunMemPackageWithOverrides works as [RunMemPackage], however after parsing,
// declarations are filtered removing duplicate declarations.  To control which
// declaration overrides which, use [ReadMemPackageFromList], putting the
// overrides at the top of the list.
// If save is true, mpkg must be of type.IsStorable().
// NOTE: Production systems must separately check mpkg.type if save, typically
// you will want to ensure that it is MPUserAll, not MPUserProd or MPUserTest.
// NOTE: Does not validate the mpkg, except when saving validates a mpkg with
// its type.
func (m *Machine) RunMemPackageWithOverrides(mpkg *std.MemPackage, save bool) (*PackageNode, *PackageValue) {
	return m.runMemPackage(mpkg, save, true)
}

func (m *Machine) runMemPackage(mpkg *std.MemPackage, save, overrides bool) (*PackageNode, *PackageValue) {
	// validate mpkg.Type.
	mptype := mpkg.Type.(MemPackageType)
	if save && !mptype.IsStorable() {
		panic(fmt.Sprintf("mempackage type must be storable, but got %v", mptype))
	}
	// If All, demote to Prod when parsing,
	// if Test or Integration, keep it as is,
	// but in any case save everything if save.
	mptype = mptype.AsRunnable()
	// sort mpkg.
	mpkg.Sort()
	// parse files.
	files := m.ParseMemPackageAsType(mpkg, mptype)
	mod, err := gnomod.ParseMemPackage(mpkg)
	private := false
	if err == nil && mod != nil {
		private = mod.Private
	}

	// make and set package if doesn't exist.
	pn := (*PackageNode)(nil)
	pv := (*PackageValue)(nil)
	if m.Package != nil && m.Package.PkgPath == mpkg.Path {
		pv = m.Package
		loc := PackageNodeLocation(mpkg.Path)
		pn = m.Store.GetBlockNode(loc).(*PackageNode)
	} else {
		pn = NewPackageNode(Name(mpkg.Name), mpkg.Path, &FileSet{})
		pv = pn.NewPackage(m.Alloc)
		pv.SetPrivate(private)
		m.Store.SetBlockNode(pn)
		m.Store.SetCachePackage(pv)
	}
	m.SetActivePackage(pv)
	// run files.
	updates := m.runFileDecls(overrides, files.Files...)
	// populate pv.fBlocksMap.
	pv.deriveFBlocksMap(m.Store)
	// save package value and mempackage.
	// XXX save condition will be removed once gonative is removed.
	var throwaway *Realm
	if save {
		// store new package values and types
		throwaway = m.saveNewPackageValuesAndTypes()
		if throwaway != nil {
			m.Realm = throwaway
		}
	}
	// run init functions
	m.runInitFromUpdates(pv, updates)
	// save again after init.
	if save {
		m.resavePackageValues(throwaway)
		// store mempackage; we already validated type.
		m.Store.AddMemPackage(mpkg, mpkg.Type.(MemPackageType))
		if throwaway != nil {
			m.Realm = nil
		}
	}

	return pn, pv
}

type redeclarationErrors []Name

func (r redeclarationErrors) Error() string {
	var b strings.Builder
	b.WriteString("redeclarations for identifiers: ")
	for idx, s := range r {
		b.WriteString(strconv.Quote(string(s)))
		if idx != len(r)-1 {
			b.WriteString(", ")
		}
	}
	return b.String()
}

func (r redeclarationErrors) add(newI Name) redeclarationErrors {
	if slices.Contains(r, newI) {
		return r
	}
	return append(r, newI)
}

// checkDuplicates returns an error if there are duplicate declarations in the fset.
func checkDuplicates(fset *FileSet) error {
	defined := make(map[Name]struct{}, 128)
	var duplicated redeclarationErrors
	for _, f := range fset.Files {
		for _, d := range f.Decls {
			var name Name
			switch d := d.(type) {
			case *FuncDecl:
				if d.Name == "init" {
					continue
				}
				name = d.Name
				if d.IsMethod {
					name = Name(destar(d.Recv.Type).String()) + "." + name
				}
			case *TypeDecl:
				name = d.Name
			case *ValueDecl:
				for _, nx := range d.NameExprs {
					if nx.Name == blankIdentifier {
						continue
					}
					if _, ok := defined[nx.Name]; ok {
						duplicated = duplicated.add(nx.Name)
					}
					defined[nx.Name] = struct{}{}
				}
				continue
			default:
				continue
			}
			if name == blankIdentifier {
				continue
			}
			if _, ok := defined[name]; ok {
				duplicated = duplicated.add(name)
			}
			defined[name] = struct{}{}
		}
	}
	if len(duplicated) > 0 {
		return duplicated
	}
	return nil
}

func destar(x Expr) Expr {
	if x, ok := x.(*StarExpr); ok {
		return x.X
	}
	return x
}

// Stacktrace returns the stack trace of the machine.
// It collects the executions and frames from the machine's frames and statements.
func (m *Machine) Stacktrace() (stacktrace Stacktrace) {
	if len(m.Frames) == 0 {
		return
	}

	calls := make([]StacktraceCall, 0, len(m.Frames))
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if fr.IsCall() && fr.Func.Name != "panic" {
			calls = append(calls, StacktraceCall{
				CallExpr: fr.Source.(*CallExpr),
				IsDefer:  fr.IsDefer,
				FuncLoc:  fr.Func.GetSource(m.Store).GetLocation(),
			})
		}
	}

	// if the stacktrace is too long, we trim it down to maxStacktraceSize
	if len(calls) > maxStacktraceSize {
		const halfMax = maxStacktraceSize / 2

		stacktrace.NumFramesElided = len(calls) - maxStacktraceSize
		calls = append(calls[:halfMax], calls[len(calls)-halfMax:]...)
		calls = calls[:len(calls):len(calls)] // makes remaining part of "calls" GC'able
	}

	stacktrace.Calls = calls

	if m.LastFrame().Func != nil && m.LastFrame().Func.IsNative() {
		stacktrace.LastLine = -1 // special line for native.
	} else {
		if m.Lastline != 0 {
			stacktrace.LastLine = m.Lastline
			return
		}

		ls := m.PeekStmt(1)
		if bs, ok := ls.(*bodyStmt); ok {
			stacktrace.LastLine = bs.LastStmt().GetLine()
			return
		}
	}
	return
}

// Convenience for tests.
// Production must not use this, because realm package init
// must happen after persistence and realm finalization,
// then changes from init persisted again.
// m.Package must match fns's package path.
// XXX delete?
func (m *Machine) RunFiles(fns ...*FileNode) {
	pv := m.Package
	if pv == nil {
		panic("RunFiles requires Machine.Package")
	}
	rlm := pv.GetRealm()
	if rlm == nil && pv.IsRealm() {
		rlm = NewRealm(pv.PkgPath) // throwaway
	}
	updates := m.runFileDecls(IsStdlib(pv.PkgPath), fns...)
	if rlm != nil {
		pb := pv.GetBlock(m.Store)
		for _, update := range updates {
			// XXX simplify.
			if hiv, ok := update.V.(*HeapItemValue); ok {
				rlm.DidUpdate(pb, nil, hiv)
			} else {
				rlm.DidUpdate(pb, nil, update.GetFirstObject(m.Store))
			}
		}
	}
	m.runInitFromUpdates(pv, updates)
	if rlm != nil {
		rlm.FinalizeRealmTransaction(m.Store)
	}
}

// PreprocessFiles runs Preprocess on the given files. It is used to detect
// compile-time errors in the package. It is also used to preprocess files from
// the package getter for tests, e.g. from "gnovm/tests/files/extern/*", or from
// "examples/*".
//   - fixFrom: the version of gno to fix from.
func (m *Machine) PreprocessFiles(pkgName, pkgPath string, fset *FileSet, save, withOverrides bool, fixFrom string) (*PackageNode, *PackageValue) {
	if !withOverrides {
		if err := checkDuplicates(fset); err != nil {
			panic(fmt.Errorf("running package %q: %w", pkgName, err))
		}
	}
	pn := NewPackageNode(Name(pkgName), pkgPath, fset)
	if fixFrom != "" {
		pn.SetAttribute(ATTR_FIX_FROM, fixFrom)
	}
	pv := pn.NewPackage(nilAllocator)
	pb := pv.GetBlock(m.Store)
	m.SetActivePackage(pv)
	m.Store.SetBlockNode(pn)
	PredefineFileSet(m.Store, pn, fset)
	for _, fn := range fset.Files {
		fn = Preprocess(m.Store, pn, fn).(*FileNode)
		// After preprocessing, save blocknodes to store.
		SaveBlockNodes(m.Store, fn)
		// Make block for fn.
		// Each file for each *PackageValue gets its own file *Block,
		// with values copied over from each file's
		// *FileNode.StaticBlock.
		fb := m.Alloc.NewBlock(fn, pb)
		fb.Values = make([]TypedValue, len(fn.StaticBlock.Values))
		copy(fb.Values, fn.StaticBlock.Values)
		pv.AddFileBlock(fn.FileName, fb)
	}
	// Get new values across all files in package.
	pn.PrepareNewValues(nilAllocator, pv)
	// save package value.
	var throwaway *Realm
	if save {
		// store new package values and types
		throwaway = m.saveNewPackageValuesAndTypes()
		if throwaway != nil {
			m.Realm = throwaway
		}
		m.resavePackageValues(throwaway)
		if throwaway != nil {
			m.Realm = nil
		}
	}
	return pn, pv
}

// Add files to the package's *FileSet and run decls in them.
// This will also run each init function encountered.
// Returns the updated typed values of package.
// m.Package must match fns's package path.
func (m *Machine) runFileDecls(withOverrides bool, fns ...*FileNode) []TypedValue {
	// Files' package names must match the machine's active one.
	// if there is one.
	for _, fn := range fns {
		if fn.PkgName != "" && fn.PkgName != m.Package.PkgName {
			panic(fmt.Sprintf("expected package name [%s] but got [%s]!",
				m.Package.PkgName, fn.PkgName))
		}
	}
	// Add files to *PackageNode.FileSet.
	pv := m.Package
	pb := pv.GetBlock(m.Store)
	pn := pb.GetSource(m.Store).(*PackageNode)
	fs := &FileSet{Files: fns}
	fdeclared := map[Name]struct{}{}
	if pn.FileSet == nil {
		pn.FileSet = fs
	} else {
		// collect pre-existing declared names
		for _, fn := range pn.FileSet.Files {
			for _, decl := range fn.Decls {
				for _, name := range decl.GetDeclNames() {
					fdeclared[name] = struct{}{}
				}
			}
		}
		// add fns to pre-existing fileset.
		pn.FileSet.AddFiles(fns...)
	}
	if !withOverrides {
		if err := checkDuplicates(pn.FileSet); err != nil {
			panic(fmt.Errorf("running package %q: %w", pv.PkgPath, err))
		}
	}

	// Predefine declarations across all files.
	PredefineFileSet(m.Store, pn, fs)

	// Preprocess each new file.
	for _, fn := range fns {
		// Preprocess file.
		// NOTE: Most of the declaration is handled by
		// Preprocess and any constant values set on
		// pn.StaticBlock, and those values are copied to the
		// runtime package value via PrepareNewValues.  Then,
		// non-constant var declarations and file-level imports
		// are re-set in runDeclaration(,true).
		fn = Preprocess(m.Store, pn, fn).(*FileNode)
		if debug {
			debug.Printf("PREPROCESSED FILE: %v\n", fn)
		}
		// After preprocessing, save blocknodes to store.
		SaveBlockNodes(m.Store, fn)
		// Make block for fn.
		// Each file for each *PackageValue gets its own file *Block,
		// with values copied over from each file's
		// *FileNode.StaticBlock.
		fb := m.Alloc.NewBlock(fn, pb)
		fb.Values = make([]TypedValue, len(fn.StaticBlock.Values))
		copy(fb.Values, fn.StaticBlock.Values)
		pv.AddFileBlock(fn.FileName, fb)
	}

	// Get new values across all files in package.
	updates := pn.PrepareNewValues(m.Alloc, pv)

	// to detect loops in var declarations.
	loopfindr := []Name{}
	// recursive function for var declarations.
	var runDeclarationFor func(fn *FileNode, decl Decl)
	runDeclarationFor = func(fn *FileNode, decl Decl) {
		// get fileblock of fn.
		// fb := pv.GetFileBlock(nil, fn.FileName)
		// get dependencies of decl.
		deps := make(map[Name]struct{})
		findDependentNames(decl, deps)
		for dep := range deps {
			// if dep already defined as import, skip.
			if _, ok := fn.GetLocalIndex(dep); ok {
				continue
			}
			// if dep already in fdeclared, skip.
			if _, ok := fdeclared[dep]; ok {
				continue
			}
			fn, depdecl, exists := pn.FileSet.GetDeclForSafe(dep)
			// special case: if doesn't exist:
			if !exists {
				if isUverseName(dep) { // then is reserved keyword in uverse.
					continue
				} else { // is an undefined dependency.
					panic(fmt.Sprintf(
						"%s/%s:%s: dependency %s not defined in fileset with files %v",
						pv.PkgPath, fn.FileName, decl.GetPos().String(), dep, fs.FileNames()))
				}
			}
			// if dep already in loopfindr, abort.
			if slices.Contains(loopfindr, dep) {
				if _, ok := (*depdecl).(*FuncDecl); ok {
					// recursive function dependencies
					// are OK with func decls.
					continue
				} else {
					panic(fmt.Sprintf(
						"%s/%s:%s: loop in variable initialization: dependency trail %v circularly depends on %s",
						pv.PkgPath, fn.FileName, decl.GetPos().String(), loopfindr, dep))
				}
			}
			// run dependency declaration
			loopfindr = append(loopfindr, dep)
			runDeclarationFor(fn, *depdecl)
			loopfindr = loopfindr[:len(loopfindr)-1]
		}
		// run declaration
		fb := pv.GetFileBlock(m.Store, fn.FileName)
		m.PushBlock(fb)
		m.runDeclaration(decl)
		m.PopBlock()
		for _, n := range decl.GetDeclNames() {
			fdeclared[n] = struct{}{}
		}
	}

	// Declarations (and variable initializations).  This must happen
	// after all files are preprocessed, because value decl may be out of
	// order and depend on other files.

	// Run declarations.
	for _, fn := range fns {
		for _, decl := range fn.Decls {
			runDeclarationFor(fn, decl)
		}
	}

	return updates
}

// Run new init functions.
// Go spec: "To ensure reproducible initialization
// behavior, build systems are encouraged to present
// multiple files belonging to the same package in
// lexical file name order to a compiler."
// If m.Realm is set `init(cur realm)` works too.
func (m *Machine) runInitFromUpdates(pv *PackageValue, updates []TypedValue) {
	// Only for the init functions make the origin caller
	// the package addr.
	for _, tv := range updates {
		if tv.IsDefined() && tv.T.Kind() == FuncKind && tv.V != nil {
			fv, ok := tv.V.(*FuncValue)
			if !ok {
				continue // skip native functions.
			}
			if strings.HasPrefix(string(fv.Name), "init.") {
				fb := pv.GetFileBlock(m.Store, fv.FileName)
				m.PushBlock(fb)
				maybeCrossing := m.Realm != nil
				m.runFunc(StageAdd, fv.Name, maybeCrossing)
				m.PopBlock()
			}
		}
	}
}

// Save the machine's package using realm finalization deep crawl.
// Also saves declared types.
// This happens before any init calls.
// Returns a throwaway realm package is not a realm,
// such as stdlibs or /p/ packages.
func (m *Machine) saveNewPackageValuesAndTypes() (throwaway *Realm) {
	// save package value and dependencies.
	pv := m.Package
	if pv.IsRealm() {
		rlm := pv.Realm
		rlm.MarkNewReal(pv)
		rlm.FinalizeRealmTransaction(m.Store)
		// save package realm info.
		m.Store.SetPackageRealm(rlm)
	} else { // use a throwaway realm.
		rlm := NewRealm(pv.PkgPath)
		rlm.MarkNewReal(pv)
		rlm.FinalizeRealmTransaction(m.Store)
		throwaway = rlm
	}
	// save declared types.
	if bv, ok := pv.Block.(*Block); ok {
		for _, tv := range bv.Values {
			if tvv, ok := tv.V.(TypeValue); ok {
				if dt, ok := tvv.Type.(*DeclaredType); ok {
					m.Store.SetType(dt)
				}
			}
		}
	}
	return
}

// Resave any changes to realm after init calls.
// Pass in the realm from m.saveNewPackageValuesAndTypes()
// in case a throwaway was created.
func (m *Machine) resavePackageValues(rlm *Realm) {
	// save package value and dependencies.
	pv := m.Package
	if pv.IsRealm() {
		rlm = pv.Realm
		rlm.FinalizeRealmTransaction(m.Store)
		// re-save package realm info.
		m.Store.SetPackageRealm(rlm)
	} else { // use the throwaway realm.
		rlm.FinalizeRealmTransaction(m.Store)
	}
	// types were already saved, and should not change
	// even after running the init function.
}

func (m *Machine) runFunc(st Stage, fn Name, maybeCrossing bool) {
	if maybeCrossing {
		pv := m.Package
		pb := pv.GetBlock(m.Store)
		pn := pb.GetSource(m.Store).(*PackageNode)
		ft := pn.GetStaticTypeOf(m.Store, fn).(*FuncType)
		if ft.IsCrossing() {
			// .cur is a special keyword for non-crossing calls of
			// a crossing function where `cur` is not available
			// from m.RunFuncMaybeCrossing().
			//
			// `main(cur realm)` and `init(cur realm)` are
			// considered to have already crossed at "frame -1", so
			// we do not want to cross-call main, and the behavior
			// is identical to main(), like wise init().
			m.RunStatement(st, S(Call(Nx(fn), Nx(".cur"))))
			return
		}
	}
	m.RunStatement(st, S(Call(Nx(fn))))
}

func (m *Machine) RunMain() {
	m.runFunc(StageRun, "main", false)
}

// This is used for realm filetests which may declare
// either main() or main(cur crossing).
func (m *Machine) RunMainMaybeCrossing() {
	m.runFunc(StageRun, "main", true)
}

// Evaluate throwaway expression in new block scope.
// If x is a function call, it may return any number of
// results including 0.  Otherwise it returns 1.
// Input must not have been preprocessed, that is,
// it should not be the child of any parent.
func (m *Machine) Eval(x Expr) []TypedValue {
	if debug {
		m.Printf("Machine.Eval(%v)\n", x)
	}
	if bm.OpsEnabled || bm.StorageEnabled {
		// reset the benchmark
		bm.InitMeasure()
	}
	if bm.StorageEnabled {
		defer bm.FinishStore()
	}
	// X must not have been preprocessed.
	if x.GetAttribute(ATTR_PREPROCESSED) != nil {
		panic(fmt.Sprintf(
			"Machine.Eval(x) expression already preprocessed: %s",
			x.String()))
	}
	// Preprocess input using last block context.
	last := m.LastBlock().GetSource(m.Store)
	// Transform expression to ensure isolation.
	// This is to ensure that the parent context
	// doesn't get modified.
	// XXX Just use a BlockStmt?
	if _, ok := x.(*CallExpr); !ok {
		x = Call(Fn(nil, Flds("x", InterfaceT(nil)),
			Ss(
				Return(x),
			)))
	}
	// else,x already creates its own scope.
	// Preprocess x.
	x = Preprocess(m.Store, last, x).(Expr)
	// Evaluate x.
	start := len(m.Values)
	m.PushOp(OpHalt)
	m.PushExpr(x)
	m.PushOp(OpEval)
	m.Run(StageRun)
	res := m.ReapValues(start)
	return res
}

// Evaluate any preprocessed expression statically.
// This is primiarily used by the preprocessor to evaluate
// static types and values.
func (m *Machine) EvalStatic(last BlockNode, x Expr) TypedValue {
	if debug {
		m.Printf("Machine.EvalStatic(%v, %v)\n", last, x)
	}
	// X must have been preprocessed.
	if x.GetAttribute(ATTR_PREPROCESSED) == nil {
		panic(fmt.Sprintf(
			"Machine.EvalStatic(x) expression not yet preprocessed: %s",
			x.String()))
	}
	// Temporarily push last to m.Blocks.
	m.PushBlock(last.GetStaticBlock().GetBlock())
	// Evaluate x.
	start := len(m.Values)
	m.PushOp(OpHalt)
	m.PushOp(OpPopBlock)
	m.PushExpr(x)
	m.PushOp(OpEval)
	m.Run(StagePre)
	res := m.ReapValues(start)
	if len(res) != 1 {
		panic("should not happen")
	}
	return res[0]
}

// Evaluate the type of any preprocessed expression statically.
// This is primiarily used by the preprocessor to evaluate
// static types of nodes.
func (m *Machine) EvalStaticTypeOf(last BlockNode, x Expr) Type {
	if debug {
		m.Printf("Machine.EvalStaticTypeOf(%v, %v)\n", last, x)
	}
	// X must have been preprocessed or a predefined func lit expr.
	if x.GetAttribute(ATTR_PREPROCESSED) == nil &&
		x.GetAttribute(ATTR_PREPROCESS_SKIPPED) == nil &&
		x.GetAttribute(ATTR_PREPROCESS_INCOMPLETE) == nil {
		panic(fmt.Sprintf(
			"Machine.EvalStaticTypeOf(x) expression not yet preprocessed: %s",
			x.String()))
	}
	// Temporarily push last to m.Blocks.
	m.PushBlock(last.GetStaticBlock().GetBlock())
	// Evaluate x.
	start := len(m.Values)
	m.PushOp(OpHalt)
	m.PushOp(OpPopBlock)
	m.PushExpr(x)
	m.PushOp(OpStaticTypeOf)
	m.Run(StagePre)
	res := m.ReapValues(start)
	if len(res) != 1 {
		panic("should not happen")
	}
	tv := res[0].V.(TypeValue)
	return tv.Type
}

// Runs a statement on a block. The block must not be a package node's block,
// but it may be a file block or anything else.  New names may be declared by
// the statement, so the block is expanded with its own source.
func (m *Machine) RunStatement(st Stage, s Stmt) {
	lb := m.LastBlock()
	last := lb.GetSource(m.Store)
	switch last.(type) {
	case *FileNode, *PackageNode:
		// NOTE: type decls and value decls are also statements, and
		// they add a name to m.LastBlock, except if last block is a
		// file/package block it adds to the parent package block.
		if d, ok := s.(Decl); ok {
			m.RunDeclaration(d)
			return // already pn.PrepareNewValues()'d.
		}
	}
	// preprocess s and expand last if needed.
	func() {
		oldNames := last.GetNumNames()
		defer func() {
			// if preprocess panics during `a := ...`,
			// the static block will have a new slot but not
			// the runtime block, causing issues later.
			newNames := last.GetNumNames()
			if oldNames != newNames {
				lb.ExpandWith(m.Alloc, last)
			}
		}()
		s = Preprocess(m.Store, last, s).(Stmt)
	}()
	// run s.
	m.PushOp(OpHalt)
	m.PushStmt(s)
	m.PushOp(OpExec)
	m.Run(st)
}

// Runs a declaration after preprocessing d.  If d was already preprocessed,
// call runDeclaration() instead.  No blocknodes are saved to store, and
// declarations are not realm compatible.
func (m *Machine) RunDeclaration(d Decl) {
	if fd, ok := d.(*FuncDecl); ok && fd.Name == "init" {
		// XXX or, consider running it, but why would this be needed?
		// from a repl there is no need for init() functions.
		// Also, there are complications with realms, where
		// the realm must be persisted before init(), and persisted again.
		panic("Machine.RunDeclaration cannot be used for init functions")
	}
	// Preprocess input using package block.  There should only
	// be one block right now, and it's a *PackageNode.
	pn := m.LastBlock().GetSource(m.Store).(*PackageNode)
	d = Preprocess(m.Store, pn, d).(Decl)
	// do not SaveBlockNodes(m.Store, d).
	pn.PrepareNewValues(m.Alloc, m.Package)
	m.runDeclaration(d)
	if debug {
		if pn != m.Package.GetBlock(m.Store).GetSource(m.Store) {
			panic("package mismatch")
		}
	}
}

// Declarations to be run within a body (not at the file or
// package level, for which evaluations happen during
// preprocessing).
func (m *Machine) runDeclaration(d Decl) {
	switch d := d.(type) {
	case *FuncDecl:
		// nothing to do.
		// closure and package already set
		// during PackageNode.NewPackage().
	case *ValueDecl:
		m.PushOp(OpHalt)
		m.PushStmt(d)
		m.PushOp(OpExec)
		m.Run(StageAdd)
	case *TypeDecl:
		m.PushOp(OpHalt)
		m.PushStmt(d)
		m.PushOp(OpExec)
		m.Run(StageAdd)
	default:
		// Do nothing for package constants.
	}
}

//----------------------------------------
// Op

type Op uint8

const (

	/* Control operators */
	OpInvalid             Op = 0x00 // invalid
	OpHalt                Op = 0x01 // halt (e.g. last statement)
	OpNoop                Op = 0x02 // no-op
	OpExec                Op = 0x03 // exec next statement
	OpPrecall             Op = 0x04 // sets X (func) to frame
	OpEnterCrossing       Op = 0x05 // before OpCall of a crossing function
	OpCall                Op = 0x06 // call(Frame.Func, [...])
	OpCallNativeBody      Op = 0x07 // call body is native
	OpDefer               Op = 0x0A // defer call(X, [...])
	OpCallDeferNativeBody Op = 0x0B // call body is native
	OpGo                  Op = 0x0C // go call(X, [...])
	OpSelect              Op = 0x0D // exec next select case
	OpSwitchClause        Op = 0x0E // exec next switch clause
	OpSwitchClauseCase    Op = 0x0F // exec next switch clause case
	OpTypeSwitch          Op = 0x10 // exec type switch clauses (all)
	OpIfCond              Op = 0x11 // eval cond
	OpPopValue            Op = 0x12 // pop X
	OpPopResults          Op = 0x13 // pop n call results
	OpPopBlock            Op = 0x14 // pop block NOTE breaks certain invariants.
	OpPopFrameAndReset    Op = 0x15 // pop frame and reset.
	OpPanic1              Op = 0x16 // pop exception and pop call frames. XXX DEPRECATED
	OpPanic2              Op = 0x17 // pop call frames.
	OpReturn              Op = 0x1A // return ...
	OpReturnAfterCopy     Op = 0x1B // return ... (with named results)
	OpReturnFromBlock     Op = 0x1C // return results (after defers)
	OpReturnToBlock       Op = 0x1D // copy results to block (before defer) XXX rename to OpCopyResultsToBlock

	/* Unary & binary operators */
	OpUpos  Op = 0x20 // + (unary)
	OpUneg  Op = 0x21 // - (unary)
	OpUnot  Op = 0x22 // ! (unary)
	OpUxor  Op = 0x23 // ^ (unary)
	OpUrecv Op = 0x25 // <- (unary) // TODO make expr
	OpLor   Op = 0x26 // ||
	OpLand  Op = 0x27 // &&
	OpEql   Op = 0x28 // ==
	OpNeq   Op = 0x29 // !=
	OpLss   Op = 0x2A // <
	OpLeq   Op = 0x2B // <=
	OpGtr   Op = 0x2C // >
	OpGeq   Op = 0x2D // >=
	OpAdd   Op = 0x2E // +
	OpSub   Op = 0x2F // -
	OpBor   Op = 0x30 // |
	OpXor   Op = 0x31 // ^
	OpMul   Op = 0x32 // *
	OpQuo   Op = 0x33 // /
	OpRem   Op = 0x34 // %
	OpShl   Op = 0x35 // <<
	OpShr   Op = 0x36 // >>
	OpBand  Op = 0x37 // &
	OpBandn Op = 0x38 // &^

	/* Other expression operators */
	OpEval         Op = 0x40 // eval next expression
	OpBinary1      Op = 0x41 // X op ?
	OpIndex1       Op = 0x42 // X[Y]
	OpIndex2       Op = 0x43 // (_, ok :=) X[Y]
	OpSelector     Op = 0x44 // X.Y
	OpSlice        Op = 0x45 // X[Low:High:Max]
	OpStar         Op = 0x46 // *X (deref or pointer-to)
	OpRef          Op = 0x47 // &X
	OpTypeAssert1  Op = 0x48 // X.(Type)
	OpTypeAssert2  Op = 0x49 // (_, ok :=) X.(Type)
	OpStaticTypeOf Op = 0x4A // static type of X
	OpCompositeLit Op = 0x4B // X{???}
	OpArrayLit     Op = 0x4C // [Len]{...}
	OpSliceLit     Op = 0x4D // []{value,...}
	OpSliceLit2    Op = 0x4E // []{key:value,...}
	OpMapLit       Op = 0x4F // X{...}
	OpStructLit    Op = 0x50 // X{...}
	OpFuncLit      Op = 0x51 // func(T){Body}
	OpConvert      Op = 0x52 // Y(X)

	/* Type operators */
	OpFieldType     Op = 0x70 // Name: X `tag`
	OpArrayType     Op = 0x71 // [X]Y{}
	OpSliceType     Op = 0x72 // []X{}
	OpPointerType   Op = 0x73 // *X
	OpInterfaceType Op = 0x74 // interface{...}
	OpChanType      Op = 0x75 // [<-]chan[<-]X
	OpFuncType      Op = 0x76 // func(params...)results...
	OpMapType       Op = 0x77 // map[X]Y
	OpStructType    Op = 0x78 // struct{...}

	/* Statement operators */
	OpAssign      Op = 0x80 // Lhs = Rhs
	OpAddAssign   Op = 0x81 // Lhs += Rhs
	OpSubAssign   Op = 0x82 // Lhs -= Rhs
	OpMulAssign   Op = 0x83 // Lhs *= Rhs
	OpQuoAssign   Op = 0x84 // Lhs /= Rhs
	OpRemAssign   Op = 0x85 // Lhs %= Rhs
	OpBandAssign  Op = 0x86 // Lhs &= Rhs
	OpBandnAssign Op = 0x87 // Lhs &^= Rhs
	OpBorAssign   Op = 0x88 // Lhs |= Rhs
	OpXorAssign   Op = 0x89 // Lhs ^= Rhs
	OpShlAssign   Op = 0x8A // Lhs <<= Rhs
	OpShrAssign   Op = 0x8B // Lhs >>= Rhs
	OpDefine      Op = 0x8C // X... := Y...
	OpInc         Op = 0x8D // X++
	OpDec         Op = 0x8E // X--

	/* Decl operators */
	OpValueDecl Op = 0x90 // var/const ...
	OpTypeDecl  Op = 0x91 // type ...

	/* Loop (sticky) operators (>= 0xD0) */
	OpSticky            Op = 0xD0 // not a real op.
	OpBody              Op = 0xD1 // if/block/switch/select.
	OpForLoop           Op = 0xD2
	OpRangeIter         Op = 0xD3
	OpRangeIterString   Op = 0xD4
	OpRangeIterMap      Op = 0xD5
	OpRangeIterArrayPtr Op = 0xD6
	OpReturnCallDefers  Op = 0xD7 // XXX rename to OpCallDefers
	OpVoid              Op = 0xFF // For profiling simple operation
)

const GasFactorCPU int64 = 1

//----------------------------------------
// "CPU" steps.

func (m *Machine) incrCPU(cycles int64) {
	if m.GasMeter != nil {
		gasCPU := overflow.Mulp(cycles, GasFactorCPU)
		m.GasMeter.ConsumeGas(gasCPU, "CPUCycles") // May panic if out of gas.
	}
	m.Cycles += cycles
}

const (
	// CPU cycles
	/* Control operators */
	OpCPUInvalid             = 1
	OpCPUHalt                = 1
	OpCPUNoop                = 1
	OpCPUExec                = 25
	OpCPUPrecall             = 207
	OpCPUEnterCrossing       = 100 // XXX
	OpCPUCall                = 256
	OpCPUCallNativeBody      = 424 // Todo benchmark this properly
	OpCPUDefer               = 64
	OpCPUCallDeferNativeBody = 33
	OpCPUGo                  = 1 // Not yet implemented
	OpCPUSelect              = 1 // Not yet implemented
	OpCPUSwitchClause        = 38
	OpCPUSwitchClauseCase    = 143
	OpCPUTypeSwitch          = 171
	OpCPUIfCond              = 38
	OpCPUPopValue            = 1
	OpCPUPopResults          = 1
	OpCPUPopBlock            = 3
	OpCPUPopFrameAndReset    = 15
	OpCPUPanic1              = 121
	OpCPUPanic2              = 21
	OpCPUReturn              = 38
	OpCPUReturnAfterCopy     = 38 // XXX
	OpCPUReturnFromBlock     = 36
	OpCPUReturnToBlock       = 23

	/* Unary & binary operators */
	OpCPUUpos       = 7
	OpCPUUneg       = 25
	OpCPUUnot       = 6
	OpCPUUxor       = 14
	OpCPUUrecv      = 1 // Not yet implemented
	OpCPULor        = 26
	OpCPULand       = 24
	OpCPUEql        = 160
	OpCPUEqlElement = 150 // per-element cost for array/struct equality comparisons
	OpCPUNeq        = 95
	OpCPULss        = 13
	OpCPULeq        = 19
	OpCPUGtr        = 20
	OpCPUGeq        = 26
	OpCPUAdd        = 18
	OpCPUSub        = 6
	OpCPUBor        = 23
	OpCPUXor        = 13
	OpCPUMul        = 19
	OpCPUQuo        = 16
	OpCPURem        = 18
	OpCPUShl        = 22
	OpCPUShr        = 20
	OpCPUBand       = 9
	OpCPUBandn      = 15

	/* Other expression operators */
	OpCPUEval        = 29
	OpCPUBinary1     = 19
	OpCPUIndex1      = 77
	OpCPUIndex2      = 195
	OpCPUSelector    = 32
	OpCPUSlice       = 103
	OpCPUStar        = 40
	OpCPURef         = 125
	OpCPUTypeAssert1 = 30
	OpCPUTypeAssert2 = 25
	// TODO: OpCPUStaticTypeOf is an arbitrary number.
	// A good way to benchmark this is yet to be determined.
	OpCPUStaticTypeOf = 100
	OpCPUCompositeLit = 50
	OpCPUArrayLit     = 137
	OpCPUSliceLit     = 183
	OpCPUSliceLit2    = 467
	OpCPUMapLit       = 475
	OpCPUStructLit    = 179
	OpCPUFuncLit      = 61
	OpCPUConvert      = 16

	/* Type operators */
	OpCPUFieldType     = 59
	OpCPUArrayType     = 57
	OpCPUSliceType     = 55
	OpCPUPointerType   = 1 // Not yet implemented
	OpCPUInterfaceType = 75
	OpCPUChanType      = 57
	OpCPUFuncType      = 81
	OpCPUMapType       = 59
	OpCPUStructType    = 174

	/* Statement operators */
	OpCPUAssign      = 79
	OpCPUAddAssign   = 85
	OpCPUSubAssign   = 57
	OpCPUMulAssign   = 55
	OpCPUQuoAssign   = 50
	OpCPURemAssign   = 46
	OpCPUBandAssign  = 54
	OpCPUBandnAssign = 44
	OpCPUBorAssign   = 55
	OpCPUXorAssign   = 48
	OpCPUShlAssign   = 68
	OpCPUShrAssign   = 76
	OpCPUDefine      = 111
	OpCPUInc         = 76
	OpCPUDec         = 46

	/* Decl operators */
	OpCPUValueDecl = 113
	OpCPUTypeDecl  = 100

	/* Loop (sticky) operators (>= 0xD0) */
	OpCPUSticky            = 1 // Not a real op
	OpCPUBody              = 43
	OpCPUForLoop           = 27
	OpCPURangeIter         = 105
	OpCPURangeIterString   = 55
	OpCPURangeIterMap      = 48
	OpCPURangeIterArrayPtr = 46
	OpCPUReturnCallDefers  = 78
)

//----------------------------------------
// main run loop.

func (m *Machine) Run(st Stage) {
	m.Stage = st
	if bm.OpsEnabled {
		defer func() {
			// output each machine run results to file
			bm.FinishRun()
		}()
	}
	if bm.NativeEnabled {
		defer func() {
			// output each machine run results to file
			bm.FinishNative()
		}()
	}
	defer func() {
		r := recover()

		if r != nil {
			switch r := r.(type) {
			case *Exception:
				if r.Stacktrace.IsZero() {
					r.Stacktrace = m.Stacktrace()
				}
				m.pushPanic(r.Value)
				m.Run(st)
			default:
				panic(r)
			}
		}
	}()

	for {
		if m.Debugger.enabled {
			m.Debug()
		}
		op := m.PopOp()
		if bm.OpsEnabled {
			// benchmark the operation.
			bm.StartOpCode(byte(OpVoid))
			bm.StopOpCode()
			// we do not benchmark static evaluation.
			if op != OpStaticTypeOf {
				bm.StartOpCode(byte(op))
			}
		}
		// TODO: this can be optimized manually, even into tiers.
		switch op {
		/* Control operators */
		case OpHalt:
			m.incrCPU(OpCPUHalt)
			if bm.OpsEnabled {
				bm.StopOpCode()
			}
			return
		case OpNoop:
			m.incrCPU(OpCPUNoop)
			continue
		case OpExec:
			m.incrCPU(OpCPUExec)
			m.doOpExec(op)
		case OpPrecall:
			m.incrCPU(OpCPUPrecall)
			m.doOpPrecall()
		case OpEnterCrossing:
			m.incrCPU(OpCPUEnterCrossing)
			m.doOpEnterCrossing()
		case OpCall:
			m.incrCPU(OpCPUCall)
			m.doOpCall()
		case OpCallNativeBody:
			m.incrCPU(OpCPUCallNativeBody)
			m.doOpCallNativeBody()
		case OpReturn:
			m.incrCPU(OpCPUReturn)
			m.doOpReturn()
		case OpReturnAfterCopy:
			m.incrCPU(OpCPUReturnAfterCopy)
			m.doOpReturnAfterCopy()
		case OpReturnFromBlock:
			m.incrCPU(OpCPUReturnFromBlock)
			m.doOpReturnFromBlock()
		case OpReturnToBlock:
			m.incrCPU(OpCPUReturnToBlock)
			m.doOpReturnToBlock()
		case OpDefer:
			m.incrCPU(OpCPUDefer)
			m.doOpDefer()
		case OpPanic1:
			panic("deprecated")
		case OpPanic2:
			m.incrCPU(OpCPUPanic2)
			m.doOpPanic2()
		case OpCallDeferNativeBody:
			m.incrCPU(OpCPUCallDeferNativeBody)
			m.doOpCallDeferNativeBody()
		case OpGo:
			m.incrCPU(OpCPUGo)
			panic("not yet implemented")
		case OpSelect:
			m.incrCPU(OpCPUSelect)
			panic("not yet implemented")
		case OpSwitchClause:
			m.incrCPU(OpCPUSwitchClause)
			m.doOpSwitchClause()
		case OpSwitchClauseCase:
			m.incrCPU(OpCPUSwitchClauseCase)
			m.doOpSwitchClauseCase()
		case OpTypeSwitch:
			m.incrCPU(OpCPUTypeSwitch)
			m.doOpTypeSwitch()
		case OpIfCond:
			m.incrCPU(OpCPUIfCond)
			m.doOpIfCond()
		case OpPopValue:
			m.incrCPU(OpCPUPopValue)
			m.PopValue()
		case OpPopResults:
			m.incrCPU(OpCPUPopResults)
			m.PopResults()
		case OpPopBlock:
			m.incrCPU(OpCPUPopBlock)
			m.PopBlock()
		case OpPopFrameAndReset:
			m.incrCPU(OpCPUPopFrameAndReset)
			m.PopFrameAndReset()
		/* Unary operators */
		case OpUpos:
			m.incrCPU(OpCPUUpos)
			m.doOpUpos()
		case OpUneg:
			m.incrCPU(OpCPUUneg)
			m.doOpUneg()
		case OpUnot:
			m.incrCPU(OpCPUUnot)
			m.doOpUnot()
		case OpUxor:
			m.incrCPU(OpCPUUxor)
			m.doOpUxor()
		case OpUrecv:
			m.incrCPU(OpCPUUrecv)
			m.doOpUrecv()
		/* Binary operators */
		case OpLor:
			m.incrCPU(OpCPULor)
			m.doOpLor()
		case OpLand:
			m.incrCPU(OpCPULand)
			m.doOpLand()
		case OpEql:
			m.incrCPU(OpCPUEql)
			m.doOpEql()
		case OpNeq:
			m.incrCPU(OpCPUNeq)
			m.doOpNeq()
		case OpLss:
			m.incrCPU(OpCPULss)
			m.doOpLss()
		case OpLeq:
			m.incrCPU(OpCPULeq)
			m.doOpLeq()
		case OpGtr:
			m.incrCPU(OpCPUGtr)
			m.doOpGtr()
		case OpGeq:
			m.incrCPU(OpCPUGeq)
			m.doOpGeq()
		case OpAdd:
			m.incrCPU(OpCPUAdd)
			m.doOpAdd()
		case OpSub:
			m.incrCPU(OpCPUSub)
			m.doOpSub()
		case OpBor:
			m.incrCPU(OpCPUBor)
			m.doOpBor()
		case OpXor:
			m.incrCPU(OpCPUXor)
			m.doOpXor()
		case OpMul:
			m.incrCPU(OpCPUMul)
			m.doOpMul()
		case OpQuo:
			m.incrCPU(OpCPUQuo)
			m.doOpQuo()
		case OpRem:
			m.incrCPU(OpCPURem)
			m.doOpRem()
		case OpShl:
			m.incrCPU(OpCPUShl)
			m.doOpShl()
		case OpShr:
			m.incrCPU(OpCPUShr)
			m.doOpShr()
		case OpBand:
			m.incrCPU(OpCPUBand)
			m.doOpBand()
		case OpBandn:
			m.incrCPU(OpCPUBandn)
			m.doOpBandn()
		/* Expression operators */
		case OpEval:
			m.incrCPU(OpCPUEval)
			m.doOpEval()
		case OpBinary1:
			m.incrCPU(OpCPUBinary1)
			m.doOpBinary1()
		case OpIndex1:
			m.incrCPU(OpCPUIndex1)
			m.doOpIndex1()
		case OpIndex2:
			m.incrCPU(OpCPUIndex2)
			m.doOpIndex2()
		case OpSelector:
			m.incrCPU(OpCPUSelector)
			m.doOpSelector()
		case OpSlice:
			m.incrCPU(OpCPUSlice)
			m.doOpSlice()
		case OpStar:
			m.incrCPU(OpCPUStar)
			m.doOpStar()
		case OpRef:
			m.incrCPU(OpCPURef)
			m.doOpRef()
		case OpTypeAssert1:
			m.incrCPU(OpCPUTypeAssert1)
			m.doOpTypeAssert1()
		case OpTypeAssert2:
			m.incrCPU(OpCPUTypeAssert2)
			m.doOpTypeAssert2()
		case OpStaticTypeOf:
			m.incrCPU(OpCPUStaticTypeOf)
			m.doOpStaticTypeOf()
		case OpCompositeLit:
			m.incrCPU(OpCPUCompositeLit)
			m.doOpCompositeLit()
		case OpArrayLit:
			m.incrCPU(OpCPUArrayLit)
			m.doOpArrayLit()
		case OpSliceLit:
			m.incrCPU(OpCPUSliceLit)
			m.doOpSliceLit()
		case OpSliceLit2:
			m.incrCPU(OpCPUSliceLit2)
			m.doOpSliceLit2()
		case OpFuncLit:
			m.incrCPU(OpCPUFuncLit)
			m.doOpFuncLit()
		case OpMapLit:
			m.incrCPU(OpCPUMapLit)
			m.doOpMapLit()
		case OpStructLit:
			m.incrCPU(OpCPUStructLit)
			m.doOpStructLit()
		case OpConvert:
			m.incrCPU(OpCPUConvert)
			m.doOpConvert()
		/* Type operators */
		case OpFieldType:
			m.incrCPU(OpCPUFieldType)
			m.doOpFieldType()
		case OpArrayType:
			m.incrCPU(OpCPUArrayType)
			m.doOpArrayType()
		case OpSliceType:
			m.incrCPU(OpCPUSliceType)
			m.doOpSliceType()
		case OpChanType:
			m.incrCPU(OpCPUChanType)
			m.doOpChanType()
		case OpFuncType:
			m.incrCPU(OpCPUFuncType)
			m.doOpFuncType()
		case OpMapType:
			m.incrCPU(OpCPUMapType)
			m.doOpMapType()
		case OpStructType:
			m.incrCPU(OpCPUStructType)
			m.doOpStructType()
		case OpInterfaceType:
			m.incrCPU(OpCPUInterfaceType)
			m.doOpInterfaceType()
		/* Statement operators */
		case OpAssign:
			m.incrCPU(OpCPUAssign)
			m.doOpAssign()
		case OpAddAssign:
			m.incrCPU(OpCPUAddAssign)
			m.doOpAddAssign()
		case OpSubAssign:
			m.incrCPU(OpCPUSubAssign)
			m.doOpSubAssign()
		case OpMulAssign:
			m.incrCPU(OpCPUMulAssign)
			m.doOpMulAssign()
		case OpQuoAssign:
			m.incrCPU(OpCPUQuoAssign)
			m.doOpQuoAssign()
		case OpRemAssign:
			m.incrCPU(OpCPURemAssign)
			m.doOpRemAssign()
		case OpBandAssign:
			m.incrCPU(OpCPUBandAssign)
			m.doOpBandAssign()
		case OpBandnAssign:
			m.incrCPU(OpCPUBandnAssign)
			m.doOpBandnAssign()
		case OpBorAssign:
			m.incrCPU(OpCPUBorAssign)
			m.doOpBorAssign()
		case OpXorAssign:
			m.incrCPU(OpCPUXorAssign)
			m.doOpXorAssign()
		case OpShlAssign:
			m.incrCPU(OpCPUShlAssign)
			m.doOpShlAssign()
		case OpShrAssign:
			m.incrCPU(OpCPUShrAssign)
			m.doOpShrAssign()
		case OpDefine:
			m.incrCPU(OpCPUDefine)
			m.doOpDefine()
		case OpInc:
			m.incrCPU(OpCPUInc)
			m.doOpInc()
		case OpDec:
			m.incrCPU(OpCPUDec)
			m.doOpDec()
		/* Decl operators */
		case OpValueDecl:
			m.incrCPU(OpCPUValueDecl)
			m.doOpValueDecl()
		case OpTypeDecl:
			m.incrCPU(OpCPUTypeDecl)
			m.doOpTypeDecl()
		/* Loop (sticky) operators */
		case OpBody:
			m.incrCPU(OpCPUBody)
			m.doOpExec(op)
		case OpForLoop:
			m.incrCPU(OpCPUForLoop)
			m.doOpExec(op)
		case OpRangeIter:
			m.incrCPU(OpCPURangeIter)
			m.doOpExec(op)
		case OpRangeIterArrayPtr:
			m.incrCPU(OpCPURangeIterArrayPtr)
			m.doOpExec(op)
		case OpRangeIterString:
			m.incrCPU(OpCPURangeIterString)
			m.doOpExec(op)
		case OpRangeIterMap:
			m.incrCPU(OpCPURangeIterMap)
			m.doOpExec(op)
		case OpReturnCallDefers:
			m.incrCPU(OpCPUReturnCallDefers)
			m.doOpReturnCallDefers()
		default:
			panic(fmt.Sprintf("unexpected opcode %s", op.String()))
		}
		if bm.OpsEnabled {
			if op != OpStaticTypeOf {
				bm.StopOpCode()
			}
		}
	}
}

//----------------------------------------
// push pop methods.

func (m *Machine) PushOp(op Op) {
	if debug {
		m.Printf("+o %v\n", op)
	}

	m.Ops = append(m.Ops, op)
}

func (m *Machine) PopOp() Op {
	op := m.Ops[len(m.Ops)-1]
	if debug {
		m.Printf("-o %v\n", op)
	}
	if OpSticky <= op {
		// do not pop persistent op types.
	} else {
		m.Ops = m.Ops[:len(m.Ops)-1]
	}
	return op
}

func (m *Machine) ForcePopOp() {
	if debug {
		m.Printf("-o! %v\n", m.Ops[len(m.Ops)-1])
	}
	m.Ops = m.Ops[:len(m.Ops)-1]
}

// Offset starts at 1.
// DEPRECATED use PeekStmt1() instead.
func (m *Machine) PeekStmt(offset int) Stmt {
	if debug {
		if offset != 1 {
			panic("should not happen")
		}
	}
	return m.Stmts[len(m.Stmts)-offset]
}

func (m *Machine) PeekStmt1() Stmt {
	numStmts := len(m.Stmts)
	s := m.Stmts[numStmts-1]
	if bs, ok := s.(*bodyStmt); ok {
		return bs.Active
	} else {
		return m.Stmts[numStmts-1]
	}
}

func (m *Machine) PushStmt(s Stmt) {
	if debug {
		m.Printf("+s %v\n", s)
	}
	m.Stmts = append(m.Stmts, s)
}

func (m *Machine) PushStmts(ss ...Stmt) {
	if debug {
		for _, s := range ss {
			m.Printf("+s %v\n", s)
		}
	}
	m.Stmts = append(m.Stmts, ss...)
}

func (m *Machine) PopStmt() Stmt {
	numStmts := len(m.Stmts)
	s := m.Stmts[numStmts-1]
	if debug {
		m.Printf("-s %v\n", s)
	}
	if bs, ok := s.(*bodyStmt); ok {
		return bs.PopActiveStmt()
	}

	m.Stmts = m.Stmts[:numStmts-1]

	return s
}

func (m *Machine) ForcePopStmt() (s Stmt) {
	numStmts := len(m.Stmts)
	s = m.Stmts[numStmts-1]
	if debug {
		m.Printf("-s %v\n", s)
	}
	// TODO debug lines and assertions.
	m.Stmts = m.Stmts[:len(m.Stmts)-1]
	return
}

// Offset starts at 1.
func (m *Machine) PeekExpr(offset int) Expr {
	return m.Exprs[len(m.Exprs)-offset]
}

func (m *Machine) PushExpr(x Expr) {
	if debug {
		m.Printf("+x %v\n", x)
	}
	m.Exprs = append(m.Exprs, x)
}

func (m *Machine) PopExpr() Expr {
	numExprs := len(m.Exprs)
	x := m.Exprs[numExprs-1]
	if debug {
		m.Printf("-x %v\n", x)
	}
	m.Exprs = m.Exprs[:numExprs-1]
	return x
}

// Returns reference to value in Values stack.  Offset starts at 1.
func (m *Machine) PeekValue(offset int) *TypedValue {
	return &m.Values[len(m.Values)-offset]
}

// Returns a slice of the values stack.
// Use or copy the result, as well as the slice.
func (m *Machine) PeekValues(n int) []TypedValue {
	return m.Values[len(m.Values)-n : len(m.Values)]
}

// XXX delete?
func (m *Machine) PeekType(offset int) Type {
	return m.Values[len(m.Values)-offset].T
}

func (m *Machine) PushValueFromBlock(tv TypedValue) {
	if hiv, ok := tv.V.(*HeapItemValue); ok {
		tv = hiv.Value
	}
	m.PushValue(tv)
}

func (m *Machine) PushValue(tv TypedValue) {
	if debug {
		m.Printf("+v %v\n", tv)
	}
	m.Values = append(m.Values, tv)
}

// Resulting reference is volatile.
func (m *Machine) PopValue() (tv *TypedValue) {
	tv = &m.Values[len(m.Values)-1]
	if debug {
		m.Printf("-v %v\n", tv)
	}
	m.Values = m.Values[:len(m.Values)-1]
	return tv
}

// Returns a slice of n values in the stack and decrements NumValues.
// NOTE: The results are on the values stack, so they must be copied or used
// immediately. If you need to use the machine before or during usage,
// consider using PopCopyValues().
// NOTE: the values are in stack order, oldest first, the opposite order of
// multiple pop calls.  This is used for params assignment, for example.
func (m *Machine) PopValues(n int) []TypedValue {
	popped := m.Values[len(m.Values)-n : len(m.Values)]
	m.Values = m.Values[:len(m.Values)-n]
	if debug {
		for i, tv := range popped {
			m.Printf("-vs[%d/%d] %v\n", i, n, tv)
		}
	}
	return popped
}

// Like PopValues(), but copies the values onto given slice.
func (m *Machine) PopCopyValues(res []TypedValue) {
	n := len(res)
	ptvs := m.PopValues(n)
	for i := 0; i < n; i++ {
		res[i] = ptvs[i].Copy(m.Alloc)
	}
}

// Decrements NumValues by number of last results.
func (m *Machine) PopResults() {
	if debug {
		for range m.NumResults {
			m.PopValue()
		}
	} else {
		m.Values = m.Values[:len(m.Values)-m.NumResults]
	}
	m.NumResults = 0
}

// Pops values with index start or greater.
func (m *Machine) ReapValues(start int) []TypedValue {
	end := len(m.Values)
	rs := make([]TypedValue, end-start)
	copy(rs, m.Values[start:end])
	m.Values = m.Values[:start]
	return rs
}

func (m *Machine) PushBlock(b *Block) {
	if debug {
		m.Println("+B")
	}
	m.Blocks = append(m.Blocks, b)
}

func (m *Machine) PopBlock() (b *Block) {
	if debug {
		m.Println("-B")
	}
	numBlocks := len(m.Blocks)
	b = m.Blocks[numBlocks-1]
	m.Blocks = m.Blocks[:numBlocks-1]
	return b
}

// The result is a volatile reference in the machine's type stack.
// Mutate and forget.
func (m *Machine) LastBlock() *Block {
	return m.Blocks[len(m.Blocks)-1]
}

// Pushes a frame with one less statement.
func (m *Machine) PushFrameBasic(s Stmt) {
	label := s.GetLabel()
	fr := Frame{
		Label:     label,
		Source:    s,
		NumOps:    len(m.Ops),
		NumValues: len(m.Values),
		NumExprs:  len(m.Exprs),
		NumStmts:  len(m.Stmts),
		NumBlocks: len(m.Blocks),
	}
	if debug {
		m.Printf("+F %#v\n", fr)
	}
	m.Frames = append(m.Frames, fr)
}

// TODO: track breaks/panics/returns on frame and
// ensure the counts are consistent, otherwise we mask
// bugs with frame pops.
func (m *Machine) PushFrameCall(cx *CallExpr, fv *FuncValue, recv TypedValue, isDefer bool) {
	withCross := cx.IsWithCross()
	numValues := 0
	if isDefer {
		// defer frame calls do not get their args and func from the
		// values stack (they were stored in fr.Defers).
		numValues = len(m.Values)
	} else {
		numValues = len(m.Values) - cx.NumArgs - 1
	}
	fr := Frame{
		Source:        cx,
		NumOps:        len(m.Ops),
		NumValues:     numValues,
		NumExprs:      len(m.Exprs),
		NumStmts:      len(m.Stmts),
		NumBlocks:     len(m.Blocks),
		Func:          fv,
		Receiver:      recv,
		NumArgs:       cx.NumArgs,
		IsVarg:        cx.Varg,
		LastPackage:   m.Package,
		LastRealm:     m.Realm,
		WithCross:     withCross,
		DidCrossing:   false,
		Defers:        nil,
		IsDefer:       isDefer,
		LastException: m.Exception,
	}
	if debug {
		if m.Package == nil {
			panic("should not happen")
		}
	}
	if debug {
		m.Printf("+F %#v\n", fr)
	}
	// If m.Exception is the same as the last call frame's LastException,
	// there has been no new exceptions, so there is nothing for a defer
	// call to catch.
	pfr := m.PeekCallFrame(1)
	if isDefer && m.Exception == pfr.LastException {
		m.Exception = nil
	}

	// NOTE: fr cannot be mutated from hereon, as it is a value.
	// If it must be mutated after append, use m.LastFrame() instead.
	m.Frames = append(m.Frames, fr)

	// Set the package.
	// .Package always refers to the code being run,
	// and may differ from .Realm.
	pv := fv.GetPackage(m.Store)
	if pv == nil {
		panic(fmt.Sprintf("package value missing in store: %s", fv.PkgPath))
	}
	m.Package = pv

	// If with cross, always switch to pv.Realm.
	// If method, this means the object cannot be modified if
	// stored externally by this method; but other methods can.
	if withCross {
		// since gno 0.9 cross type-checking makes this impossible.
		// XXX move this into if debug { ... }
		if !fv.IsCrossing() {
			// panic; notcrossing
			mrpath := "<no realm>"
			if m.Realm != nil {
				mrpath = m.Realm.Path
			}
			prpath := pv.PkgPath
			panic(fmt.Sprintf(
				"cannot cross-call a non-crossing function %s.%v from %s",
				prpath,
				fv.String(),
				mrpath,
			))
		}
		m.Realm = pv.GetRealm()
		return
	}

	// Non-crossing call of a crossing function like Public(cur, ...).
	if fv.IsCrossing() {
		if m.Realm != pv.Realm {
			// Illegal crossing to external realm.
			// (the function was variable and run-time check was necessary).
			// panic; not explicit
			mrpath := "<no realm>"
			if m.Realm != nil {
				mrpath = m.Realm.Path
			}
			prpath := "<no realm>"
			if pv.Realm != nil {
				prpath = pv.Realm.Path
			}
			panic(fmt.Sprintf(
				"cannot cur-call to external realm function %s.%v from %s",
				prpath,
				fv.String(),
				mrpath,
			))
		}
		// OK even if recv.Realm is different.
		return
	}

	// Not cross nor crossing.
	// Only "soft" switch to storage realm of receiver.
	var rlm *Realm
	if recv.IsDefined() { // method call
		obj := recv.GetFirstObject(m.Store)
		if obj == nil { // nil receiver
			// no switch
			return
		} else {
			recvOID := obj.GetObjectInfo().ID
			if recvOID.IsZero() ||
				(m.Realm != nil && recvOID.PkgID == m.Realm.ID) {
				// no switch
				return
			} else {
				// Implicit switch to storage realm.
				// Neither cross nor didswitch.
				recvPkgOID := ObjectIDFromPkgID(recvOID.PkgID)
				objpv := m.Store.GetObject(recvPkgOID).(*PackageValue)
				rlm = objpv.GetRealm()
				m.Realm = rlm
				// DO NOT set DidCrossing here. Make
				// DidCrossing only happen upon explicit
				// cross(fn)(...) calls and subsequent calls to
				// crossing functions from the same realm, to
				// avoid user confusion. Otherwise whether
				// DidCrossing happened or not depends on where
				// the receiver resides, which isn't explicit
				// enough to avoid confusion.
				//   fr.DidCrossing = true
				return
			}
		}
	} else { // top level function
		// no switch
		return
	}
}

func (m *Machine) PopFrame() Frame {
	numFrames := len(m.Frames)
	f := m.Frames[numFrames-1]
	if debug {
		m.Printf("-F %#v\n", f)
	}
	m.Frames = m.Frames[:numFrames-1]

	return f
}

// jump to target frame, and
// set machine accordingly.
func (m *Machine) GotoJump(depthFrames, depthBlocks int) {
	if depthFrames >= len(m.Frames) {
		panic("should not happen, depthFrames exeeds total frames")
	}
	// pop frames if with depth not zero
	if depthFrames != 0 {
		// the last popped frame
		fr := m.Frames[len(m.Frames)-depthFrames]
		// pop frames
		m.Frames = m.Frames[:len(m.Frames)-depthFrames]
		// reset
		m.Ops = m.Ops[:fr.NumOps]
		m.Values = m.Values[:fr.NumValues]
		m.Exprs = m.Exprs[:fr.NumExprs]
		m.Stmts = m.Stmts[:fr.NumStmts]
		m.Blocks = m.Blocks[:fr.NumBlocks]
		// pop stmts
		m.Stmts = m.Stmts[:len(m.Stmts)-depthFrames]
	}

	if depthBlocks >= len(m.Blocks) {
		panic("should not happen, depthBlocks exeeds total blocks")
	}
	// pop blocks
	m.Blocks = m.Blocks[:len(m.Blocks)-depthBlocks]
}

func (m *Machine) PopFrameAndReset() {
	fr := m.PopFrame()
	m.Ops = m.Ops[:fr.NumOps]
	m.Values = m.Values[:fr.NumValues]
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts]
	m.Blocks = m.Blocks[:fr.NumBlocks]
	m.PopStmt() // may be sticky
}

// TODO: optimize by passing in last frame.
func (m *Machine) PopFrameAndReturn() {
	fr := m.PopFrame()
	if debug {
		if !fr.IsCall() {
			panic("unexpected non-call (loop) frame")
		}
	}
	rtypes := fr.Func.GetType(m.Store).Results
	numRes := len(rtypes)
	m.Ops = m.Ops[:fr.NumOps]
	m.NumResults = numRes
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts]
	m.Blocks = m.Blocks[:fr.NumBlocks]
	// shift and convert results to typed-nil if undefined and not iface
	// kind.  and not func result type isn't interface kind.
	resStart := len(m.Values) - numRes
	for i := range numRes {
		res := m.Values[resStart+i]
		if res.IsUndefined() && rtypes[i].Type.Kind() != InterfaceKind {
			res.T = rtypes[i].Type
		}
		m.Values[fr.NumValues+i] = res
	}
	m.Values = m.Values[:fr.NumValues+numRes]
	m.Package = fr.LastPackage
	m.Realm = fr.LastRealm
	if m.Exception != nil {
		// Inner defer exceptions replace the outer defer
		// ones.  You can still reach the previous exceptions
		// via m.Exception.Previous*.
	} else if fr.IsDefer {
		pfr := m.PeekCallFrame(1)
		m.Exception = pfr.LastException // may or may not be nil
	}
}

func (m *Machine) PeekFrameAndContinueFor() {
	fr := m.LastFrame()
	m.Ops = m.Ops[:fr.NumOps+1]
	m.Values = m.Values[:fr.NumValues]
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts+1]
	m.Blocks = m.Blocks[:fr.NumBlocks+1]
	ls := m.PeekStmt(1).(*bodyStmt)
	ls.NextBodyIndex = ls.BodyLen
}

func (m *Machine) PeekFrameAndContinueRange() {
	fr := m.LastFrame()
	m.Ops = m.Ops[:fr.NumOps+1]
	m.Values = m.Values[:fr.NumValues+1]
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts+1]
	m.Blocks = m.Blocks[:fr.NumBlocks+1]
	ls := m.PeekStmt(1).(*bodyStmt)
	ls.NextBodyIndex = ls.BodyLen
}

func (m *Machine) NumFrames() int {
	return len(m.Frames)
}

// Returns the current frame.
func (m *Machine) LastFrame() *Frame {
	return &m.Frames[len(m.Frames)-1]
}

// MustPeekCallFrame returns the last call frame with an offset of n. It panics if the frame is not found.
func (m *Machine) MustPeekCallFrame(n int) *Frame {
	fr := m.peekCallFrame(n)
	if fr == nil {
		panic("frame not found")
	}
	return fr
}

// PeekCallFrame behaves the same as MustPeekCallFrame, but rather than panicking,
// returns nil if the frame is not found.
func (m *Machine) PeekCallFrame(n int) *Frame {
	return m.peekCallFrame(n)
}

// TODO: this function and PopUntilLastCallFrame() is used in conjunction
// spanning two disjoint operations upon return. Optimize.
// If n is 1, returns the immediately last call frame.
func (m *Machine) peekCallFrame(n int) *Frame {
	if n == 0 {
		panic("n must be positive")
	}
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if fr.IsCall() {
			if n == 1 {
				return fr
			} else {
				n-- // continue
			}
		}
	}

	return nil
}

// Returns the last defer call frame or nil.
func (m *Machine) LastDeferCallFrame() *Frame {
	return &m.Frames[len(m.Frames)-1]
}

// pops the last non-call (loop) frames
// and returns the last call frame (which is left on stack).
func (m *Machine) PopUntilLastCallFrame() *Frame {
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if fr.IsCall() {
			m.Frames = m.Frames[:i+1]
			return fr
		}
	}
	return nil
}

// pops until revive (call) frame.
func (m *Machine) PopUntilLastReviveFrame() *Frame {
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if fr.IsRevive {
			m.Frames = m.Frames[:i+1]
			return fr
		}
	}
	return nil
}

func (m *Machine) PushForPointer(lx Expr) {
	switch lx := lx.(type) {
	case *NameExpr:
		// no Lhs eval needed.
	case *IndexExpr:
		// evaluate Index
		m.PushExpr(lx.Index)
		m.PushOp(OpEval)
		// evaluate X
		m.PushExpr(lx.X)
		m.PushOp(OpEval)
	case *SelectorExpr:
		// evaluate X
		m.PushExpr(lx.X)
		m.PushOp(OpEval)
	case *StarExpr:
		// evaluate X (a reference)
		m.PushExpr(lx.X)
		m.PushOp(OpEval)
	case *CompositeLitExpr: // for *RefExpr e.g. &mystruct{}
		// evaluate lx.
		m.PushExpr(lx)
		m.PushOp(OpEval)
	default:
		panic(fmt.Sprintf(
			"illegal assignment X expression type %v",
			reflect.TypeOf(lx)))
	}
}

// Pop a pointer (for writing only).
func (m *Machine) PopAsPointer(lx Expr) PointerValue {
	pv, ro := m.PopAsPointer2(lx)
	if ro {
		m.Panic(typedString("cannot directly modify readonly tainted object (w/o method): " + lx.String()))
	}
	return pv
}

// Returns true iff:
//   - m.Realm is nil (single user mode), or
//   - tv is a ref to (external) package path, or
//   - tv is N_Readonly, or
//   - tv is not an object ("first object" ID is zero), or
//   - tv is an unreal object (no object id), or
//   - tv is an object residing in external realm
func (m *Machine) IsReadonly(tv *TypedValue) bool {
	// Returns true iff:
	//  - m.Realm is nil (single user mode)
	if m.Realm == nil {
		return false
	}
	//  - tv is a ref to package path
	if rv, ok := tv.V.(RefValue); ok && rv.PkgPath != "" {
		if rv.PkgPath == m.Package.PkgPath {
			return false // local package
		} else {
			return true // external package
		}
	}
	//   - tv is N_Readonly, or
	//   - tv is not an object ("first object" ID is zero), or
	//   - tv is an unreal object (no object id), or
	//   - tv is an object residing in external realm
	return tv.IsReadonlyBy(m.Realm.ID)
}

// Returns ro = true if the base is readonly,
// or if the base's storage realm != m.Realm and both are non-nil,
// and the lx isn't a name (base is a block),
// and the lx isn't a composite lit expr.
func (m *Machine) PopAsPointer2(lx Expr) (pv PointerValue, ro bool) {
	switch lx := lx.(type) {
	case *NameExpr:
		switch lx.Type {
		case NameExprTypeNormal:
			lb := m.LastBlock()
			pv = lb.GetPointerTo(m.Store, lx.Path)
			ro = false // always mutable
		case NameExprTypeHeapUse:
			lb := m.LastBlock()
			pv = lb.GetPointerTo(m.Store, lx.Path)
			ro = false // always mutable
		case NameExprTypeHeapClosure:
			panic("should not happen")
		default:
			panic("unexpected NameExpr in PopAsPointer")
		}
	case *IndexExpr:
		iv := m.PopValue()
		xv := m.PopValue()
		pv = xv.GetPointerAtIndex(m.Realm, m.Alloc, m.Store, iv)

		ro = m.IsReadonly(xv)
	case *SelectorExpr:
		xv := m.PopValue()
		pv = xv.GetPointerToFromTV(m.Alloc, m.Store, lx.Path)
		ro = m.IsReadonly(xv)
	case *StarExpr:
		xv := m.PopValue()
		var ok bool
		if pv, ok = xv.V.(PointerValue); !ok {
			if xv.V == nil {
				m.Panic(typedString("nil pointer dereference"))
			}
			panic("should not happen, not pointer nor nil")
		}
		ro = m.IsReadonly(xv)
	case *CompositeLitExpr: // for *RefExpr
		tv := *m.PopValue()
		hv := m.Alloc.NewHeapItem(tv)
		pv = PointerValue{
			TV:    &hv.Value,
			Base:  hv,
			Index: 0,
		}
		ro = false // always mutable
	default:
		panic("should not happen")
	}
	return
}

// for testing.
func (m *Machine) CheckEmpty() error {
	found := ""
	if len(m.Ops) > 0 {
		found = "op"
	} else if len(m.Values) > 0 {
		found = "value"
	} else if len(m.Exprs) > 0 {
		found = "expr"
	} else if len(m.Stmts) > 0 {
		found = "stmt"
	} else if len(m.Blocks) > 0 {
		for _, b := range m.Blocks {
			_, isPkg := b.GetSource(m.Store).(*PackageNode)
			if isPkg {
				// ok
			} else {
				found = "(non-package) block"
			}
		}
	} else if len(m.Frames) > 0 {
		found = "frame"
	} else if m.NumResults > 0 {
		found = ".NumResults != 0"
	}
	if found != "" {
		return fmt.Errorf("found leftover %s", found)
	} else {
		return nil
	}
}

func (m *Machine) PanicString(ex string) {
	m.Panic(typedString(ex))
}

// This function does go-panic.
// To stop execution immediately stdlib native code MUST use this rather than
// pushPanic().
// Some code in realm.go and values.go will panic(&Exception{...}) directly.
// Keep this code in sync with those calls.
// Note that m.Run() will fill in the stacktrace if it isn't present.
func (m *Machine) Panic(etv TypedValue) {
	// Construct a new exception.
	ex := &Exception{
		Value:      etv,
		Stacktrace: m.Stacktrace(),
	}
	// Panic immediately.
	panic(ex)
}

// This function does not go-panic:
// caller must return manually.
// It should ONLY be called from doOp* Op handlers,
// and should return immediately from the origin Op.
func (m *Machine) pushPanic(etv TypedValue) {
	// Construct a new exception.
	ex := &Exception{
		Value:      etv,
		Stacktrace: m.Stacktrace(),
	}
	// Pop after capturing stacktrace.
	fr := m.PopUntilLastCallFrame()
	// Link ex.Previous.
	if m.Exception == nil {
		// Recall the last m.Exception before frame.
		m.Exception = ex.WithPrevious(fr.LastException)
	} else {
		// Replace existing m.Exception with new.
		m.Exception = ex.WithPrevious(m.Exception)
	}

	m.PushOp(OpPanic2)
	m.PushOp(OpReturnCallDefers)
}

// Recover is the underlying implementation of the recover() function in the
// GnoVM. It returns nil if there was no exception to be recovered, otherwise
// it returns the [Exception], which also contains the value passed into panic().
func (m *Machine) Recover() *Exception {
	// The return value of recover is nil when **the goroutine is not
	// panicking** or recover was not called directly by a deferred
	// function.
	if m.Exception == nil {
		return nil
	}
	// The return value of recover is nil when the goroutine is not
	// panicking or **recover was not called directly by a deferred
	// function**.
	fr := m.PeekCallFrame(1) // this Recover() call.
	if fr.IsDefer {          // not **called directly**
		return nil
	}
	fr = m.PeekCallFrame(2) // what contained recover().
	if !fr.IsDefer {        // not **by a deferred function**
		return nil
	}
	// Suppose a function G defers a function D that calls recover and a
	// panic occurs in a function on the same goroutine in which G is
	// executing. When the running of deferred functions reaches D, the
	// return value of D's call to recover will be the value passed to the
	// call of panic.
	ex := m.Exception
	// If D returns normally, without starting a new panic, the panicking
	// sequence stops. In that case, the state of functions called between
	// G and the call to panic is discarded, and normal execution resumes.
	//
	// NOTE: recover() > m.Recover() will clear m.Exception but m.Exception
	// may become re-set during PopFrameAndReturn() (returning from a defer
	// call) to an older value when popping a frame with .LastException set
	// from doOpReturnCallDefers() > m.PushFrameCall(isDefer=true).
	m.Exception = nil
	return ex
}

//----------------------------------------
// inspection methods

func (m *Machine) Println(args ...any) {
	if debug {
		if enabled {
			_, file, line, _ := runtime.Caller(2) // get caller info
			caller := fmt.Sprintf("%-.12s:%-4d", path.Base(file), line)
			prefix := fmt.Sprintf("DEBUG: %17s: ", caller)
			s := prefix + strings.Repeat("|", len(m.Ops))
			fmt.Println(append([]any{s}, args...)...)
		}
	}
}

func (m *Machine) Printf(format string, args ...any) {
	if debug {
		if enabled {
			_, file, line, _ := runtime.Caller(2) // get caller info
			caller := fmt.Sprintf("%-.12s:%-4d", path.Base(file), line)
			prefix := fmt.Sprintf("DEBUG: %17s: ", caller)
			s := prefix + strings.Repeat("|", len(m.Ops))
			fmt.Printf(s+" "+format, args...)
		}
	}
}

func (m *Machine) String() string {
	if m == nil {
		return "Machine:nil"
	}
	// Calculate some reasonable total length to avoid reallocation
	// Assuming an average length of 32 characters per string
	var (
		vsLength         = len(m.Values) * 32
		ssLength         = len(m.Stmts) * 32
		xsLength         = len(m.Exprs) * 32
		bsLength         = 1024
		obsLength        = len(m.Blocks) * 32
		fsLength         = len(m.Frames) * 32
		exceptionsLength = m.Exception.NumExceptions() * 32
		totalLength      = vsLength + ssLength + xsLength + bsLength + obsLength + fsLength + exceptionsLength
	)
	var sb strings.Builder
	builder := &sb // Pointer for use in fmt.Fprintf.
	builder.Grow(totalLength)
	fmt.Fprintf(builder, "Machine:\n    Stage: %v\n    Op: %v\n    Values: (len: %d)\n", m.Stage, m.Ops[:len(m.Ops)], len(m.Values))
	for i := len(m.Values) - 1; i >= 0; i-- {
		fmt.Fprintf(builder, "          #%d %v\n", i, m.Values[i])
	}
	builder.WriteString("    Exprs:\n")
	for i := len(m.Exprs) - 1; i >= 0; i-- {
		fmt.Fprintf(builder, "          #%d %v\n", i, m.Exprs[i])
	}
	builder.WriteString("    Stmts:\n")
	for i := len(m.Stmts) - 1; i >= 0; i-- {
		fmt.Fprintf(builder, "          #%d %v\n", i, m.Stmts[i])
	}
	builder.WriteString("    Blocks:\n")
	for i := len(m.Blocks) - 1; i > 0; i-- {
		b := m.Blocks[i]
		if b == nil {
			continue
		}
		gen := builder.Len()/3 + 1
		gens := "@" // strings.Repeat("@", gen)
		if pv, ok := b.Source.(*PackageNode); ok {
			// package blocks have too much, so just
			// print the pkgpath.
			fmt.Fprintf(builder, "          %s(%d) %s\n", gens, gen, pv.PkgPath)
		} else {
			bsi := b.StringIndented("            ")
			fmt.Fprintf(builder, "          %s(%d) %s\n", gens, gen, bsi)
		}
		// Update b
		switch bp := b.Parent.(type) {
		case RefValue:
			fmt.Fprintf(builder, "            (block ref %v)\n", bp.ObjectID)
		}
	}
	builder.WriteString("    Blocks (other):\n")
	for i := len(m.Blocks) - 2; i >= 0; i-- {
		b := m.Blocks[i]
		if b == nil || b.Source == nil {
			continue
		}
		if _, ok := b.Source.(*PackageNode); ok {
			break // done, skip *PackageNode.
		} else {
			fmt.Fprintf(builder, "          #%d %s\n", i,
				b.StringIndented("            "))
		}
	}
	builder.WriteString("    Frames:\n")
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fmt.Fprintf(builder, "          #%d %s\n", i, m.Frames[i])
	}
	if m.Realm != nil {
		fmt.Fprintf(builder, "    Realm:\n      %s\n", m.Realm.Path)
	}
	if m.Exception != nil {
		builder.WriteString("    Exception:\n")
		fmt.Fprintf(builder, "      %s\n", m.Exception.Sprint(m))
	}
	return builder.String()
}

func (m *Machine) ExceptionStacktrace() string {
	if m.Exception == nil {
		return ""
	}
	var builder strings.Builder
	last := m.Exception
	first := m.Exception
	var numPrevious int
	for ; first.Previous != nil; first = first.Previous {
		numPrevious++
	}
	builder.WriteString(first.StringWithStacktrace(m))
	if numPrevious >= 2 {
		fmt.Fprintf(&builder, "... %d panic(s) elided ...\n", numPrevious-1)
	}
	if numPrevious >= 1 {
		builder.WriteString(last.StringWithStacktrace(m))
	}
	return builder.String()
}
