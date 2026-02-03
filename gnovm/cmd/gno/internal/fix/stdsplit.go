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
		// chain.Address & std.Address are converted separately to `address`
		// std.{Encode,Decode}Bech32 are removed and should be manually converted
		// RawAddress / RawAddressSize handled separately
		"std.Emit":          newSplitFunc("chain.Emit"),
		"std.DerivePkgAddr": newSplitFunc("chain.PackageAddress"),
		"std.Coin":          newSplitFunc("chain.Coin"),
		"std.Coins":         newSplitFunc("chain.Coins"),
		"std.NewCoin":       newSplitFunc("chain.NewCoin"),
		"std.NewCoins":      newSplitFunc("chain.NewCoins"),
		"std.CoinDenom":     newSplitFunc("chain.CoinDenom"),

		"std.AssertOriginCall": newSplitFunc("chain/runtime.AssertOriginCall"),
		"std.PreviousRealm":    newSplitFunc("chain/runtime.PreviousRealm"),
		"std.CurrentRealm":     newSplitFunc("chain/runtime.CurrentRealm"),
		"std.NewUserRealm":     newSplitFunc("testing.NewUserRealm"),
		"std.NewCodeRealm":     newSplitFunc("testing.NewCodeRealm"),
		"std.OriginCaller":     newSplitFunc("chain/runtime.OriginCaller"),
		"std.ChainDomain":      newSplitFunc("chain/runtime.ChainDomain"),
		"std.ChainHeight":      newSplitFunc("chain/runtime.ChainHeight"),
		"std.ChainID":          newSplitFunc("chain/runtime.ChainID"),
		"std.CallerAt":         newSplitFunc("chain/runtime.CallerAt"),
		"std.Realm":            newSplitFunc("chain/runtime.Realm"),

		"std.Banker":               newSplitFunc("chain/banker.Banker"),
		"std.NewBanker":            newSplitFunc("chain/banker.NewBanker"),
		"std.BankerType":           newSplitFunc("chain/banker.BankerType"),
		"std.OriginSend":           newSplitFunc("chain/banker.OriginSend"),
		"std.BankerTypeReadonly":   newSplitFunc("chain/banker.BankerTypeReadonly"),
		"std.BankerTypeOriginSend": newSplitFunc("chain/banker.BankerTypeOriginSend"),
		"std.BankerTypeRealmSend":  newSplitFunc("chain/banker.BankerTypeRealmSend"),
		"std.BankerTypeRealmIssue": newSplitFunc("chain/banker.BankerTypeRealmIssue"),

		"std.SetParamBool":       newSplitFunc("chain/params.SetBool"),
		"std.SetParamBytes":      newSplitFunc("chain/params.SetBytes"),
		"std.SetParamInt64":      newSplitFunc("chain/params.SetInt64"),
		"std.SetParamString":     newSplitFunc("chain/params.SetString"),
		"std.SetParamStrings":    newSplitFunc("chain/params.SetStrings"),
		"std.UpdateParamStrings": newSplitFunc("chain/params.UpdateStrings"),
		"std.SetParamUint64":     newSplitFunc("chain/params.SetUint64"),

		// Previous stdsplit iterations.
		"chain.DerivePkgAddr":        newSplitFunc("chain.PackageAddress"),
		"chain.DerivePkgAddress":     newSplitFunc("chain.PackageAddress"),
		"chain/runtime.NewUserRealm": newSplitFunc("testing.NewUserRealm"),
		"chain/runtime.NewCodeRealm": newSplitFunc("testing.NewCodeRealm"),
		"chain/runtime.CoinDenom":    newSplitFunc("chain.CoinDenom"),
		"chain/banker.Coin":          newSplitFunc("chain.Coin"),
		"chain/banker.Coins":         newSplitFunc("chain.Coins"),
		"chain/banker.NewCoin":       newSplitFunc("chain.NewCoin"),
		"chain/banker.NewCoins":      newSplitFunc("chain.NewCoins"),
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
		"testing":       "testing",
	}

	var toRename []string

	apply(
		f,
		func(c *astutil.Cursor, sc scopes) bool {
			n := c.Node()

			// This type-switch contains the business logic for the std split.
			switch n := n.(type) {
			case *ast.ImportSpec:
				unq := importPath(n)

				if n.Name != nil {
					sc.declare(n.Name, n)
				} else if id := knownImportIdentifiers[unq]; id != "" {
					sc.declare(ast.NewIdent(id), n)
				} else {
					// Don't assume other identifiers without importing them.
					return false
				}
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
				switch joined {
				case "chain.Address", "std.Address":
					// Replace both with simple `address`.
					c.Replace(&ast.Ident{
						NamePos: n.Pos(),
						Name:    "address",
					})
					fixed = true
					return false
				case "std.RawAddressSize":
					// Special case: this no longer exists, but provide a simple
					// replacement as a direct literal.
					c.Replace(&ast.BasicLit{
						ValuePos: n.Pos(),
						Kind:     token.INT,
						Value:    "20",
					})
					fixed = true
					return false
				case "std.RawAddress":
					// Special case: this no longer exists, substitute with a [20]byte.
					c.Replace(&ast.ArrayType{
						Lbrack: n.Pos(),
						Len: &ast.BasicLit{
							ValuePos: n.Pos() + 1,
							Kind:     token.INT,
							Value:    "20",
						},
						Elt: &ast.Ident{
							NamePos: n.Pos() + 2,
							Name:    "byte",
						},
					})
					fixed = true
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
					sc[:1].declare(ast.NewIdent(ident), decl)
				}
				if sc.lookup(ident) != decl {
					// Will be tackled in post
					toRename = append(toRename, target.ident)
				}
				newPkgIdent := &ast.Ident{NamePos: id.Pos(), Name: ident}
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
			if isBlockNode(c.Node()) {
				// We're popping a block; if our current scope contains one
				// of the identifiers in toRename, and none of the parent scopes
				// (except the root, for the import) does, then apply() on the
				// node again to rename the definition(s) and all usages.
				// Usages that actually refer to the import are not inspected
				// using astutil, and so are not considered by our scope analyzer.
				lastScope := sc[len(sc)-1]
				newToRename := toRename[:0]
				for _, tr := range toRename {
					if du := lastScope[tr]; du != nil {
						newName := tr + "_"
						for lastScope[newName] != nil {
							newName += "_"
						}
						du.rename(newName)
					}
					if sc[1:].lookup(tr) != nil {
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
