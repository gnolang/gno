package fix

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

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

var splitFuncs map[string]splitFunc

func makeSplitFuncs() {
	splitFuncs = map[string]splitFunc{
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
}

func stdsplit(f *ast.File) (fixed bool) {
	if splitFuncs == nil {
		makeSplitFuncs()
	}

	knownImportIdentifiers := map[string]string{
		"std":           "std",
		"chain":         "chain",
		"chain/runtime": "runtime",
		"chain/params":  "params",
		"chain/banker":  "banker",
	}

	var toRename []string
	ignoreIdents := map[*ast.Ident]struct{}{}

	apply(
		f,
		func(c *astutil.Cursor, sc scopes) bool {
			n := c.Node()

			// This type-switch contains the business logic for the std split.
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
				if joined == "std.RawAddressSize" {
					// Special case: this no longer exists, but provide a simple
					// replacement as a direct literal.
					c.Replace(&ast.BasicLit{
						ValuePos: n.Pos(),
						Kind:     token.INT,
						Value:    "20",
					})
					return false
				}
				target, ok := splitFuncs[joined]
				if !ok {
					// There's nothing to convert.
					// NOTE: this is also the case for some outright removed
					// functions, but these will error in lint and the user can
					// fix them.
					break
				}
				ident := target.ident
				pos := slices.IndexFunc(f.Imports, func(decl *ast.ImportSpec) bool {
					return importPath(decl) == target.importPath
				})
				var decl *ast.ImportSpec
				if pos >= 0 {
					if name := f.Imports[pos].Name; name != nil {
						ident = name.Name
					} else {
						ident = knownImportIdentifiers[target.importPath]
					}
					decl = f.Imports[pos]
				} else {
					importName := ""
					for sc[0][ident] != nil {
						ident += "_"
						importName = ident
					}
					if !addImport(f, target.importPath, importName) {
						panic("import should not exist")
					}

					decl = f.Imports[len(f.Imports)-1]
					sc[0][ident] = decl
				}
				if sc.lookup(ident) != decl {
					// Will be tackled in post
					toRename = append(toRename, target.ident)
				}
				newPkgIdent := &ast.Ident{NamePos: id.Pos(), Name: ident}
				ignoreIdents[newPkgIdent] = struct{}{}
				c.Replace(&ast.SelectorExpr{
					X:   newPkgIdent,
					Sel: &ast.Ident{NamePos: n.Sel.Pos(), Name: target.fname},
				})
				fixed = true
				return false
			}

			return true
		},
		func(c *astutil.Cursor, sc scopes) bool {
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
				// We're popping a block; if our current scope contains one
				// of the identifiers in toRename, and none of the parent scopes
				// (except the root, for the import) does, then apply() on the
				// node again to rename the definition(s) and all usages.
				// Usages that actually refer to the import are detected through
				// ignoreIdents.

				newToRename := toRename[:0]
				for _, tr := range toRename {
					if sc[len(sc)-1][tr] != nil &&
						scopes(sc[1:len(sc)-1]).lookup(tr) == nil {
						astutil.Apply(
							n,
							func(c *astutil.Cursor) bool {
								switch n := c.Node().(type) {
								case *ast.Ident:
									_, ignore := ignoreIdents[n]
									if !ignore && n.Name == tr {
										// NOTE: there could still be collisions,
										// but these would be extreme edge cases.
										n.Name += "_"
									}
									// This may overcorrect some slice literals
									// which use names which are being replaced;
									// let the user fix them.
								}
								return true
							},
							nil,
						)
					} else {
						newToRename = append(newToRename, tr)
					}
				}
				toRename = newToRename
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
