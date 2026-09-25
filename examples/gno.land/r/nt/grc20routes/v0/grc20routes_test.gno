package grc20routes

import (
	"strings"
	"testing"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/r/nt/grc20reg/v0"
)

const selfPath = "gno.land/r/nt/grc20routes/v0"

var (
	nextID  seqid.ID
	ledgers = map[string]*grc20.PrivateLedger{} // symbol -> ledger
)

// mint creates a token owned by this realm, registers it in grc20reg and keeps
// its ledger, so tests can produce the exact allowance state a real token realm
// would produce. Both registries key on cur.Previous().PkgPath(), so from inside
// this package they resolve to selfPath and the keys line up.
//
// Non-crossing (`_ int, rlm realm`) so it borrows the test's own live frame.
func mint(_ int, rlm realm, symbol string) string {
	tok, led := grc20.NewToken("Test "+symbol, symbol, 6, nextID.Next(), rlm)
	ledgers[symbol] = led
	return grc20reg.Register(cross(rlm), tok, "")
}

// approveAs produces the allowance a keyed realm's Approve(cur, symbol, spender,
// amount) would produce for `owner`. The ledger is the authority a token realm
// holds over its own token; using it directly is what makes this a faithful
// stand-in for a realm this test cannot deploy.
func approveAs(symbol string, owner address, amount int64) {
	if err := ledgers[symbol].Approve(owner, ProbeAddress(), amount); err != nil {
		panic(err)
	}
}

// ---------------------------------------------------------------- convention

func TestUnrecordedTokenFallsBackToCanonical(cur realm, t *testing.T) {
	key := mint(0, cur, "CANON")

	r, source := Get(key)
	uassert.Equal(t, SourceConvention, source, "nothing recorded for this token")
	uassert.Equal(t, CanonicalApprove, r.Func(OpApprove))
	uassert.Equal(t, CanonicalTransfer, r.Func(OpTransfer))
	uassert.Equal(t, CanonicalTransferFrom, r.Func(OpTransferFrom))
	uassert.Equal(t, 0, len(r.Prefix), "canonical shape takes no leading arguments")

	uassert.Equal(t, selfPath, PkgPath(key))
}

// Every token in this file lives on this one realm, so a keyed proof below marks
// the whole realm. The convention cases therefore have to run before it — they
// are first in the file, and gno runs tests in declaration order.
func TestJSONFallsBackAndNamesTheSource(cur realm, t *testing.T) {
	key := mint(0, cur, "JSNUN")

	out := JSON(key)
	uassert.True(t, strings.Contains(out, `"source":"convention"`),
		"a consumer can tell convention from declaration from proof")
	uassert.True(t, strings.Contains(out, `"prefix":[]`))
	uassert.True(t, strings.Contains(out, `"approve":"Approve"`))
}

// ------------------------------------------------------------------ declared

func TestRegisterAndReadBack(cur realm, t *testing.T) {
	key := mint(0, cur, "KEYED")

	got := Register(cross(cur), "KEYED", Keyed("KEYED"))
	uassert.Equal(t, key, got, "Register returns the same key grc20reg uses")

	r, source := Get(key)
	uassert.Equal(t, SourceDeclared, source)
	uassert.Equal(t, "Approve", r.Func(OpApprove))
	urequire.Equal(t, 1, len(r.Prefix))
	uassert.Equal(t, "KEYED", r.Prefix[0])
}

func TestRegisterOverwritesOwnDeclaration(cur realm, t *testing.T) {
	key := mint(0, cur, "UPDATE")

	Register(cross(cur), "UPDATE", Keyed("old"))
	Register(cross(cur), "UPDATE", Keyed("new"))

	r, _ := Get(key)
	urequire.Equal(t, 1, len(r.Prefix))
	uassert.Equal(t, "new", r.Prefix[0], "a realm may correct its own route")
}

func TestCustomEntryPointNames(cur realm, t *testing.T) {
	key := mint(0, cur, "VCHR")

	Register(cross(cur), "VCHR", Route{
		Prefix: []string{"ibc/542B346608D"},
		Funcs: []Entry{
			{OpApprove, "VoucherApprove"},
			{OpTransfer, "VoucherSend"},
		},
	})

	r, _ := Get(key)
	uassert.Equal(t, "VoucherApprove", r.Func(OpApprove))
	uassert.Equal(t, "", r.Func(OpTransferFrom), "an absent entry point is a declaration, not an omission")
	uassert.Equal(t, "VoucherApprove", Func(key, OpApprove))
	uassert.False(t, strings.Contains(JSON(key), "transfer_from"),
		"an operation the realm does not expose is simply absent")
}

// The point of an open operation set: wugnot-style wrapper calls are user writes
// too, and a fixed approve/transfer/transferFrom record could never name them.
func TestOperationsBeyondTheGrc20Three(cur realm, t *testing.T) {
	key := mint(0, cur, "WRAP")

	Register(cross(cur), "WRAP", Canonical().
		With("deposit", "Deposit").
		With("withdraw", "Withdraw"))

	uassert.Equal(t, "Deposit", Func(key, "deposit"))
	uassert.Equal(t, "Withdraw", Func(key, "withdraw"))
	uassert.Equal(t, "Approve", Func(key, OpApprove), "the standard three survive With")

	// Entries are sorted, so the rendered object is stable.
	uassert.True(t, strings.Contains(JSON(key),
		`"funcs":{"approve":"Approve","deposit":"Deposit","transfer":"Transfer","transfer_from":"TransferFrom","withdraw":"Withdraw"}`))
}

func TestWithReplacesAnExistingBinding(cur realm, t *testing.T) {
	key := mint(0, cur, "REPL")

	Register(cross(cur), "REPL", Canonical().With(OpApprove, "ApproveV2"))

	uassert.Equal(t, "ApproveV2", Func(key, OpApprove))
	uassert.Equal(t, "Transfer", Func(key, OpTransfer))
}

func TestRegisterRequiresGrc20reg(cur realm, t *testing.T) {
	uassert.AbortsWithMessage(t, cur,
		"grc20routes: not registered in grc20reg: "+selfPath+".GHOST",
		func() { Register(cross(cur), "GHOST", Canonical()) })
}

// A symbol is concatenated into a registry key, so anything that could change how
// fqname.Parse splits that key, or break out of the JSON string, is refused.
// grc20reg's lookup would reject these anyway; this keeps the argument local.
func TestRegisterRejectsCraftedSymbols(cur realm, t *testing.T) {
	mint(0, cur, "SAFE")

	cases := []struct{ symbol, msg string }{
		{"a.b", "grc20routes: invalid symbol character: ."},
		{"a/b", "grc20routes: invalid symbol character: /"},
		{`a"b`, `grc20routes: invalid symbol character: "`},
		{"a:b", "grc20routes: invalid symbol character: :"},
		{"a#b", "grc20routes: invalid symbol character: #"},
		{"", "grc20routes: empty symbol"},
	}
	for _, tc := range cases {
		sym := tc.symbol
		uassert.AbortsWithMessage(t, cur, tc.msg,
			func() { Register(cross(cur), sym, Canonical()) }, sym)
	}
}

// A realm operating under a sub identity keys into grc20reg as "<path>#<sub>",
// but that is not a MsgCall destination. The route reports the callable host.
func TestDestinationStripsSubRealmPath(cur realm, t *testing.T) {
	uassert.Equal(t, "gno.land/r/demo/x", hostOf("gno.land/r/demo/x#admin"))
	uassert.Equal(t, "gno.land/r/demo/x", hostOf("gno.land/r/demo/x"))
}

// -------------------------------------------------------------------- proven

// The case the whole proven path exists for: a realm that hosts several tokens
// and was deployed before this one, so it can never call Register.
func TestKeyedProofCoversEveryTokenOnTheRealm(cur realm, t *testing.T) {
	prover := testutils.TestAddress("prover")
	keyA := mint(0, cur, "PKA")
	keyB := mint(0, cur, "PKB")
	keyC := mint(0, cur, "PKC") // never named in the proof
	testing.SetOriginCaller(prover)

	BeginKeyedProof(cross(cur), keyA, keyB, 111, 222)
	approveAs("PKA", prover, 111)
	approveAs("PKB", prover, 222)
	FinishKeyedProof(cross(cur), keyA, keyB)

	uassert.True(t, IsKeyedRealm(selfPath))

	for _, tc := range []struct{ key, symbol string }{
		{keyA, "PKA"}, {keyB, "PKB"}, {keyC, "PKC"},
	} {
		r, source := Get(tc.key)
		uassert.Equal(t, SourceProven, source, tc.key)
		urequire.Equal(t, 1, len(r.Prefix), tc.key)
		uassert.Equal(t, tc.symbol, r.Prefix[0],
			"one proof covers every token the realm hosts, each keyed by its own symbol")
	}
}

// An unkeyed realm reaches one token per call, so it cannot make two different
// registry keys carry their nonces at once. That is the whole security argument;
// this is it failing as it must.
func TestKeyedProofFailsWhenOnlyOneTokenMoves(cur realm, t *testing.T) {
	prover := testutils.TestAddress("half")
	keyA := mint(0, cur, "HFA")
	keyB := mint(0, cur, "HFB")
	testing.SetOriginCaller(prover)

	BeginKeyedProof(cross(cur), keyA, keyB, 333, 444)
	approveAs("HFA", prover, 333) // only one of the two

	uassert.AbortsWithMessage(t, cur,
		"grc20routes: "+keyB+" does not carry its nonce",
		func() { FinishKeyedProof(cross(cur), keyA, keyB) })
}

// A nonce that was already parked cannot stand in for one the proof caused.
func TestKeyedProofRejectsPreExistingAllowance(cur realm, t *testing.T) {
	prover := testutils.TestAddress("stale")
	keyA := mint(0, cur, "STA")
	keyB := mint(0, cur, "STB")
	approveAs("STA", prover, 555) // parked before the proof opens
	testing.SetOriginCaller(prover)

	uassert.AbortsWithMessage(t, cur,
		"grc20routes: nonce already parked on the probe; pick another",
		func() { BeginKeyedProof(cross(cur), keyA, keyB, 555, 666) })
}

func TestKeyedProofRejectsMismatchedInputs(cur realm, t *testing.T) {
	prover := testutils.TestAddress("bad")
	keyA := mint(0, cur, "BDA")
	keyB := mint(0, cur, "BDB")
	testing.SetOriginCaller(prover)

	uassert.AbortsWithMessage(t, cur, "grc20routes: keyed proof needs two different tokens",
		func() { BeginKeyedProof(cross(cur), keyA, keyA, 1, 2) })
	uassert.AbortsWithMessage(t, cur, "grc20routes: nonces must differ",
		func() { BeginKeyedProof(cross(cur), keyA, keyB, 7, 7) })
	uassert.AbortsWithMessage(t, cur, "grc20routes: nonces must be positive",
		func() { BeginKeyedProof(cross(cur), keyA, keyB, 0, 2) })
	uassert.AbortsWithMessage(t, cur, "grc20routes: no proof in progress for "+selfPath,
		func() { FinishKeyedProof(cross(cur), keyA, keyB) })
}

func TestKeyedProofIsPerProver(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	keyA := mint(0, cur, "PPA")
	keyB := mint(0, cur, "PPB")
	testing.SetOriginCaller(alice)
	BeginKeyedProof(cross(cur), keyA, keyB, 11, 22)

	// Bob cannot finish a proof he did not open, and Alice's allowances are not
	// his to borrow.
	testing.SetOriginCaller(bob)
	uassert.AbortsWithMessage(t, cur, "grc20routes: no proof in progress for "+selfPath,
		func() { FinishKeyedProof(cross(cur), keyA, keyB) })
}

// ------------------------------------------------------------- precedence

func TestDeclaredBeatsProven(cur realm, t *testing.T) {
	prover := testutils.TestAddress("prec")
	keyA := mint(0, cur, "PRA")
	keyB := mint(0, cur, "PRB")

	// The realm states the exact shape for one of its tokens...
	Register(cross(cur), "PRA", Route{Funcs: []Entry{{OpApprove, "ApproveExact"}}})

	// ...and then someone proves the realm is keyed, which would otherwise
	// resolve PRA to Keyed("PRA").
	testing.SetOriginCaller(prover)
	BeginKeyedProof(cross(cur), keyA, keyB, 77, 88)
	approveAs("PRA", prover, 77)
	approveAs("PRB", prover, 88)
	FinishKeyedProof(cross(cur), keyA, keyB)

	r, source := Get(keyA)
	uassert.Equal(t, SourceDeclared, source, "a declaration outranks a proof")
	uassert.Equal(t, "ApproveExact", r.Func(OpApprove))
	uassert.Equal(t, 0, len(r.Prefix))

	_, source = Get(keyB)
	uassert.Equal(t, SourceProven, source, "and does not disturb its siblings")
}

// ------------------------------------------------------------------- output

func TestJSONShape(cur realm, t *testing.T) {
	key := mint(0, cur, "JSN")
	Register(cross(cur), "JSN", Keyed("JSN"))

	want := `{"token_key":"` + key + `",` +
		`"pkg_path":"` + selfPath + `",` +
		`"prefix":["JSN"],"source":"declared",` +
		`"funcs":{"approve":"Approve","transfer":"Transfer","transfer_from":"TransferFrom"}}`
	uassert.Equal(t, want, JSON(key))
}

func TestRejectsUnknownToken(cur realm, t *testing.T) {
	// Get is a plain read in this realm, so its refusal is an ordinary panic,
	// not a cross-realm abort.
	uassert.PanicsWithMessage(t, cur,
		"grc20routes: unknown token: gno.land/r/demo/nope.NOPE",
		func() { Get("gno.land/r/demo/nope.NOPE") })
}

// The JSON output is assembled without escaping, so the validator is what keeps
// it well-formed. Anything that could break out of a JSON string must be refused
// at the door.
func TestValidationRejectsUnsafeValues(cur realm, t *testing.T) {
	mint(0, cur, "VALID")

	cases := []struct {
		name  string
		route Route
		msg   string
	}{
		{
			"quote in func name",
			Route{Funcs: []Entry{{OpApprove, `A"pprove`}}},
			`grc20routes: invalid character in approve name: "`,
		},
		{
			"unexported func name",
			Route{Funcs: []Entry{{OpApprove, "approve"}}},
			"grc20routes: approve must be an exported function name",
		},
		{
			"quote in operation name",
			Route{Funcs: []Entry{{`a"b`, "Approve"}}},
			`grc20routes: invalid character in operation name: "`,
		},
		{
			"uppercase operation name",
			Route{Funcs: []Entry{{"Approve", "Approve"}}},
			"grc20routes: invalid character in operation name: A",
		},
		{
			"duplicate operation",
			Route{Funcs: []Entry{{OpApprove, "Approve"}, {OpApprove, "Other"}}},
			"grc20routes: duplicate operation: approve",
		},
		{
			"empty entry point",
			Route{Funcs: []Entry{{OpApprove, ""}}},
			"grc20routes: empty entry point for approve",
		},
		{
			"quote in prefix",
			Route{Funcs: []Entry{{OpApprove, "Approve"}}, Prefix: []string{`a"b`}},
			`grc20routes: invalid prefix character: "`,
		},
		{
			"backslash in prefix",
			Route{Funcs: []Entry{{OpApprove, "Approve"}}, Prefix: []string{`a\b`}},
			`grc20routes: invalid prefix character: \`,
		},
		{
			"no entry point at all",
			Route{},
			"grc20routes: route declares no entry point",
		},
		{
			"empty prefix argument",
			Route{Funcs: []Entry{{OpApprove, "Approve"}}, Prefix: []string{""}},
			"grc20routes: empty prefix argument",
		},
		{
			"too many prefix arguments",
			Route{Funcs: []Entry{{OpApprove, "Approve"}}, Prefix: []string{"a", "b", "c", "d", "e"}},
			"grc20routes: too many prefix arguments",
		},
	}

	for _, tc := range cases {
		r := tc.route
		uassert.AbortsWithMessage(t, cur, tc.msg,
			func() { Register(cross(cur), "VALID", r) }, tc.name)
	}
}

// A route always points at the realm named by its key, on both paths. The realm
// path is never an argument, so there is no input that aims a route elsewhere.
func TestRouteAlwaysPointsAtItsOwnRealm(cur realm, t *testing.T) {
	key := mint(0, cur, "OWNED")
	Register(cross(cur), "OWNED", Canonical())

	uassert.Equal(t, selfPath, PkgPath(key))
	uassert.True(t, strings.HasPrefix(key, selfPath+"."))
}
