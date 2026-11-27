package vesting

import (
	"chain"
	"strconv"
	"strings"
)

// Grant is one beneficiary's schedule. The hosting realm keeps the
// pointer private: the exported methods mutate, so handing the pointer
// to another realm would hand over the books.
type Grant struct {
	id          int64
	admin       address
	beneficiary address
	label       string
	asset       Asset
	terms       Terms
	revocable   bool
	claimed     int64
	returned    int64
	revoked     bool
	revokedAt   int64
	frozen      int64
	hostPkg     string
	hostAddr    address
	subpath     string
	addr        address
}

// New validates a grant but does not bind its sub-realm address.
func New(hostPkg string, hostAddr, admin, beneficiary address, asset Asset, terms Terms, revocable bool, label string) (*Grant, error) {
	if err := validHost(hostPkg, hostAddr); err != nil {
		return nil, err
	}
	if !admin.IsValid() || !beneficiary.IsValid() {
		return nil, ErrAddress
	}
	if admin == hostAddr || beneficiary == hostAddr {
		return nil, ErrSelf
	}
	if !validLabel(label) {
		return nil, ErrLabel
	}
	if err := asset.validate(); err != nil {
		return nil, err
	}
	if err := terms.Validate(); err != nil {
		return nil, err
	}
	return &Grant{
		admin:       admin,
		beneficiary: beneficiary,
		label:       label,
		asset:       asset,
		terms:       terms,
		revocable:   revocable,
		hostPkg:     hostPkg,
		hostAddr:    hostAddr,
	}, nil
}

func validHost(pkg string, addr address) error {
	if !strings.HasPrefix(pkg, "gno.land/r/") || strings.Contains(pkg, "#") || len(pkg) > 200 {
		return ErrHost
	}
	if !addr.IsValid() || addr != chain.PackageAddress(pkg) {
		return ErrHost
	}
	return nil
}

// Bind assigns id and the sub-realm address g/<id>. It does not call Sub.
func (g *Grant) Bind(id int64) (address, error) {
	if g == nil {
		return "", ErrUnbound
	}
	if g.subpath != "" {
		return "", ErrBound
	}
	if id <= 0 {
		return "", ErrID
	}
	sub := "g/" + strconv.FormatInt(id, 10)
	if len(g.hostPkg)+1+len(sub) > 256 {
		return "", ErrHost
	}
	addr := chain.DerivePkgSubAddr(g.hostPkg, sub)
	if addr == g.admin || addr == g.beneficiary || addr == g.hostAddr || !addr.IsValid() {
		return "", ErrSelf
	}
	g.id = id
	g.subpath = sub
	g.addr = addr
	chain.Emit(
		"GrantCreated",
		"grant", strconv.FormatInt(id, 10),
		"admin", g.admin.String(),
		"beneficiary", g.beneficiary.String(),
		"asset", g.asset.Kind.String(),
		"total", strconv.FormatInt(g.terms.Total, 10),
		"curve", g.terms.Curve.String(),
		"revocable", boolText(g.revocable),
		"address", addr.String(),
	)
	return addr, nil
}

func boolText(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// ID is the factory id, or 0 before Bind.
func (g *Grant) ID() int64 {
	if g == nil {
		return 0
	}
	return g.id
}

// Admin is the account that may revoke a revocable grant.
func (g *Grant) Admin() address {
	if g == nil {
		return ""
	}
	return g.admin
}

// Beneficiary is the only account that may claim or name a new beneficiary.
func (g *Grant) Beneficiary() address {
	if g == nil {
		return ""
	}
	return g.beneficiary
}

// Label is the display name. It may be empty.
func (g *Grant) Label() string {
	if g == nil {
		return ""
	}
	return g.label
}

// Asset is the locked coin or token.
func (g *Grant) Asset() Asset {
	if g == nil {
		return Asset{}
	}
	return g.asset
}

// Schedule is a copy of the terms.
func (g *Grant) Schedule() Terms {
	if g == nil {
		return Terms{}
	}
	return g.terms
}

// Revocable reports whether the admin may revoke.
func (g *Grant) Revocable() bool { return g != nil && g.revocable }

// Revoked reports whether the admin has revoked.
func (g *Grant) Revoked() bool { return g != nil && g.revoked }

// RevokedAt is the unix second of the revoke, or 0.
func (g *Grant) RevokedAt() int64 {
	if g == nil {
		return 0
	}
	return g.revokedAt
}

// Claimed is the amount already released to beneficiaries.
func (g *Grant) Claimed() int64 {
	if g == nil {
		return 0
	}
	return g.claimed
}

// Returned is the unvested amount sent back to the admin on revoke.
func (g *Grant) Returned() int64 {
	if g == nil {
		return 0
	}
	return g.returned
}

// Subpath is the cur.Sub argument, empty before Bind.
func (g *Grant) Subpath() string {
	if g == nil {
		return ""
	}
	return g.subpath
}

// Address is the account that holds the grant's funds.
func (g *Grant) Address() address {
	if g == nil {
		return ""
	}
	return g.addr
}

// HostPkg is the factory realm path.
func (g *Grant) HostPkg() string {
	if g == nil {
		return ""
	}
	return g.hostPkg
}

// Vested is the amount unlocked at now. After a revoke it stays at the
// amount that was vested then, including when the clock moves forward.
func (g *Grant) Vested(now int64) (int64, error) {
	if err := g.ready(now); err != nil {
		return 0, err
	}
	return g.vested(now)
}

// Releasable is Vested minus what beneficiaries have already claimed.
func (g *Grant) Releasable(now int64) (int64, error) {
	v, err := g.Vested(now)
	if err != nil {
		return 0, err
	}
	if v < g.claimed {
		return 0, nil
	}
	return subInt(v, g.claimed)
}

// Reclaimable is the unvested amount Revoke would return to the admin.
// It does not write. The realm checks the balance against this before
// Revoke updates the books.
func (g *Grant) Reclaimable(now int64) (int64, error) {
	if err := g.ready(now); err != nil {
		return 0, err
	}
	if !g.revocable {
		return 0, ErrIrrevocable
	}
	if g.revoked {
		return 0, ErrRevoked
	}
	v, err := g.vested(now)
	if err != nil {
		return 0, err
	}
	return subInt(g.terms.Total, v)
}

// Outstanding is the amount the books still expect to be sitting on the
// grant address: total, minus claimed, minus returned to the admin.
func (g *Grant) Outstanding() (int64, error) {
	if g == nil || g.subpath == "" {
		return 0, ErrUnbound
	}
	left, err := subInt(g.terms.Total, g.claimed)
	if err != nil {
		return 0, err
	}
	return subInt(left, g.returned)
}

// View copies the grant and fills Vested and Releasable at now.
func (g *Grant) View(now int64) (GrantView, error) {
	v, err := g.Vested(now)
	if err != nil {
		return GrantView{}, err
	}
	rel, err := g.Releasable(now)
	if err != nil {
		return GrantView{}, err
	}
	return GrantView{
		ID:          g.id,
		Admin:       g.admin,
		Beneficiary: g.beneficiary,
		Label:       g.label,
		Asset:       g.asset,
		Terms:       g.terms,
		Claimed:     g.claimed,
		Returned:    g.returned,
		Revocable:   g.revocable,
		Revoked:     g.revoked,
		RevokedAt:   g.revokedAt,
		Address:     g.addr,
		Vested:      v,
		Releasable:  rel,
	}, nil
}

// Claim releases the releasable amount to the beneficiary. The realm
// moves the coins after this returns. A second call with nothing left
// returns ErrNothing.
func (g *Grant) Claim(caller address, now int64) (int64, error) {
	if err := g.authBeneficiary(caller, now); err != nil {
		return 0, err
	}
	rel, err := g.Releasable(now)
	if err != nil {
		return 0, err
	}
	if rel == 0 {
		return 0, ErrNothing
	}
	next, err := addInt(g.claimed, rel)
	if err != nil {
		return 0, err
	}
	g.claimed = next
	chain.Emit(
		"GrantClaimed",
		"grant", strconv.FormatInt(g.id, 10),
		"beneficiary", caller.String(),
		"amount", strconv.FormatInt(rel, 10),
		"claimed", strconv.FormatInt(g.claimed, 10),
	)
	return rel, nil
}

// Revoke freezes vesting. The returned amount is the part that had not
// vested; the realm sends it to the admin. What had already vested stays
// claimable by the beneficiary. A second revoke fails.
func (g *Grant) Revoke(caller address, now int64) (int64, error) {
	if err := g.authAdmin(caller, now); err != nil {
		return 0, err
	}
	if !g.revocable {
		return 0, ErrIrrevocable
	}
	if g.revoked {
		return 0, ErrRevoked
	}
	v, err := VestedAmount(g.terms, now)
	if err != nil {
		return 0, err
	}
	if v < g.claimed {
		v = g.claimed
	}
	returned, err := subInt(g.terms.Total, v)
	if err != nil {
		return 0, err
	}
	g.revoked = true
	g.revokedAt = now
	g.frozen = v
	g.returned = returned
	chain.Emit(
		"GrantRevoked",
		"grant", strconv.FormatInt(g.id, 10),
		"admin", caller.String(),
		"vested", strconv.FormatInt(v, 10),
		"returned", strconv.FormatInt(returned, 10),
	)
	return returned, nil
}

// SetBeneficiary moves the claim. Only the current beneficiary may call it.
func (g *Grant) SetBeneficiary(caller, next address) error {
	if err := g.authBeneficiary(caller, 0); err != nil {
		return err
	}
	if !next.IsValid() {
		return ErrAddress
	}
	if next == g.beneficiary {
		return ErrSame
	}
	if next == g.addr || next == g.hostAddr {
		return ErrSelf
	}
	prev := g.beneficiary
	g.beneficiary = next
	chain.Emit(
		"BeneficiaryChanged",
		"grant", strconv.FormatInt(g.id, 10),
		"from", prev.String(),
		"to", next.String(),
	)
	return nil
}

// TransferAdmin moves the revoke right. An irrevocable grant still has
// an admin of record, and that record can move, but Revoke keeps failing.
func (g *Grant) TransferAdmin(caller, next address) error {
	if err := g.authAdmin(caller, 0); err != nil {
		return err
	}
	if !next.IsValid() {
		return ErrAddress
	}
	if next == g.admin {
		return ErrSame
	}
	if next == g.addr || next == g.hostAddr {
		return ErrSelf
	}
	prev := g.admin
	g.admin = next
	chain.Emit(
		"AdminTransferred",
		"grant", strconv.FormatInt(g.id, 10),
		"from", prev.String(),
		"to", next.String(),
	)
	return nil
}

func (g *Grant) ready(now int64) error {
	if g == nil || g.subpath == "" {
		return ErrUnbound
	}
	if now < 0 {
		return ErrTime
	}
	return nil
}

func (g *Grant) authBeneficiary(caller address, now int64) error {
	if err := g.ready(now); err != nil {
		return err
	}
	if caller != g.beneficiary {
		return ErrNotBeneficiary
	}
	return nil
}

func (g *Grant) authAdmin(caller address, now int64) error {
	if err := g.ready(now); err != nil {
		return err
	}
	if caller != g.admin {
		return ErrNotAdmin
	}
	return nil
}

func (g *Grant) vested(now int64) (int64, error) {
	if g.revoked {
		if g.frozen < g.claimed {
			return g.claimed, nil
		}
		return g.frozen, nil
	}
	v, err := VestedAmount(g.terms, now)
	if err != nil {
		return 0, err
	}
	if v < g.claimed {
		return g.claimed, nil
	}
	return v, nil
}
