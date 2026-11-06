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
//
//   - Proving keyedness requires moving TWO different tokens' allowances at the
//     same realm, to two prover-chosen nonces. A realm whose Approve takes no
//     leading argument reaches one token and cannot do it. Only a realm that
//     really does route by its leading argument can.
//
//   - Entry point names on the proven path are forced to the canonical ones. The
//     prover names nothing. The only thing established is "this realm is keyed by
//     symbol", which is exactly what the two-token effect shows.
//
//   - The prover must be a signing user (MsgCall, IsUserCall). A realm prover is
//     refused, because grc20reg.Approve sets the calling REALM's own allowance on
//     any registered token, which would let it produce both nonces without ever
//     calling the token's realm.
//
// What the proof observes is two allowances, not the arguments that produced
// them, so the only claim it supports is "keyed by its registry SYMBOL". It can
// be satisfied by a realm that is not keyed that way, in three ways:
//
//   - separately-named approves per token (ApproveFOO, ApproveBAR);
//   - a realm keyed by some other identifier (an IBC denom, say), proven by a
//     prover who passes that identifier instead of the symbol the procedure
//     names;
//   - a signer allowed to send MsgRun: the ephemeral realm's address is the
//     signer's own, so grc20reg.Approve run from it lands where the proof looks.
//     Mainnet gates MsgRun.
//
// In every case consumers then build a call that aborts (on arity, or on an
// unknown key). No funds move. A realm that can call this one ends it for good
// with RefuseKeyedProof; one that cannot has no remedy here, which is why the
// prover is held to a signing user. Nothing in this package can cause a token
// to move.
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
//   - A token registered under a sub identity ("<path>#<sub>.<SYM>") is never
//     given a proven or conventional route. Its MsgCall destination is the host
//     realm, where the same symbol may name the host's own token, so any route
//     built for it without the realm's say-so could approve the wrong token. It
//     resolves to its declaration, or to nothing.
//
// What an attacker CAN do is grief a realm into being marked keyed by one of the
// three shapes above. Consumers then build a call that aborts. No funds move,
// and a realm that can call this one ends it permanently with RefuseKeyedProof.
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
// There is deliberately no getter for an entry point name alone: without the
// prefix it builds a call that is one argument short on a keyed realm.
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

	// refused holds realm paths that called RefuseKeyedProof. No proof can mark
	// them, ever again.
	refused = avl.NewTree()

	// pending holds in-flight proofs, keyed by "<prover>/<realmPath>".
	pending = avl.NewTree()

	// probe is this realm's own address. Proofs park a throwaway allowance on
	// it; nothing here ever spends one.
	probe address
)

const (
	registerEvent = "register"
	provenEvent   = "proven"
	refusedEvent  = "refused"

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
	SourceNone       = "none"       // a sub-identity token nothing was declared for

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
// land at once. The prover must be a signing user, though: see mustUserCall.
func BeginKeyedProof(cur realm, tokenKeyA, tokenKeyB string, nonceA, nonceB int64) {
	prover := mustUserCall(cur)

	pathA := mustToken(tokenKeyA)
	pathB := mustToken(tokenKeyB)
	if pathA != pathB {
		panic("grc20routes: both tokens must live in the same realm")
	}
	if isSubToken(tokenKeyA) || isSubToken(tokenKeyB) {
		panic("grc20routes: keyed proof takes tokens of the realm's own identity, not a sub identity")
	}
	if refused.Has(pathA) {
		panic("grc20routes: realm refuses keyed proofs: " + pathA)
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
// On success every token that realm hosts under its own identity — including
// ones registered later — resolves to Keyed(<its own symbol>), unless its realm
// has declared something else. Tokens under a sub identity are not covered.
func FinishKeyedProof(cur realm, tokenKeyA, tokenKeyB string) {
	prover := mustUserCall(cur)
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
	if refused.Has(pathA) {
		panic("grc20routes: realm refuses keyed proofs: " + pathA)
	}
	provenKeyed.Set(pathA, true)

	chain.Emit(provenEvent, "pkgpath", pathA, "prover", prover.String())
}

// RefuseKeyedProof is how a realm that is not keyed by symbol takes itself out
// of the proven path, for good: any existing proof is dropped, and no later one
// can mark it. Per-token Register cannot do this, because a proof covers the
// whole realm, including tokens registered later.
//
// It acts on the caller's own host realm and nothing else.
func RefuseKeyedProof(cur realm) {
	rlmPath := cur.Previous().PkgPath()
	if rlmPath == "" {
		panic("grc20routes: caller is not a realm")
	}
	host := hostOf(rlmPath)
	refused.Set(host, true)
	provenKeyed.Remove(host)
	chain.Emit(refusedEvent, "pkgpath", host)
}

// mustUserCall returns the prover, who must be a signing user calling in
// directly. A realm is refused: grc20reg.Approve sets the calling realm's own
// allowance on ANY registered token, so a realm prover could make both nonces
// land without ever calling the token's realm. A user's own allowance can only
// be moved by the token's realm, or by a MsgRun signed by that user (see the
// package doc).
func mustUserCall(cur realm) address {
	prev := cur.Previous()
	if !prev.IsUserCall() {
		panic("grc20routes: a keyed proof must be sent by a signing user with MsgCall")
	}
	return prev.Address()
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
// A token registered under a sub identity gets SourceNone and an empty route
// unless its realm declared one: its destination is the host realm, where its
// symbol may name a different token, so nothing short of a declaration can say
// which call reaches it.
//
// Panics if tokenKey is not a registered token.
func Get(tokenKey string) (Route, string) {
	pkgPath := mustToken(tokenKey)

	if v := declared.Get(tokenKey); v != nil {
		return *(v.(*Route)), SourceDeclared
	}
	if isSubToken(tokenKey) {
		return Route{}, SourceNone
	}
	if provenKeyed.Has(pkgPath) {
		_, symbol := splitKey(tokenKey)
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
	pkgPath, _ := splitKey(tokenKey)
	return hostOf(pkgPath)
}

// splitKey splits a registry key at its LAST dot. A symbol never contains one
// (see validateSymbol), but a sub identity may ("<path>#a.b.<SYM>"), and
// fqname.Parse splits at the first dot after the last slash, which would carry
// part of the sub path into the symbol.
func splitKey(tokenKey string) (pkgPath, symbol string) {
	i := strings.LastIndex(tokenKey, ".")
	if i < 0 {
		return tokenKey, ""
	}
	return tokenKey[:i], tokenKey[i+1:]
}

// isSubToken reports whether tokenKey was registered under a sub identity.
func isSubToken(tokenKey string) bool {
	pkgPath, _ := splitKey(tokenKey)
	return pkgPath != hostOf(pkgPath)
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
		rlmPath, symbol := splitKey(key)
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
	pkgPath, symbol := splitKey(tokenKey)

	s := ufmt.Sprintf("# %s\n\n", md.EscapeText(symbol))
	s += "- token key: " + md.InlineCode(tokenKey) + "\n"
	s += "- realm: " + fqname.RenderLink(pkgPath, symbol) + "\n"
	s += "- source: " + source + "\n"
	if source == SourceConvention {
		s += "- _nothing recorded for this token; the canonical shape is assumed_\n"
	}
	if source == SourceNone {
		s += "- _registered under a sub identity and nothing declared; no route is known_\n"
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
		// Validated to a character set with no backtick, and inside a code
		// span, so it goes out verbatim: escaping would put backslashes into
		// the argument a reader copies.
		args += ", " + p
	}
	return ufmt.Sprintf("- %s: `%s(%s%s)`\n", e.Op, e.Func, args, opTail(e.Op))
}
