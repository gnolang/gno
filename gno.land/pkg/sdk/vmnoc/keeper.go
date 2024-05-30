package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"bytes"
	"context"
	"fmt"
	std2 "github.com/gnolang/gno/gnovm/stdlibs/std"
	"os"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	maxAllocTx    = 500 * 1000 * 1000
	maxAllocQuery = 1500 * 1000 * 1000 // higher limit for queries
)

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	baseKey    store.StoreKey
	iavlKey    store.StoreKey
	acck       auth.AccountKeeper
	stdlibsDir string
	gnoStore   gno.Store
	maxCycles  int64
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(
	baseKey store.StoreKey,
	iavlKey store.StoreKey,
	acck auth.AccountKeeper,
	stdlibsDir string,
	maxCycles int64,
) *VMKeeper {
	// TODO: create an Options struct to avoid too many constructor parameters
	vmk := &VMKeeper{
		baseKey:    baseKey,
		iavlKey:    iavlKey,
		acck:       acck,
		stdlibsDir: stdlibsDir,
		maxCycles:  maxCycles,
	}
	return vmk
}

func (vm *VMKeeper) Initialize(ms store.MultiStore) {
	if vm.gnoStore != nil {
		panic("should not happen")
	}
	alloc := gno.NewAllocator(maxAllocTx)
	baseSDKStore := ms.GetStore(vm.baseKey)
	iavlSDKStore := ms.GetStore(vm.iavlKey)
	vm.gnoStore = gno.NewStore(alloc, baseSDKStore, iavlSDKStore)
	vm.initBuiltinPackagesAndTypes(vm.gnoStore)
	if vm.gnoStore.NumMemPackages() > 0 {
		// for now, all mem packages must be re-run after reboot.
		// TODO remove this, and generally solve for in-mem garbage collection
		// and memory management across many objects/types/nodes/packages.
		m2 := gno.NewMachineWithOptions(
			gno.MachineOptions{
				PkgPath: "",
				Output:  os.Stdout, // XXX
				Store:   vm.gnoStore,
			})
		defer m2.Release()
		gno.DisableDebug()
		m2.PreprocessAllFilesAndSaveBlockNodes()
		gno.EnableDebug()
	}
}

func (vm *VMKeeper) getGnoStore() gno.Store {
	if vm.gnoStore == nil {
		panic("VMKeeper must first be initialized")
	}

	return vm.gnoStore
}

// AddPackage adds a package with given fileset.
func (vm *VMKeeper) AddPackage(msg MsgAddPackage) (err error) {
	pkgPath := msg.Package.Path
	memPkg := msg.Package
	gnostore := vm.getGnoStore()

	if err := msg.Package.Validate(); err != nil {
		return ErrInvalidPkgPath(err.Error())
	}
	if pv := gnostore.GetPackage(pkgPath, false); pv != nil {
		return ErrInvalidPkgPath("package already exists: " + pkgPath)
	}

	if gno.ReGnoRunPath.MatchString(pkgPath) {
		return ErrInvalidPkgPath("reserved package name: " + pkgPath)
	}

	if err := gno.TypeCheckMemPackage(memPkg, gnostore); err != nil {
		return ErrTypeCheck(err)
	}

	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Alloc:     gnostore.GetAllocator(),
			Context:   stdlibs.DefaultContext{},
			MaxCycles: vm.maxCycles,
		})
	defer m2.Release()
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM addpkg panic: %v\n%s\n",
				r, m2.String())
		}
	}()
	m2.RunMemPackage(memPkg, true)

	return nil
}

// Call calls a public Gno function (for delivertx).
func (vm *VMKeeper) Call(msg MsgCall) (res string, err error) {
	pkgPath := msg.PkgPath // to import
	fnc := msg.Func
	gnostore := vm.getGnoStore()

	pv := gnostore.GetPackage(pkgPath, false)
	pl := gno.PackageNodeLocation(pkgPath)
	pn := gnostore.GetBlockNode(pl).(*gno.PackageNode)
	ft := pn.GetStaticTypeOf(gnostore, gno.Name(fnc)).(*gno.FuncType)

	mpn := gno.NewPackageNode("main", "main", nil)
	mpn.Define("pkg", gno.TypedValue{T: &gno.PackageType{}, V: pv})
	mpv := mpn.NewPackage()

	argslist := ""
	for i := range msg.Args {
		if i > 0 {
			argslist += ","
		}
		argslist += fmt.Sprintf("arg%d", i)
	}
	expr := fmt.Sprintf(`pkg.%s(%s)`, fnc, argslist)
	xn := gno.MustParseExpr(expr)

	if err != nil {
		return "", err
	}
	cx := xn.(*gno.CallExpr)
	if cx.Varg {
		panic("variadic calls not yet supported")
	}
	if len(msg.Args) != len(ft.Params) {
		panic(fmt.Sprintf("wrong number of arguments in call to %s: want %d got %d", fnc, len(ft.Params), len(msg.Args)))
	}
	for i, arg := range msg.Args {
		argType := ft.Params[i].Type
		atv := convertArgToGno(arg, argType)
		cx.Args[i] = &gno.ConstExpr{
			TypedValue: atv,
		}
	}

	msgCtx := std2.NewDefaultContext("", 0, nil, nil)
	msgCtx.SetMsg(msg)

	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Context:   msgCtx,
			Alloc:     gnostore.GetAllocator(),
			MaxCycles: vm.maxCycles,
		})
	defer m.Release()
	m.SetActivePackage(mpv)
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM call panic: %v\n%s\n",
				r, m.String())
		}
	}()
	rtvs := m.Eval(xn)
	for i, rtv := range rtvs {
		res = res + rtv.String()
		if i < len(rtvs)-1 {
			res += "\n"
		}
	}

	res += "\n\n" // use `\n\n` as separator to separate results for single tx with multi msgs

	return res, nil
}

// Run executes arbitrary Gno code in the context of the caller's realm.
func (vm *VMKeeper) Run(msg MsgRun) (res string, err error) {
	caller := "localuser"
	gnostore := vm.getGnoStore()
	memPkg := msg.Package

	memPkg.Path = "gno.land/r/" + caller + "/run"

	if err := msg.Package.Validate(); err != nil {
		return "", ErrInvalidPkgPath(err.Error())
	}
	if err = gno.TypeCheckMemPackage(memPkg, gnostore); err != nil {
		return "", ErrTypeCheck(err)
	}

	msgCtx := std2.NewDefaultContext("", 0, nil, nil)
	msgCtx.SetMsg(msg)

	buf := new(bytes.Buffer)
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    buf,
			Store:     gnostore,
			Alloc:     gnostore.GetAllocator(),
			Context:   msgCtx,
			MaxCycles: vm.maxCycles,
		})

	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM run main addpkg panic: %v\n%s\n",
				r, m.String())
		}
	}()

	_, pv := m.RunMemPackage(memPkg, false)

	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    buf,
			Store:     gnostore,
			Alloc:     gnostore.GetAllocator(),
			Context:   msgCtx,
			MaxCycles: vm.maxCycles,
		})
	defer m2.Release()
	m2.SetActivePackage(pv)
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM run main call panic: %v\n%s\n",
				r, m2.String())
		}
	}()
	m2.RunMain()
	res = buf.String()

	return res, nil
}

// QueryFuncs returns public facing function signatures.
func (vm *VMKeeper) QueryFuncs(pkgPath string) (fsigs FunctionSignatures, err error) {
	store := vm.getGnoStore()
	// Ensure pkgPath is realm.
	if !gno.IsRealmPath(pkgPath) {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package is not realm: %s", pkgPath))
		return nil, err
	}
	// Get Package.
	pv := store.GetPackage(pkgPath, false)
	if pv == nil {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return nil, err
	}
	// Iterate over public functions.
	pblock := pv.GetBlock(store)
	for _, tv := range pblock.Values {
		if tv.T.Kind() != gno.FuncKind {
			continue // must be function
		}
		fv := tv.GetFunc()
		if fv.IsMethod {
			continue // cannot be method
		}
		fname := string(fv.Name)
		first := fname[0:1]
		if strings.ToUpper(first) != first {
			continue // must be exposed
		}
		fsig := FunctionSignature{
			FuncName: fname,
		}
		ft := fv.Type.(*gno.FuncType)
		for _, param := range ft.Params {
			pname := string(param.Name)
			if pname == "" {
				pname = "_"
			}
			ptype := gno.BaseOf(param.Type).String()
			fsig.Params = append(fsig.Params,
				NamedType{Name: pname, Type: ptype},
			)
		}
		for _, result := range ft.Results {
			rname := string(result.Name)
			if rname == "" {
				rname = "_"
			}
			rtype := gno.BaseOf(result.Type).String()
			fsig.Results = append(fsig.Results,
				NamedType{Name: rname, Type: rtype},
			)
		}
		fsigs = append(fsigs, fsig)
	}
	return fsigs, nil
}

// QueryEval evaluates a gno expression (readonly, for ABCI queries).
func (vm *VMKeeper) QueryEval(pkgPath string, expr string) (res string, err error) {
	alloc := gno.NewAllocator(maxAllocQuery)
	gnostore := vm.getGnoStore()
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	// Get Package.
	pv := gnostore.GetPackage(pkgPath, false)
	if pv == nil {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return "", err
	}
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
		return "", err
	}
	// Construct new machine.

	msgCtx := std2.NewDefaultContext("", 0, nil, nil)
	msgCtx.SetOrigPkgAddr(pkgAddr.Bech32())

	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   pkgPath,
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Context:   msgCtx,
			Alloc:     alloc,
			MaxCycles: vm.maxCycles,
		})
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM query eval panic: %v\n%s\n",
				r, m.String())
		}
	}()
	rtvs := m.Eval(xx)
	res = ""
	for i, rtv := range rtvs {
		res += rtv.String()
		if i < len(rtvs)-1 {
			res += "\n"
		}
	}
	return res, nil
}

// QueryEvalString evaluates a gno expression (readonly, for ABCI queries).
// The result is expected to be a single string (not a tuple).
// TODO: modify query protocol to allow MsgEval.
// TODO: then, rename to "EvalString".
func (vm *VMKeeper) QueryEvalString(pkgPath string, expr string) (res string, err error) {
	alloc := gno.NewAllocator(maxAllocQuery)
	gnostore := vm.getGnoStore()
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	// Get Package.
	pv := gnostore.GetPackage(pkgPath, false)
	if pv == nil {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return "", err
	}
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
		return "", err
	}

	msgCtx := std2.NewDefaultContext("", 0, nil, nil)
	msgCtx.SetOrigPkgAddr(pkgAddr.Bech32())

	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   pkgPath,
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Context:   msgCtx,
			Alloc:     alloc,
			MaxCycles: vm.maxCycles,
		})
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM query eval string panic: %v\n%s\n",
				r, m.String())
		}
	}()
	rtvs := m.Eval(xx)
	if len(rtvs) != 1 {
		return "", errors.New("expected 1 string result, got %d", len(rtvs))
	} else if rtvs[0].T.Kind() != gno.StringKind {
		return "", errors.New("expected 1 string result, got %v", rtvs[0].T.Kind())
	}
	res = rtvs[0].GetString()
	return res, nil
}

func (vm *VMKeeper) QueryFile(filepath string) (res string, err error) {
	store := vm.getGnoStore()
	dirpath, filename := std.SplitFilepath(filepath)
	if filename != "" {
		memFile := store.GetMemFile(dirpath, filename)
		if memFile == nil {
			return "", fmt.Errorf("file %q is not available", filepath) // TODO: XSS protection
		}
		return memFile.Body, nil
	} else {
		memPkg := store.GetMemPackage(dirpath)
		if memPkg == nil {
			return "", fmt.Errorf("package %q is not available", dirpath) // TODO: XSS protection
		}
		for i, memfile := range memPkg.Files {
			if i > 0 {
				res += "\n"
			}
			res += memfile.Name
		}
		return res, nil
	}
}

// logTelemetry logs the VM processing telemetry
func logTelemetry(
	cpuCycles int64,
	attributes ...attribute.KeyValue,
) {
	if !telemetry.MetricsEnabled() {
		return
	}

	// Record the operation frequency
	metrics.VMExecMsgFrequency.Add(
		context.Background(),
		1,
		metric.WithAttributes(attributes...),
	)

	// Record the CPU cycles
	metrics.VMCPUCycles.Record(
		context.Background(),
		cpuCycles,
		metric.WithAttributes(attributes...),
	)
}

func (vm *VMKeeper) QueryMemPackage(pkgPath string) *std.MemPackage {
	store := vm.getGnoStore()
	return store.GetMemPackage(pkgPath)
}

func (vm *VMKeeper) QueryMemPackages() <-chan *std.MemPackage {
	store := vm.getGnoStore()
	return store.IterMemPackage()
}
