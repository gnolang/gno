package vm

import (
	"fmt"
	"os"
	"reflect"

	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/log"
	"github.com/gnolang/gno/stdlibs"
)

type Wrapper struct {
	msgCtx stdlibs.ExecContext
	msg    MsgCall
	store  gno.Store
	m      *gno.Machine
	logger log.Logger
}

func NewWrapper(msg MsgCall, store gno.Store) *Wrapper {
	w := &Wrapper{
		msg:   msg,
		store: store,
	}
	return w
}

func (w *Wrapper) Eval() (res VMResult, err error) {
	pkgPath := w.msg.PkgPath // to import
	fnc := w.msg.Func
	// Get the package and function type.
	pv := w.store.GetPackage(pkgPath, false)
	pl := gno.PackageNodeLocation(pkgPath)
	pn := w.store.GetBlockNode(pl).(*gno.PackageNode)
	ft := pn.GetStaticTypeOf(w.store, gno.Name(fnc)).(*gno.FuncType)
	// Make main Package with imports.
	mpn := gno.NewPackageNode("main", "main", nil)
	mpn.Define("pkg", gno.TypedValue{T: &gno.PackageType{}, V: pv})
	mpv := mpn.NewPackage()
	// Parse expression.
	argslist := ""
	for i := range w.msg.Args {
		if i > 0 {
			argslist += ","
		}
		argslist += fmt.Sprintf("arg%d", i)
	}
	expr := fmt.Sprintf(`pkg.%s(%s)`, fnc, argslist)
	xn := gno.MustParseExpr(expr)

	// Convert Args to gno values.
	cx := xn.(*gno.CallExpr)
	if cx.Varg {
		panic("variadic calls not yet supported")
	}
	for i, arg := range w.msg.Args {
		argType := ft.Params[i].Type
		atv := gno.ConvertArgToGno(arg, argType)
		cx.Args[i] = &gno.ConstExpr{
			TypedValue: atv,
		}
	}

	// Construct machine and evaluate.
	w.m = gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     w.store,
			Context:   w.msgCtx,
			Alloc:     w.store.GetAllocator(),
			MaxCycles: 10 * 1000 * 1000, // 10M cycles // XXX
		})
	w.m.SetActivePackage(mpv)

	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM call panic: %v\n%s\n",
				r, w.m.String())
			return
		}
	}()

	rtvs := w.m.Eval(xn)

	// TODO: compatible with ealier response, or pack the response here with raw strs from VM?
	// maybe pack inside is better? NO, should make contract writing simple
	// check return type to determine how to handle it?

	var rstr string
	var rt reflect.Type
	// get reflect type
	for i, rtv := range rtvs {
		// assert first value
		rtvp, ok := rtv.V.(gno.PointerValue)
		if ok { // assert pass
			rt = gno.Gno2GoType(rtvp.TV)
			if rt.Kind() == reflect.Struct { // struct kind
				var result VMResult
				vrf := reflect.ValueOf(&result).Elem()
				gno.Gno2GoValue(rtvs[0].V.(gno.PointerValue).TV, vrf)
				rs := vrf.Interface().(VMResult)
				// length prefixed
				rs.Data = prefixData(rs.Data)
				return rs, nil
			}
		} else {
			rstr = rstr + rtv.String()
			if i < len(rtvs)-1 {
				rstr += "\n"
			}
			return gnoResultFromData([]byte(rstr)), nil
		}
	}
	return VMResult{}, err
}
