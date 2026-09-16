// Package grc20routes records how a *signing user* reaches a registered GRC20
// token's write methods with MsgCall.
//
// # The problem
//
// grc20reg tells you a token exists and hands you a *grc20.Token to read from.
// It deliberately cannot move a user's balance: its own write wrappers act for
// the CALLING REALM, and grc20.CallerTeller — the only teller that acts for
// whoever crossed in — hangs off *PrivateLedger behind a home-realm guard, so
// no hub can ever spend on a third party's behalf. That is the confused-deputy
// defence and it is not going away.
//
// The consequence for a wallet or a DEX front-end is that every user-facing
// write must go to the token's OWN realm. That much is discoverable: the
// registry key is "<realmPath>.<SYMBOL>", so the destination realm is already in
// the key. What is not discoverable is the CALL SHAPE, and it is not uniform:
//
//	gno.land/r/gnoland/wugnot          Approve(cur, spender, amount)
//	gno.land/r/demo/defi/grc20factory  Approve(cur, symbol, spender, amount)
//
// One realm per token is the common case; a realm hosting many tokens needs a
// leading symbol argument, and that pattern is permanent. So a consumer today
// either hard-codes a table per token, or probes arity by trial and error. Both
// go stale, and the usual escape hatch is MsgRun, which mainnet gates.
//
// # Two ways in, because one is not enough
//
// The obvious design — let the token's own realm declare its shape — covers new
// realms and nothing else. A realm deployed before this one existed cannot call
// it, and cannot be redeployed. gno.land/r/demo/defi/grc20factory is live today
// and hosts most of the registry, so a table it can never write to would miss
// the majority of tokens on day one.
//
// So there are two paths, and neither needs an admin:
//
//  1. DECLARED — the token's own realm calls Register. Exact, arbitrary entry
//     point names, available to any realm deployed from here on.
//
//  2. PROVEN — anyone demonstrates, by effect, that an already-deployed realm
//     routes by symbol. See BeginKeyedProof. No permission required, because
//     the proof is the permission.
//
// Declared beats proven beats the convention. Get and JSON fall back all the
// way, so a consumer has one code path and this is not a flag day.
//
// # Why a proof, and not a claim
//
// A table anyone could write into is a table nobody can trust. What makes the
// PROVEN path safe is that the claim is checked against chain state rather than
// believed:
//
//   - The destination realm is never an argument. It is the package-path half of
//     the registry key, so a route can only ever point at the token's own realm.
//   - Proving keyedness requires moving TWO different tokens' allowances at the
//     same realm, to two prover-chosen nonces. A realm whose Approve takes no
//     leading argument reaches one token and cannot do it. Only a realm that
//     really does route by its leading argument can.
//   - Entry point names on the proven path are forced to the canonical ones. The
//     prover names nothing. The only thing established is "this realm is keyed by
//     symbol", which is exactly what the two-token effect shows.
//
// Residual, stated plainly: a realm exposing a separately-named approve per
// token (ApproveFOO, ApproveBAR) could satisfy the two-token proof without being
// keyed, and consumers would then build a call that aborts on arity. No funds
// move, and the realm can correct it permanently with Register. Nothing in this
// package can cause a token to move.
//
// # What can and cannot be forged
//
// The identifier space is grc20reg's, and every write here is anchored to it:
//
//   - A DECLARED key is fqname.Construct(cur.Previous().PkgPath(), symbol). The
//     realm path comes from the VM, never from an argument, so a realm can only
//     ever write keys beneath its own path. To land on a victim's key V.s, a
//     caller's own path would have to satisfy "A." being a prefix of "V.s" —
//     and since symbols are alphanumeric, the only such prefix is "V." itself.
//     A would have to BE V.
//   - Register additionally refuses any key grc20reg does not already hold, so a
//     key that survives is one the token's own realm created there. That check is
//     load-bearing, not hygiene.
//   - A PROVEN entry keys on a realm path, not a token, and is reachable only by
//     moving two of that realm's own tokens. Only a realm can register tokens
//     under its own path in grc20reg, so an attacker cannot manufacture the
//     second token needed to make a claim about someone else's realm. A
//     single-token realm therefore cannot be marked keyed at all.
//   - The destination is never an argument on either path: it is the package-path
//     half of the key. There is no input that aims a route at another realm.
//
// What an attacker CAN do is grief a realm that hosts two tokens behind
// separately-named approves (ApproveFOO, ApproveBAR) into being marked keyed.
// Consumers then build a call that aborts on arity. No funds move, and the realm
// ends it permanently with Register, which outranks any proof.
//
// # Consuming it
//
// One ABCI query, no parsing of Gno values:
//
//	gnokey query vm/qeval --data \
//	  'gno.land/r/nt/grc20routes/v0.JSON("gno.land/r/demo/defi/grc20factory.FOO")'
//
//	{"token_key":"gno.land/r/demo/defi/grc20factory.FOO",
//	 "pkg_path":"gno.land/r/demo/defi/grc20factory",
//	 "prefix":["FOO"],"source":"proven",
//	 "funcs":{"approve":"Approve","transfer":"Transfer","transfer_from":"TransferFrom"}}
//
// Build the call by concatenating prefix with the operation's own arguments:
//
//	MsgCall{PkgPath: pkg_path, Func: funcs.approve, Args: prefix + [spender, amount]}
//
// Func answers for one operation if the whole route is more than you need:
//
//	gnokey query vm/qeval --data \
//	  'gno.land/r/nt/grc20routes/v0.Func("gno.land/r/gnoland/wugnot.wugnot","approve")'
//
// # Declaring a route (new realms)
//
// From the token's own realm, after grc20reg.Register:
//
//	grc20routes.Register(cross(cur), symbol, grc20routes.Keyed(symbol))
//
// The operation set is open. Anything a user is expected to call belongs here,
// not just the GRC20 three — a wugnot-shaped realm would declare:
//
//	grc20routes.Register(cross(cur), "wugnot", grc20routes.Canonical().
//		With("deposit", "Deposit").
//		With("withdraw", "Withdraw"))
//
// and a realm with its own names says so outright:
//
//	grc20routes.Register(cross(cur), "uatone", grc20routes.Route{
//		Prefix: []string{ibcDenom},
//		Funcs: []grc20routes.Entry{
//			{grc20routes.OpApprove, "VoucherApprove"},
//			{grc20routes.OpTransfer, "VoucherSend"},
//		},
//	})
package grc20routes

import (
	"chain"
	"strings"

	"gno.land/p/moul/md/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/nt/grc20reg/v0"
)

var (
	// declared maps "<realmPath>.<SYMBOL>" -> *Route, written only by the
	// token's own realm.
	declared = avl.NewTree()

	// provenKeyed holds realm paths shown to route by symbol. One proof covers
	// every token that realm hosts, present and future.
	provenKeyed = avl.NewTree()

	// pending holds in-flight proofs, keyed by "<prover>/<realmPath>".
	pending = avl.NewTree()

	// probe is this realm's own address. Proofs park a throwaway allowance on
	// it; nothing here ever spends one.
	probe address
)

const (
	registerEvent = "register"
	provenEvent   = "proven"

	// Operation names. These three are the GRC20 write set and the only ones
	// this realm gives special treatment: they are what Canonical() fills in and
	// what the proven path implies. The set is NOT closed — a realm may declare
	// any operation its users need, e.g. wugnot's "deposit" and "withdraw".
	OpApprove      = "approve"
	OpTransfer     = "transfer"
	OpTransferFrom = "transfer_from"

	// CanonicalApprove and friends are the entry-point names a token realm is
	// expected to expose for the three operations above. A consumer assumes
	// these when nothing is recorded.
	CanonicalApprove      = "Approve"
	CanonicalTransfer     = "Transfer"
	CanonicalTransferFrom = "TransferFrom"

	// Source values reported by Get and JSON.
	SourceDeclared   = "declared"   // the token's own realm said so
	SourceProven     = "proven"     // shown by effect on chain
	SourceConvention = "convention" // nothing recorded; the canonical shape

	maxIdentLen  = 64
	maxOpLen     = 32
	maxSymbolLen = 64
	maxEntries   = 16
	maxPrefixLen = 4
	maxPrefixArg = 128
)

func init(cur realm) {
	probe = cur.Address()
}

// Entry binds an operation name to an exported function on the destination
// realm. The function takes `cur realm` first, then the route's Prefix, then the
// operation's own arguments.
type Entry struct {
	Op   string // "approve", "transfer", "transfer_from", "deposit", …
	Func string // exported function name, e.g. "Approve"
}

// Route is the MsgCall shape for one token's user-facing writes.
//
// The destination realm is not a field: it is always the realm the registry key
// names, so it cannot be pointed somewhere else.
//
// Funcs is an open set rather than three fixed fields, deliberately. This realm
// cannot be redeployed once it is live — which is the whole reason the proven
// path below exists — so a shape that can only ever describe approve/transfer/
// transferFrom would be a permanent ceiling. Real token realms already exceed
// it: wugnot has Deposit and Withdraw, which a wallet needs to find the same way
// it finds Approve. An operation a route does not list is one the realm does not
// expose to users, which is a meaningful thing to declare, not an omission.
type Route struct {
	// Prefix holds literal arguments that precede the operation's own ones — a
	// token key for a realm that hosts several tokens. Usually empty, or one
	// element.
	Prefix []string

	// Funcs is sorted by Op, so the rendered output is deterministic.
	Funcs []Entry
}

// Func returns the entry point for op, or "" if this route does not expose one.
func (r Route) Func(op string) string {
	for _, e := range r.Funcs {
		if e.Op == op {
			return e.Func
		}
	}
	return ""
}

// Canonical is the shape of a single-token realm: the three standard names and
// no leading arguments. It is what a consumer assumes when nothing is recorded.
func Canonical() Route {
	return Route{Funcs: []Entry{
		{OpApprove, CanonicalApprove},
		{OpTransfer, CanonicalTransfer},
		{OpTransferFrom, CanonicalTransferFrom},
	}}
}

// Keyed is the shape of a realm hosting several tokens: the standard names,
// each taking `key` as its first argument after `cur`.
func Keyed(key string) Route {
	r := Canonical()
	r.Prefix = []string{key}
	return r
}

// With returns a copy of r with op bound to fn, replacing any existing binding.
// It is the way to extend a route past the GRC20 three:
//
//	grc20routes.Canonical().With("deposit", "Deposit").With("withdraw", "Withdraw")
func (r Route) With(op, fn string) Route {
	out := Route{Prefix: r.Prefix}
	for _, e := range r.Funcs {
		if e.Op != op {
			out.Funcs = append(out.Funcs, e)
		}
	}
	out.Funcs = append(out.Funcs, Entry{op, fn})
	return out
}

// ProbeAddress is the address proofs park their throwaway allowance on. It is
// this realm, which holds no teller and calls no TransferFrom, so the allowance
// is inert; a prover may set it back to zero afterwards.
func ProbeAddress() address { return probe }

// ---------------------------------------------------------------- declared

// Register records how users reach the caller's token named `symbol`.
//
// The key is fqname.Construct(cur.Previous().PkgPath(), symbol) — the same key
// grc20reg uses — so a realm can only ever write routes for its own tokens.
// The token must already be registered in grc20reg; register there first.
//
// Re-registering the same key overwrites, which is how a realm corrects or
// updates its own declaration. No other realm can, and a declaration always
// wins over a proof.
//
// Returns the key.
func Register(cur realm, symbol string, r Route) string {
	rlmPath := cur.Previous().PkgPath()
	if rlmPath == "" {
		panic("grc20routes: caller is not a realm")
	}
	validateSymbol(symbol)
	key := fqname.Construct(rlmPath, symbol)

	// Bind this table to grc20reg's: no routes for tokens that do not exist,
	// so a key here always resolves to a token there.
	if grc20reg.Get(key) == nil {
		panic("grc20routes: not registered in grc20reg: " + key)
	}

	validate(r)
	sortEntries(r.Funcs)
	declared.Set(key, &r)

	ops := make([]string, len(r.Funcs))
	for i, e := range r.Funcs {
		ops[i] = e.Op
	}
	chain.Emit(
		registerEvent,
		"token_key", key,
		"pkgpath", rlmPath,
		"ops", strings.Join(ops, ","),
		"prefix", strings.Join(r.Prefix, ","),
	)
	return key
}

// ------------------------------------------------------------------ proven

type proof struct {
	keyA, keyB     string
	nonceA, nonceB int64
}

// BeginKeyedProof opens a proof that the realm hosting both tokens routes by
// symbol, i.e. that its entry points read Approve(cur, symbol, spender, amount).
//
// Send three more calls, in this order:
//
//  1. BeginKeyedProof(keyA, keyB, nonceA, nonceB)
//  2. <realm>.Approve(symbolA, <probe address>, nonceA)
//  3. <realm>.Approve(symbolB, <probe address>, nonceB)
//  4. FinishKeyedProof(keyA, keyB)
//
// One transaction is tidy but not required: the open proof is held against the
// prover, and only the prover can change the prover's own allowances, so there
// is nothing for anyone else to race.
//
// Both tokens must be registered in grc20reg under the same realm, with
// different symbols. The nonces must differ from each other and from the
// allowances the prover already has parked on the probe, so an allowance that
// was there beforehand cannot stand in for one this proof caused.
//
// Nothing about this is permissioned: approving is something any holder may do,
// and a realm that is not keyed cannot make two different tokens' allowances
// land at once.
func BeginKeyedProof(cur realm, tokenKeyA, tokenKeyB string, nonceA, nonceB int64) {
	prover := cur.Previous().Address()

	pathA := mustToken(tokenKeyA)
	pathB := mustToken(tokenKeyB)
	if pathA != pathB {
		panic("grc20routes: both tokens must live in the same realm")
	}
	if tokenKeyA == tokenKeyB {
		panic("grc20routes: keyed proof needs two different tokens")
	}
	if nonceA <= 0 || nonceB <= 0 {
		panic("grc20routes: nonces must be positive")
	}
	if nonceA == nonceB {
		panic("grc20routes: nonces must differ")
	}
	if allowanceOf(tokenKeyA, prover) == nonceA || allowanceOf(tokenKeyB, prover) == nonceB {
		panic("grc20routes: nonce already parked on the probe; pick another")
	}

	pending.Set(pendingKey(prover, pathA), &proof{
		keyA:   tokenKeyA,
		keyB:   tokenKeyB,
		nonceA: nonceA,
		nonceB: nonceB,
	})
}

// FinishKeyedProof commits the proof opened by BeginKeyedProof, if both probe
// allowances now read exactly the nonces that were claimed.
//
// On success every token hosted by that realm — including ones registered
// later — resolves to Keyed(<its own symbol>), unless its realm has declared
// something else.
func FinishKeyedProof(cur realm, tokenKeyA, tokenKeyB string) {
	prover := cur.Previous().Address()
	pathA := mustToken(tokenKeyA)

	pk := pendingKey(prover, pathA)
	v := pending.Get(pk)
	if v == nil {
		panic("grc20routes: no proof in progress for " + pathA)
	}
	p := v.(*proof)
	if p.keyA != tokenKeyA || p.keyB != tokenKeyB {
		panic("grc20routes: tokens do not match the proof in progress")
	}

	// The load-bearing check. Two distinct registry keys at one realm, each
	// carrying the exact allowance claimed for it. An unkeyed Approve reaches
	// one token, so it cannot produce both.
	if allowanceOf(p.keyA, prover) != p.nonceA {
		panic("grc20routes: " + p.keyA + " does not carry its nonce")
	}
	if allowanceOf(p.keyB, prover) != p.nonceB {
		panic("grc20routes: " + p.keyB + " does not carry its nonce")
	}

	pending.Remove(pk)
	provenKeyed.Set(pathA, true)

	chain.Emit(provenEvent, "pkgpath", pathA, "prover", prover.String())
}

func pendingKey(prover address, pkgPath string) string {
	return prover.String() + "/" + pkgPath
}

func allowanceOf(tokenKey string, owner address) int64 {
	return grc20reg.MustGet(tokenKey).Allowance(owner, probe)
}

// ------------------------------------------------------------------ reading

// Get returns the route for tokenKey and where it came from: SourceDeclared,
// SourceProven, or SourceConvention when nothing is recorded and the canonical
// shape is assumed.
//
// Panics if tokenKey is not a registered token.
func Get(tokenKey string) (Route, string) {
	pkgPath := mustToken(tokenKey)

	if v := declared.Get(tokenKey); v != nil {
		return *(v.(*Route)), SourceDeclared
	}
	if provenKeyed.Has(pkgPath) {
		_, symbol := fqname.Parse(tokenKey)
		return Keyed(symbol), SourceProven
	}
	return Canonical(), SourceConvention
}

// PkgPath returns the realm a user sends the MsgCall to, which is the package
// path half of the registry key.
//
// Panics if tokenKey is not a registered token.
func PkgPath(tokenKey string) string { return mustToken(tokenKey) }

// IsKeyedRealm reports whether pkgPath has been shown to route by symbol.
func IsKeyedRealm(pkgPath string) bool { return provenKeyed.Has(pkgPath) }

// JSON returns the route for tokenKey as a single JSON object, for consumers
// that reach this realm over ABCI and would rather not parse Gno values.
//
// Panics if tokenKey is not a registered token.
//
// Every field is either produced by the VM (pkg_path, source), taken from a
// registry key, or validated on the way in (see validate), so the output needs
// no escaping and cannot be made to carry a quote or a backslash.
func JSON(tokenKey string) string {
	pkgPath := mustToken(tokenKey)
	r, source := Get(tokenKey)

	prefix := ""
	for i, p := range r.Prefix {
		if i > 0 {
			prefix += ","
		}
		prefix += `"` + p + `"`
	}

	funcs := ""
	for i, e := range r.Funcs {
		if i > 0 {
			funcs += ","
		}
		funcs += `"` + e.Op + `":"` + e.Func + `"`
	}

	return ufmt.Sprintf(
		`{"token_key":"%s","pkg_path":"%s","prefix":[%s],"source":"%s","funcs":{%s}}`,
		tokenKey, pkgPath, prefix, source, funcs,
	)
}

// Func returns the entry point for one operation on tokenKey, or "" if the token
// does not expose that operation to users. The single-value form of Get, for
// consumers that want one answer rather than the whole route:
//
//	gnokey query vm/qeval --data \
//	  'gno.land/r/nt/grc20routes/v0.Func("gno.land/r/gnoland/wugnot.wugnot","approve")'
//
// Panics if tokenKey is not a registered token.
func Func(tokenKey, op string) string {
	r, _ := Get(tokenKey)
	return r.Func(op)
}

// mustToken returns the destination realm for tokenKey, panicking if the token
// is not in grc20reg.
//
// The "#subpath" a realm carries while operating under a sub identity is
// stripped, because the result is handed to consumers as a MsgCall destination
// and a sub path is not one. grc20reg keys on the raw path, so a token
// registered from a sub realm keys as "<path>#<sub>.<SYM>" there while routing
// here to "<path>" — the realm a user can actually call. This is the same
// normalization grc20's guardHome applies for the same reason.
func mustToken(tokenKey string) string {
	if grc20reg.Get(tokenKey) == nil {
		panic("grc20routes: unknown token: " + tokenKey)
	}
	pkgPath, _ := fqname.Parse(tokenKey)
	return hostOf(pkgPath)
}

// hostOf drops the "#subpath" a realm carries under a sub identity.
func hostOf(pkgPath string) string {
	host, _, _ := chain.SplitPkgSubPath(pkgPath)
	return host
}

// --------------------------------------------------------------- validation

// validate keeps the table to values that are safe to emit as JSON without
// escaping, and that can actually name a Gno entry point.
//
// Only the declared path reaches here. The proven path names nothing: its entry
// points are the canonical constants and its prefix is the registry symbol,
// both already constrained.
func validate(r Route) {
	if len(r.Funcs) == 0 {
		panic("grc20routes: route declares no entry point")
	}
	if len(r.Funcs) > maxEntries {
		panic("grc20routes: too many entry points")
	}
	for i, e := range r.Funcs {
		validateOpName(e.Op)
		if e.Func == "" {
			panic("grc20routes: empty entry point for " + e.Op)
		}
		validateFuncName(e.Op, e.Func)
		for _, prev := range r.Funcs[:i] {
			if prev.Op == e.Op {
				panic("grc20routes: duplicate operation: " + e.Op)
			}
		}
	}
	if len(r.Prefix) > maxPrefixLen {
		panic("grc20routes: too many prefix arguments")
	}
	for _, p := range r.Prefix {
		if p == "" {
			panic("grc20routes: empty prefix argument")
		}
		if len(p) > maxPrefixArg {
			panic("grc20routes: prefix argument too long")
		}
		for _, c := range p {
			if !isAlphanumeric(c) && c != '_' && c != '-' && c != '.' && c != '/' {
				panic("grc20routes: invalid prefix character: " + string(c))
			}
		}
	}
}

// validateSymbol constrains the caller-supplied symbol before it is concatenated
// into a registry key.
//
// The grc20reg lookup in Register is the real gate — a key it does not hold is
// refused whatever it looks like — so this is defence in depth rather than the
// load-bearing check. It is here anyway because the alternative is a security
// argument that lives in another package: today grc20.validSymbol allows only
// alphanumerics, "_" and "-", which is what makes a registry key unambiguous
// under fqname.Parse and safe to emit as JSON unescaped. Restating it locally
// means this realm keeps holding even if that ever loosens.
func validateSymbol(symbol string) {
	if symbol == "" {
		panic("grc20routes: empty symbol")
	}
	if len(symbol) > maxSymbolLen {
		panic("grc20routes: symbol too long")
	}
	for _, c := range symbol {
		if !isAlphanumeric(c) && c != '_' && c != '-' {
			panic("grc20routes: invalid symbol character: " + string(c))
		}
	}
}

// validateOpName accepts a lowercase snake_case operation name. Op names reach
// the JSON output as object keys, so they are constrained the same way function
// names are.
func validateOpName(op string) {
	if op == "" {
		panic("grc20routes: empty operation name")
	}
	if len(op) > maxOpLen {
		panic("grc20routes: operation name too long: " + op)
	}
	for _, c := range op {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_' {
			panic("grc20routes: invalid character in operation name: " + string(c))
		}
	}
}

// validateFuncName requires an exported identifier.
func validateFuncName(field, name string) {
	if len(name) > maxIdentLen {
		panic("grc20routes: " + field + " name too long")
	}
	for i, c := range name {
		if i == 0 {
			if c < 'A' || c > 'Z' {
				panic("grc20routes: " + field + " must be an exported function name")
			}
			continue
		}
		if !isAlphanumeric(c) && c != '_' {
			panic("grc20routes: invalid character in " + field + " name: " + string(c))
		}
	}
}

func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// sortEntries orders a route's entries by Op so the JSON and Render output for a
// given route is always the same string. Insertion sort: maxEntries is 16.
func sortEntries(es []Entry) {
	for i := 1; i < len(es); i++ {
		e := es[i]
		j := i - 1
		for j >= 0 && es[j].Op > e.Op {
			es[j+1] = es[j]
			j--
		}
		es[j+1] = e
	}
}

// ------------------------------------------------------------------- render

func Render(path string) string {
	if path != "" {
		return renderToken(path)
	}

	s := "# grc20routes\n\n"
	s += "How a signing user reaches a registered GRC20 token's write methods with `MsgCall`.\n\n"
	s += "A token that appears nowhere below uses "
	s += ufmt.Sprintf("`%s` / `%s` / `%s` with no leading arguments.\n\n",
		CanonicalApprove, CanonicalTransfer, CanonicalTransferFrom)

	s += "## Realms proven to route by symbol\n\n"
	n := 0
	provenKeyed.Iterate("", "", func(pkgPath string, _ any) bool {
		n++
		s += "- " + md.InlineCode(pkgPath) + " — `Approve(cur, SYMBOL, spender, amount)`\n"
		return false
	})
	if n == 0 {
		s += "_None yet._\n"
	}

	s += "\n## Routes declared by their own realm\n\n"
	n = 0
	declared.Iterate("", "", func(key string, v any) bool {
		n++
		r := v.(*Route)
		rlmPath, symbol := fqname.Parse(key)
		s += "- " + fqname.RenderLink(rlmPath, symbol) + " — "
		for i, e := range r.Funcs {
			if i > 0 {
				s += ", "
			}
			s += md.InlineCode(e.Op + "=" + e.Func)
		}
		if len(r.Prefix) > 0 {
			s += " (prefix " + md.InlineCode(strings.Join(r.Prefix, ", ")) + ")"
		}
		s += "\n"
		return false
	})
	if n == 0 {
		s += "_None yet._\n"
	}
	return s
}

func renderToken(tokenKey string) string {
	r, source := Get(tokenKey)
	pkgPath, symbol := fqname.Parse(tokenKey)

	s := ufmt.Sprintf("# %s\n\n", md.EscapeText(symbol))
	s += "- token key: " + md.InlineCode(tokenKey) + "\n"
	s += "- realm: " + fqname.RenderLink(pkgPath, symbol) + "\n"
	s += "- source: " + source + "\n"
	if source == SourceConvention {
		s += "- _nothing recorded for this token; the canonical shape is assumed_\n"
	}
	s += "\n## Entry points\n\n"
	if len(r.Funcs) == 0 {
		s += "_None; this token exposes no user-callable writes._\n"
	}
	for _, e := range r.Funcs {
		s += renderEntry(e, r.Prefix)
	}
	return s
}

// opTail is the argument list an operation takes after the prefix. Unknown
// operations render without one rather than guessing.
func opTail(op string) string {
	switch op {
	case OpApprove:
		return ", spender, amount"
	case OpTransfer:
		return ", to, amount"
	case OpTransferFrom:
		return ", from, to, amount"
	}
	return ", …"
}

func renderEntry(e Entry, prefix []string) string {
	args := "cur"
	for _, p := range prefix {
		args += ", " + md.EscapeText(p)
	}
	return ufmt.Sprintf("- %s: `%s(%s%s)`\n", e.Op, e.Func, args, opTail(e.Op))
}
