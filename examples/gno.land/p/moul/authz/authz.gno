// Package authz provides flexible authorization control for privileged actions.
//
// # Authorization Strategies
//
// The package supports multiple authorization strategies:
//   - Member-based: Single user or team of users
//   - Contract-based: The contract at a given path is itself the authority
//   - Auto-accept: Allow all actions
//   - Drop: Deny all actions
//
// Core Components
//
//   - Authority interface: Base interface implemented by all authorities
//   - Authorizer: Main wrapper object for authority management
//   - MemberAuthority: Manages authorized addresses
//   - ContractAuthority: Makes the contract at a path its own authority
//   - AutoAcceptAuthority: Accepts all actions
//   - DroppedAuthority: Denies all actions
//
// Quick Start
//
//	// Initialize with contract deployer as authority
//	var member address(...)
//	var auth = authz.NewWithMembers(member)
//
//	// Create functions that require authorization
//	func UpdateConfig(cur realm, newValue string) error {
//		return auth.DoByPrevious(0, cur, "update_config", func() error {
//			config = newValue
//			return nil
//		})
//	}
//
// See example_test.gno for more usage examples.
package authz

import (
	"chain"
	"errors"
	"strings"

	"gno.land/p/moul/addrset/v0"
	"gno.land/p/moul/once/v0"
	"gno.land/p/nt/avl/rotree/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
)

// Authorizer is the main wrapper object that handles authority management.
// It is configured with a replaceable Authority implementation.
type Authorizer struct {
	auth Authority
}

// Authority represents an entity that can authorize privileged actions.
// It is implemented by MemberAuthority, ContractAuthority, AutoAcceptAuthority,
// and DroppedAuthority.
//
// Authority is the canonical safe shape for cross-package authority
// interfaces: methods are address-typed (no realm/cur crosses the interface
// boundary), and consumers correctly derive `caller` from
// `cur.Previous().Address()` under `rlm.IsCurrent()` before invoking
// Authorize. No cur-leak (class 1) is possible through this interface.
//
// However, two RESIDUAL RISKS apply:
//
//   - Class-3 impl-substitution: NewWithAuthority and Authorizer.Transfer
//     accept any Authority impl. A malicious Authority can always-approve
//     (silent privilege escalation) or always-deny (denial-of-service).
//     Consumers should pass canonical impls from this package
//     (MemberAuthority, ContractAuthority, AutoAcceptAuthority,
//     DroppedAuthority) unless they have explicit reason to register a
//     foreign impl. We do not expose an IsCanonicalAuthority allowlist
//     because the package is intentionally extensible — third-party impls
//     are the design intent.
//
//   - Class-4 closed-over-authority: NewContractAuthority and
//     NewRestrictedContractAuthority capture a caller-supplied
//     PrivilegedActionHandler closure. The handler runs synchronously
//     inside Authorize with the consumer's authority. A hostile handler
//     can swallow actions, log the caller, or execute arbitrary code
//     under the consumer's frame. Register only trusted handler functions.
//     See r/gnops/valopers/init.gno for the realistic registration shape.
//
//   - Caller- and title-forgery on the RAW interface: Authorize takes
//     `caller` and `title` as ARGUMENTS, not from the frame. A consumer
//     that exposes a bare Authority (rather than the *Authorizer that
//     wraps it) therefore lets any holder present any caller and any
//     title. Caller-forgery is contained — the only authority-mutating
//     closure lives inside Authorizer.Transfer, so a forged caller on a
//     raw Authorize runs the caller's own inert closure and cannot
//     transfer; TestForgedCallerCannotTransfer pins that boundary.
//     Title-forgery is NOT contained: the title is passed straight to the
//     contractHandler, so a handler that branches on it (routing, quotas,
//     audit trails) must not treat it as trusted. Keep the Authority
//     unexported and hand out only what callers need; see
//     r/gnops/valopers/admin.gno, which exports a description string.
//
// We do NOT seal Authority via an unexported marker method — that pattern
// is bypassable via embedding in Gno; see
// p/test/seal/filetests/z_seal_*_filetest.gno for the four bypass tests.
type Authority interface {
	// Authorize executes a privileged action if the caller is authorized
	// Additional args can be provided for context (e.g., for proposal creation)
	Authorize(caller address, title string, action PrivilegedAction, args ...any) error

	// String returns a human-readable description of the authority
	String() string
}

// PrivilegedAction defines a function that performs a privileged action.
type PrivilegedAction func() error

// PrivilegedActionHandler is called by contract-based authorities to handle
// privileged actions.
type PrivilegedActionHandler func(title string, action PrivilegedAction) error

// NewWithMembers creates a new Authorizer whose authority is a
// MemberAuthority containing the given addresses. Callers express
// authority intent at the call site:
//
//	// "auth realm is the authority"
//	a := authz.NewWithMembers(cur.Address())
//
//	// "previous realm is the authority" (from a crossing function)
//	a := authz.NewWithMembers(cur.Previous().Address())
//
//	// "EOA caller is the authority" (from init(cur realm))
//	if !cur.Previous().IsUserCall() {
//	    panic("realm must be initialized by EOA")
//	}
//	a := authz.NewWithMembers(cur.Previous().Address())
//
// This replaces the previous NewWithCurrent / NewWithPrevious /
// NewWithOrigin sugar — those baked runtime.{Current,Previous,Origin}
// reads into the constructor, which (a) prevented use from package-
// level var initializers, (b) made the EOA-origin check inside
// NewWithOrigin an indirect address comparison rather than the
// straightforward IsUserCall predicate, and (c) coupled the
// constructor to the runtime walks the rest of the migration is
// moving away from.
func NewWithMembers(addrs ...address) *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(addrs...),
	}
}

// NewWithAuthority creates a new Authorizer with a specific authority.
//
// SECURITY: `authority` is an open-interface input — any value satisfying
// Authority is accepted. A malicious impl can always-approve (privilege
// escalation) or always-deny (DoS). Prefer canonical impls from this
// package (NewMemberAuthority, NewContractAuthority, NewAutoAcceptAuthority,
// NewDroppedAuthority) unless you specifically need a foreign impl.
func NewWithAuthority(authority Authority) *Authorizer {
	return &Authorizer{
		auth: authority,
	}
}

// Authority returns the auth authority implementation
func (a *Authorizer) Authority() Authority {
	return a.auth
}

// Transfer changes the auth authority after validation. rlm must be the
// caller's own captured cur (asserted via rlm.IsCurrent()); the
// principal is rlm.Previous().Address(). Closes the address-parameter
// forgery: an external realm cannot supply Owner() as `caller` to
// bypass the underlying Authority's check.
//
// SECURITY (runtime substitution): once the current authority approves a
// Transfer, the new authority is installed and effective on the next call.
// If an attacker ever becomes the authority — even briefly — they can
// install a permanent DroppedAuthority (DoS) or an AutoAcceptAuthority
// (privilege escalation). Consumers concerned about this should wrap
// Transfer with a one-shot guard or a quorum/cooldown check.
//
// `newAuthority` is also an open-interface input — see NewWithAuthority's
// Class-3 caveat. Pass canonical impls.
func (a *Authorizer) Transfer(_ int, rlm realm, newAuthority Authority) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.auth.Authorize(caller, "transfer_authority", func() error {
		a.auth = newAuthority
		return nil
	})
}

// DoByCurrent executes a privileged action authorized as `rlm`. `rlm`
// must be the caller's own live cur (asserted via rlm.IsCurrent());
// the authorized principal is `rlm.Address()`. To authorize as the
// realm that called your function, use `DoByPrevious`.
//
//	auth.DoByCurrent(0, cur, "update_config", func() error { ... })    // current realm authorizes
//	auth.DoByPrevious(0, cur, "update_config", func() error { ... })   // calling realm authorizes
//
// The `_ int` first parameter is a deliberate sentinel that pushes
// `rlm realm` past the first-arg position so DoByCurrent stays a
// non-crossing method — otherwise it would be a crossing method and
// rlm.Previous() inside would resolve one realm deeper than the caller
// intended.
//
// SECURITY: the IsCurrent guard closes Class-2 designation forgery (see
// docs/resources/gno-security.md). A realm value's .Address() is set
// when the value is minted at a crossing frame; the value can in
// principle be stored and replayed. Without IsCurrent, a hostile realm
// could capture a high-privilege realm's cur.Previous() (e.g., when
// that realm called into it) and later pass the stored value here to
// authorize actions as that realm. IsCurrent rejects stale captures by
// requiring the value to match the topmost live crossing frame's cur.
func (a *Authorizer) DoByCurrent(_ int, rlm realm, title string, action PrivilegedAction, args ...any) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	return a.auth.Authorize(rlm.Address(), title, action, args...)
}

// DoByPrevious executes a privileged action authorized as the realm
// that called the function invoking DoByPrevious. `rlm` must be the
// caller's own live cur; the principal is derived as
// `rlm.Previous().Address()`. Mirrors the Transfer/AddMember pattern:
// always take live cur, derive the caller-of-caller internally rather
// than accepting a stored/forwarded realm value.
func (a *Authorizer) DoByPrevious(_ int, rlm realm, title string, action PrivilegedAction, args ...any) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	return a.auth.Authorize(rlm.Previous().Address(), title, action, args...)
}

// String returns a string representation of the auth authority.
//
// A non-canonical impl is wrapped as custom_authority[...] so that the
// official "dropped" is distinguishable from a "*custom*: dropped"
// (autoclaimed) one. Shared with ContractAuthority.String, which applies
// the same rule to its nested proposer.
func (a *Authorizer) String() string {
	return canonicalAuthorityString(a.auth)
}

// MemberAuthority is the default implementation using addrset for member
// management.
type MemberAuthority struct {
	members addrset.Set
}

func NewMemberAuthority(members ...address) *MemberAuthority {
	auth := &MemberAuthority{}
	for _, addr := range members {
		auth.members.Add(addr)
	}
	return auth
}

func (a *MemberAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	if !a.members.Has(caller) {
		return errors.New("unauthorized")
	}

	if err := action(); err != nil {
		return err
	}
	return nil
}

func (a *MemberAuthority) String() string {
	addrs := []string{}
	a.members.Tree().Iterate("", "", func(key string, _ any) bool {
		addrs = append(addrs, key)
		return false
	})
	addrsStr := strings.Join(addrs, ",")
	return ufmt.Sprintf("member_authority[%s]", addrsStr)
}

// AddMember adds a new member to the authority. rlm must be the caller's
// own captured cur; the principal is rlm.Previous().Address() and must
// already be a member. The IsCurrent guard closes the forgery where an
// external realm passes Owner() as caller to bypass members.Has(caller).
func (a *MemberAuthority) AddMember(_ int, rlm realm, addr address) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.Authorize(caller, "add_member", func() error {
		a.members.Add(addr)
		return nil
	})
}

// AddMembers adds a list of members to the authority. Same rlm contract
// as AddMember.
func (a *MemberAuthority) AddMembers(_ int, rlm realm, addrs ...address) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.Authorize(caller, "add_members", func() error {
		for _, addr := range addrs {
			a.members.Add(addr)
		}
		return nil
	})
}

// RemoveMember removes a member from the authority. Same rlm contract
// as AddMember.
func (a *MemberAuthority) RemoveMember(_ int, rlm realm, addr address) error {
	if !rlm.IsCurrent() {
		return errors.New("unauthorized")
	}
	caller := rlm.Previous().Address()
	return a.Authorize(caller, "remove_member", func() error {
		a.members.Remove(addr)
		return nil
	})
}

// Tree returns a read-only view of the members tree
func (a *MemberAuthority) Tree() *rotree.ReadOnlyTree {
	tree := a.members.Tree().(*avl.Tree)
	return rotree.Wrap(tree, nil)
}

// Has checks if the given address is a member of the authority
func (a *MemberAuthority) Has(addr address) bool {
	return a.members.Has(addr)
}

// ContractAuthority implements async contract-based authority
type ContractAuthority struct {
	contractPath    string
	contractAddr    address
	contractHandler PrivilegedActionHandler
	proposer        Authority // controls who can create proposals
}

// NewContractAuthority makes the contract at `path` its OWN authority:
// the default proposer accepts only chain.PackageAddress(path), so a
// privileged action proceeds only when driven from that contract's own
// frame.
//
// NewRestrictedContractAuthority does NOT widen this — it REPLACES it.
// See its godoc; with an explicit proposer, `path` is a label and
// contractAddr is never consulted.
//
// The `caller` compared against that address is established upstream by
// Authorizer.DoByCurrent / DoByPrevious / Transfer under rlm.IsCurrent(),
// so an external realm cannot present the contract address.
//
// BREAKING: the default proposer used to be
// NewAutoAcceptAuthority ("anyone can propose"). Consumers that relied on
// arbitrary callers driving a plain ContractAuthority will start getting
// "unauthorized"; that is the fix, not a regression. Pass an explicit
// proposer to NewRestrictedContractAuthority to opt back in.
//
// SECURITY — the gate binds the contract's IDENTITY, not its intent.
// Inside any crossing frame of the contract, rlm.Address() is
// unconditionally PackageAddress(path), so `DoByCurrent` from within the
// contract is a tautology; only DoByPrevious and Transfer gain a real
// check. Consequently ANY exported function of the contract that returns
// a crossing closure (or anything else carrying its frame) hands this
// authority to its caller. Keep privileged closures unexported.
//
// So DO NOT write NewContractAuthority(ownPath) + DoByCurrent. It reads
// like "only I may do this" and compiles to "anyone may do this". Point
// `path` at the principal you actually want to assert — usually the
// governance realm that will drive the action — and use DoByPrevious, so
// the comparison is against whoever crossed in:
//
//	// in r/gnops/valopers/init.gno
//	auth = authz.NewWithAuthority(
//	    authz.NewContractAuthority("gno.land/r/gov/dao", handler),
//	)
//	// in the privileged write
//	auth.DoByPrevious(0, rlm, "update-instructions", action)
//
// Own-path is right only when something OTHER than the contract's own
// frame supplies the caller — i.e. you drive it via DoByPrevious from a
// realm you have deliberately let call in, or via Transfer.
//
// SECURITY — a gate is only as good as the principal it names. Pointing
// `path` at a realm whose frame ANY caller can mint buys nothing. Check,
// for whatever principal you choose, that nothing exported by that realm
// hands out its frame identity:
//
//   - `gno.land/r/gov/dao` is safe to assert ONLY because
//     SimpleExecutor.Execute rejects invocation from outside the proxy
//     (r/gov/dao/types.gno). Before that gate existed, dao's exported
//     NewSimpleExecutor + Execute let any realm mint a frame whose
//     Previous() was r/gov/dao, so this whole pairing authenticated
//     nothing. Do not assume a governance path is unforgeable; verify it.
//   - Whatever the principal, an exported function of YOUR realm whose
//     signature is assignable to a callback type the principal invokes
//     (for r/gov/dao: `func(realm) error`) is itself the leak — the
//     principal will happily call it for an attacker. Keep privileged
//     entrypoints out of that shape, or take extra parameters.
//
// Neither shape defends against re-exporting the privileged closure: its
// holder picks the Previous() they present by wrapping it in an executor
// of their own. Unexported closures remain the load-bearing rule.
//
// A syntactically valid but WRONG `path` is a permanent brick, not an
// error: the authority accepts nobody, and Transfer routes through the
// same gate, so it cannot be rotated out either. Construction rejects
// malformed paths, but it cannot tell a wrong path from a right one —
// e.g. "gno.land/r/gov/dao/impl/v0" (the path r/gov/dao's own
// allowedDAOs list holds) is well-formed and permanently dead, because
// an executor's Previous() is the PROXY path, never the impl. Give the
// consumer a governance-gated rotation entrypoint before deploying; see
// r/gnops/valopers/admin.gno's NewAuthorityRotationProposalRequest.
//
// SECURITY (Class-4 captured callback): `handler` is a caller-supplied
// closure that runs SYNCHRONOUSLY inside Authorize, with the consumer's
// authority. A hostile handler can swallow actions, log the caller, or
// execute arbitrary code under the consumer's frame. The package-internal
// wrappedAction enforces at-most-once invocation and NOTHING ELSE — in
// particular it does not check the caller, and the handler may call it
// however it likes (multiple times, never, out of order). Register only
// trusted handler functions; treat handler registration as the trust
// boundary.
//
// Panics if `path` is empty or malformed, or if `handler` is nil.
func NewContractAuthority(path string, handler PrivilegedActionHandler) *ContractAuthority {
	addr := chain.PackageAddress(assertValidContractPath(path))
	// The default proposer is a real Authority, not an absent one. Encoding
	// the strictest policy as a nil field would mean the secure default is
	// what you get by leaving something OUT: a struct literal that skips
	// this constructor, a persisted value from an older build, or re-adding
	// `if proposer == nil { proposer = NewAutoAcceptAuthority() }` would all
	// silently be wide open again. A one-field struct costs the same as the
	// nil it replaces and cannot be reached by omission.
	return newContractAuthority(path, addr, handler, &contractIdentityAuthority{addr: addr})
}

// NewRestrictedContractAuthority creates a contract authority whose
// proposer REPLACES the contract-identity gate.
//
// This is the escape hatch, not a second lock. `proposer` becomes the ONLY
// authorization decision: contractAddr is not consulted on this path, so
// `path` degrades to a label that appears in String() and asserts nothing.
// NewRestrictedContractAuthority(p, h, NewAutoAcceptAuthority()) is exactly
// the old default behaviour ("anyone can propose"), stated
// explicitly rather than reached by default.
//
// In particular NewRestrictedContractAuthority("gno.land/r/gov/dao", h,
// NewMemberAuthority(alice)) does NOT mean "GovDAO and alice". It means
// "alice, and not GovDAO".
//
// SECURITY:
//   - `handler` is the same Class-4 captured-callback risk as
//     NewContractAuthority — runs synchronously inside Authorize with the
//     consumer's authority. Register only trusted handler functions.
//   - `proposer` is an open-interface input (Class-3 impl-substitution).
//     A hostile proposer Authority can always-approve creation of any
//     proposal, defeating the restriction. Pass canonical impls only.
//     String() renders a non-canonical proposer wrapped as
//     custom_authority[...] so such a substitution is at least visible.
func NewRestrictedContractAuthority(path string, handler PrivilegedActionHandler, proposer Authority) Authority {
	if proposer == nil {
		panic("proposer cannot be nil")
	}
	addr := chain.PackageAddress(assertValidContractPath(path))
	return newContractAuthority(path, addr, handler, proposer)
}

// newContractAuthority is the single construction path: both exported
// constructors validate, then land here, so a guard added once applies to
// both. `handler` is rejected here rather than surfaced at Authorize time
// because Authorize checks contractHandler == nil BEFORE consulting the
// proposer, and Transfer routes through Authorize — so a nil-handler
// authority could never be rotated out. It is a permanent brick with no
// on-chain recovery for anyone: the same bricked-governance failure
// mode, reached by accident.
func newContractAuthority(path string, addr address, handler PrivilegedActionHandler, proposer Authority) *ContractAuthority {
	if handler == nil {
		panic("contract handler cannot be nil")
	}
	return &ContractAuthority{
		contractPath:    path,
		contractAddr:    addr,
		contractHandler: handler,
		proposer:        proposer,
	}
}

// assertValidContractPath rejects paths that chain.PackageAddress would
// hash happily but that no realm could ever present, which would brick the
// authority permanently (nobody authorizes, and Transfer routes through the
// same gate so nobody can rotate out). Returns `path` so it composes.
//
// The authoritative rule is gnolang.ReGnoUserPkgPath ("all paths must be
// lowercase ascii alphanumeric characters", gnovm/pkg/gnolang/mempackage.go),
// but no validator is exported to Gno code — chain's isValidSubpath /
// assertValidSubpath are unexported, and there is no chain.IsValidPkgPath.
// So this is a deliberately permissive superset: segment ("/" segment)*,
// segment = [a-z0-9] ([a-z0-9_.-]* [a-z0-9])?. It accepts every real
// package path and catches what actually gets typed wrong — trailing or
// embedded whitespace, empty segments, uppercase, stray punctuation. It
// CANNOT catch a well-formed path that is simply the wrong principal. See
// NewContractAuthority's godoc on rotation.
func assertValidContractPath(path string) string {
	if path == "" {
		panic("contract path cannot be empty")
	}
	if !strings.Contains(path, "/") {
		panic("contract path must be a package path, e.g. gno.land/r/gov/dao")
	}
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if !isValidContractPathSegment(path[start:i]) {
				panic("contract path must be '/'-separated segments of [a-z0-9] with '_.-' allowed inside a segment, e.g. gno.land/r/gov/dao")
			}
			start = i + 1
		}
	}
	return path
}

func isValidContractPathSegment(seg string) bool {
	if seg == "" {
		return false
	}
	for i := 0; i < len(seg); i++ {
		switch c := seg[i]; {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '_' || c == '.' || c == '-':
			// Allowed only strictly inside a segment.
			if i == 0 || i == len(seg)-1 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// contractIdentityAuthority is the default proposer installed by
// NewContractAuthority: the contract at the bound path, and nobody else.
// Unexported and unmutable by design — AddMember/RemoveMember are
// meaningless for a fixed contract identity, which is why this is not a
// MemberAuthority holding one address.
type contractIdentityAuthority struct {
	addr address
}

func (a *contractIdentityAuthority) Authorize(caller address, _ string, action PrivilegedAction, _ ...any) error {
	if caller != a.addr {
		return errors.New("unauthorized")
	}
	return action()
}

func (a *contractIdentityAuthority) String() string { return "contract-identity" }

func (a *ContractAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	if a.contractHandler == nil {
		return errors.New("contract handler is not set")
	}

	// setup a once instance to ensure the action is executed only once
	executionOnce := once.Once{}

	// wrappedAction enforces at-most-once invocation. The previous
	// gate `unsafe.CurrentRealm() == contractAddr` is removed: it
	// was .Title()-bypassable (runtime.CurrentRealm walks past
	// non-crossing frames to the most-recent crossing ancestor) and
	// the trust boundary is now upstream — Authorizer.DoByCurrent /
	// DoByPrevious require rlm.IsCurrent() and pass a non-forgeable
	// principal to Authorize, while the consumer realm's handler
	// closure is the Class-4 trust root by lexical capture at
	// registration time.
	wrappedAction := func() error {
		return executionOnce.DoErr(func() error {
			return action()
		})
	}

	handle := func() error {
		if err := a.contractHandler(title, wrappedAction); err != nil {
			return err
		}
		return nil
	}

	// The proposer IS the authorization decision. For an authority built by
	// NewContractAuthority it is a contractIdentityAuthority bound to
	// contractAddr — the contract itself and nobody else; for one built by
	// NewRestrictedContractAuthority it is whatever the consumer installed,
	// and contractAddr is deliberately not consulted (see that constructor's
	// godoc). `caller` is established upstream by Authorizer.DoByCurrent /
	// DoByPrevious / Transfer under rlm.IsCurrent(), so an external realm
	// cannot present an arbitrary principal here.
	//
	// A nil proposer is not a policy, it is a malformed value: both
	// constructors always install one, so nil means the struct was built by
	// a literal that bypassed them. Fail closed rather than dereference.
	if a.proposer == nil {
		return errors.New("proposer is not set")
	}
	return a.proposer.Authorize(caller, title+"_proposal", handle, args...)
}

// String renders the contract path AND the proposer.
//
// The proposer half is security-relevant, not cosmetic: it is the only
// thing distinguishing a gated authority from a wide-open one. Rendering
// the path alone made
//
//	NewContractAuthority(path, handler)                              // gated
//	NewRestrictedContractAuthority(path, handler, AutoAccept{})      // open to all
//
// byte-identical, so any consumer test asserting on this string — and
// any on-chain reader inspecting it — was blind to the difference. A
// consumer realm could have its authority swapped for a fully permissive
// one and its assertions would stay green.
//
// "contract-identity" names the default installed by NewContractAuthority:
// the contract at contractPath, and nobody else, may drive this authority.
//
// The proposer is rendered through canonicalAuthorityString, NOT by calling
// a.proposer.String() directly. Authority is an open interface, so a foreign
// impl can return any text it likes — including "contract-identity". That
// made a fully permissive authority byte-identical to the gated default
// again, defeating the very assertions this rendering exists to support.
// Non-canonical impls are wrapped as custom_authority[...], mirroring what
// Authorizer.String has always done at the outer level.
//
// Read the `contract=` half with care: it is load-bearing only for the
// contract-identity default. With any other proposer it is a label —
// see NewRestrictedContractAuthority.
func (a *ContractAuthority) String() string {
	return ufmt.Sprintf(
		"contract_authority[contract=%s,proposer=%s]",
		a.contractPath,
		canonicalAuthorityString(a.proposer),
	)
}

// canonicalAuthorityString renders an Authority, wrapping any
// implementation that is not one of this package's own as
// custom_authority[...] so a foreign impl cannot impersonate a canonical
// one by choosing its String() text. Shared by ContractAuthority.String
// (for the nested proposer) and Authorizer.String (for the installed
// authority).
func canonicalAuthorityString(auth Authority) string {
	if auth == nil {
		// Only reachable via a struct literal that bypassed the
		// constructors; Authorize fails closed on the same condition.
		return "<unset>"
	}
	switch auth.(type) {
	case *MemberAuthority, *ContractAuthority, *AutoAcceptAuthority,
		*droppedAuthority, *contractIdentityAuthority:
		return auth.String()
	default:
		return ufmt.Sprintf("custom_authority[%s]", auth.String())
	}
}

// AutoAcceptAuthority implements an authority that accepts all actions
// AutoAcceptAuthority is a simple authority that automatically accepts all
// actions.
// It can be used as a proposer authority to allow anyone to create proposals.
type AutoAcceptAuthority struct{}

func NewAutoAcceptAuthority() *AutoAcceptAuthority {
	return &AutoAcceptAuthority{}
}

func (a *AutoAcceptAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	return action()
}

func (a *AutoAcceptAuthority) String() string {
	return "auto_accept_authority"
}

// droppedAuthority implements an authority that denies all actions
type droppedAuthority struct{}

func NewDroppedAuthority() Authority {
	return &droppedAuthority{}
}

func (a *droppedAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	return errors.New("dropped authority: all actions are denied")
}

func (a *droppedAuthority) String() string {
	return "dropped_authority"
}
