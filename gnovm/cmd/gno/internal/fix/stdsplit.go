package fix

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

var stdsplitFix = register(Fix{
	Name: "stdsplit",
	Date: "2025-04-15",
	Desc: "rewrites imports and calls to std packages into the new functions",
	F:    stdsplit,
})

type splitFunc struct {
	importPath string
	ident      string
	fname      string
}

func newSplitFunc(joined string) splitFunc {
	path, name, ok := strings.Cut(joined, ".")
	if !ok {
		panic(fmt.Errorf("invalid joined string: %q", joined))
	}
	il := strings.LastIndexByte(path, '/')
	return splitFunc{
		importPath: path,
		ident:      path[il+1:],
		fname:      name,
	}
}

func stdsplit(f *ast.File) (fixed bool) {
	splitFuncs := map[string]splitFunc{
		"std.Address":       newSplitFunc("chain.Address"),
		"std.Emit":          newSplitFunc("chain.Emit"),
		"std.EncodeBech32":  newSplitFunc("chain.EncodeBech32"),
		"std.DecodeBech32":  newSplitFunc("chain.DecodeBech32"),
		"std.DerivePkgAddr": newSplitFunc("chain.DerivePkgAddr"),

		"std.AssertOriginCall": newSplitFunc("chain/runtime.AssertOriginCall"),
		"std.PreviousRealm":    newSplitFunc("chain/runtime.PreviousRealm"),
		"std.CurrentRealm":     newSplitFunc("chain/runtime.CurrentRealm"),
		"std.NewUserRealm":     newSplitFunc("chain/runtime.NewUserRealm"),
		"std.NewCodeRealm":     newSplitFunc("chain/runtime.NewCodeRealm"),
		"std.OriginCaller":     newSplitFunc("chain/runtime.OriginCaller"),
		"std.ChainDomain":      newSplitFunc("chain/runtime.ChainDomain"),
		"std.ChainHeight":      newSplitFunc("chain/runtime.ChainHeight"),
		"std.ChainID":          newSplitFunc("chain/runtime.ChainID"),
		"std.CoinDenom":        newSplitFunc("chain/runtime.CoinDenom"),
		"std.CallerAt":         newSplitFunc("chain/runtime.CallerAt"),
		"std.Realm":            newSplitFunc("chain/runtime.Realm"),

		"std.Banker":               newSplitFunc("chain/banker.Banker"),
		"std.NewBanker":            newSplitFunc("chain/banker.NewBanker"),
		"std.BankerType":           newSplitFunc("chain/banker.BankerType"),
		"std.OriginSend":           newSplitFunc("chain/banker.OriginSend"),
		"std.Coin":                 newSplitFunc("chain/banker.Coin"),
		"std.Coins":                newSplitFunc("chain/banker.Coins"),
		"std.NewCoin":              newSplitFunc("chain/banker.NewCoin"),
		"std.NewCoins":             newSplitFunc("chain/banker.NewCoins"),
		"std.BankerTypeReadonly":   newSplitFunc("chain/banker.BankerTypeReadonly"),
		"std.BankerTypeOriginSend": newSplitFunc("chain/banker.BankerTypeOriginSend"),
		"std.BankerTypeRealmSend":  newSplitFunc("chain/banker.BankerTypeRealmSend"),
		"std.BankerTypeRealmIssue": newSplitFunc("chain/banker.BankerTypeRealmIssue"),

		"std.SetParamBool":    newSplitFunc("chain/params.SetBool"),
		"std.SetParamBytes":   newSplitFunc("chain/params.SetBytes"),
		"std.SetParamInt64":   newSplitFunc("chain/params.SetInt64"),
		"std.SetParamString":  newSplitFunc("chain/params.SetString"),
		"std.SetParamStrings": newSplitFunc("chain/params.SetStrings"),
		"std.SetParamUint64":  newSplitFunc("chain/params.SetUint64"),

		// TODO: compat package for AddressSet
	}

	// From a previous batch of std changes: https://github.com/gnolang/gno/pull/3374
	splitFuncs["std.GetOrigSend"] = splitFuncs["std.OriginSend"]
	splitFuncs["std.GetOrigCaller"] = splitFuncs["std.OriginCaller"]
	splitFuncs["std.PrevRealm"] = splitFuncs["std.PreviousRealm"]
	splitFuncs["std.GetCallerAt"] = splitFuncs["std.CallerAt"]
	splitFuncs["std.GetChainID"] = splitFuncs["std.ChainID"]
	splitFuncs["std.GetBanker"] = splitFuncs["std.NewBanker"]
	splitFuncs["std.GetChainDomain"] = splitFuncs["std.ChainDomain"]
	splitFuncs["std.GetHeight"] = splitFuncs["std.ChainHeight"]

	knownImportIdentifiers := map[string]string{
		"std":           "std",
		"chain":         "chain",
		"chain/runtime": "runtime",
		"chain/params":  "params",
		"chain/banker":  "banker",
	}
	sc := scopes{scope{}}
	astutil.Apply(
		f,
		func(c *astutil.Cursor) bool {
			n := c.Node()

			// This type-switch contains the business logic for the std split.
			switch n := n.(type) {
			case *ast.SelectorExpr:
				id, ok := n.X.(*ast.Ident)
				if !ok {
					break
				}
				def, ok := sc.lookup(id.Name).(*ast.ImportSpec)
				if !ok {
					break
				}
				ip := importPath(def)
				joined := ip + "." + n.Sel.Name
				target, ok := splitFuncs[joined]
				if !ok {
					if ip == "std" {
						// TODO: handle more gracefully.
						panic(fmt.Errorf(
							"file contains function std.%s that cannot be converted",
							n.Sel.Name,
						))
					}
					break
				}
				if addImport(f, target.importPath) {
					// TODO: handle colliding identifiers, both at the top level
					// and for shadowing.
					if sc[0][target.ident] != nil {
						panic(fmt.Errorf(
							"cannot add import for %q when top-level identifier %q is already defined",
							target.importPath,
							target.ident,
						))
					}
					decl := f.Imports[len(f.Imports)-1]
					sc[0][target.ident] = decl
					if sc.lookup(target.ident) != decl {
						panic(fmt.Errorf(
							"identifier %q is shadowed and cannot be added as import for %q",
							target.ident,
							target.importPath,
						))
					}
				}
				c.Replace(&ast.SelectorExpr{
					X:   &ast.Ident{Name: target.ident},
					Sel: &ast.Ident{Name: target.fname},
				})
				fixed = true
				return false
			}

			// This contains the logic for handling scopes.
			switch n := n.(type) {
			case *ast.ImportSpec:
				unq := importPath(n)
				id := knownImportIdentifiers[unq]
				if n.Name != nil {
					id = n.Name.Name
				} else if id == "" {
					// Don't assume other identifiers without importing them.
					return false
				}
				sc.declare(id, n)
			case *ast.TypeSpec:
				sc.declare(n.Name.Name, n)
			case *ast.ValueSpec:
				for _, name := range n.Names {
					sc.declare(name.Name, n)
				}
			case *ast.AssignStmt:
				if n.Tok == token.DEFINE {
					for _, name := range n.Lhs {
						// only declare if it doesn't exist in the last scope,
						// := allows the LHS to contain already defined values
						// which are then simply assigned instead of declared.
						name := name.(*ast.Ident).Name
						if _, ok := sc[len(sc)-1][name]; !ok {
							sc.declare(name, n)
						}
					}
				}
			case *ast.FuncDecl:
				name := n.Name.Name
				if n.Recv != nil && len(n.Recv.List) > 0 {
					tp := recvType(n.Recv.List[0].Type)
					if tp != nil {
						name = tp.Name + "." + name
					}
				}
				sc.declare(name, n)
				sc.push()
				sc.declareList(n.Recv)
				sc.declareList(n.Type.Params)
				sc.declareList(n.Type.Results)
			case *ast.FuncLit:
				sc.push()
				sc.declareList(n.Type.Params)
				sc.declareList(n.Type.Results)
			case *ast.RangeStmt:
				sc.push()
				if n.Tok == token.DEFINE {
					if id, ok := n.Key.(*ast.Ident); ok {
						sc.declare(id.Name, n)
					}
					if id, ok := n.Value.(*ast.Ident); ok {
						sc.declare(id.Name, n)
					}
				}
			case *ast.BlockStmt,
				*ast.IfStmt,
				*ast.SwitchStmt,
				*ast.TypeSwitchStmt,
				*ast.CaseClause,
				*ast.CommClause,
				*ast.ForStmt,
				*ast.SelectStmt:
				sc.push()
			}
			return true
		},
		func(c *astutil.Cursor) bool {
			n := c.Node()
			switch n.(type) {
			case *ast.BlockStmt,
				*ast.FuncLit,
				*ast.IfStmt,
				*ast.SwitchStmt,
				*ast.TypeSwitchStmt,
				*ast.CaseClause,
				*ast.CommClause,
				*ast.ForStmt,
				*ast.SelectStmt,
				*ast.RangeStmt:
				sc.pop()
			}
			return true
		},
	)
	if deleteImport(f, "std") {
		fixed = true
	}
	return
}

func recvType(x ast.Expr) *ast.Ident {
	if sx, ok := x.(*ast.StarExpr); ok {
		x = sx.X
	}

	if id, ok := x.(*ast.Ident); ok {
		return id
	}
	return nil
}
