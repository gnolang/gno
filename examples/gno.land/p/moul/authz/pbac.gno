package authz

import (
	"std"
)

// XXX: switch to avl.Tree
// XXX: equivalent of Ownable -> way of Transfering easily
// XXX: helper to transform an Authority into a Membership so that we can expect people to expose their teams' authority
// XXX: Actionable? Doer?
// XXX: make Authority an interface.
// XXX: rename Authority to something else?
// XXX: rename action -> perm/role

func NewPBAC() (*PBAC, *PBACPrivateConfig) {
	// XXX: take `admins std.Address...` OR `admins Membership`
	pbac := PBAC{}
	return &pbac, &pbac.cfg
}

// PBAC is the core interface for permission-based access control.
// It provides methods to check and enforce access permissions.
type PBAC struct {
	cfg PBACPrivateConfig
}

const Admin = "_admin_"

// Can checks if the given address is authorized to perform the action.
func (p *PBAC) Can(addr std.Address, action string) bool {
	return p.cfg.MembershipFor(action).Has(addr)
}

// CanByCurrent checks if the current realm is authorized to perform an action.
//
// Example: if authz.CanByCurrent(auth, "propose") {}
func (p *PBAC) CanByCurrent(action string) bool {
	addr := std.CurrentRealm().Address()
	return p.Can(addr, action)
}

// CanByPrevious checks if the previous realm is authorized to perform an
// action.
//
// Example: if authz.CanByPrevious(auth, "propose") {}
func (p *PBAC) CanByPrevious(action string) bool {
	addr := std.PreviousRealm().Address()
	return p.Can(addr, action)
}

// AssertCan panics if the caller is not authorized to perform an action.
//
// Example: authz.AssertCan(auth, caller, "vote")
func (p *PBAC) AssertCan(addr std.Address, action string) {
	if !p.cfg.MembershipFor(action).Has(addr) {
		panic("unauthorized: " + addr.String() + " cannot " + action)
	}
}

// MembershipFor returns a safe Membership for the specified action.
//
// Example: teamDevs := teamrlm.Authority.MembershipFor("developers")
func (p *PBAC) MembershipFor(action string) Membership {
	return &membershipFromPBAC{p, action}
}

type membershipFromPBAC struct {
	pbac   *PBAC
	action string
}

func (m *membershipFromPBAC) Has(addr std.Address) bool {
	return m.pbac.Can(addr, m.action)
}

// AssertCanByCurrent panics if the current realm is not authorized to perform
// an action.
//
// Example: authz.AssertCanByCurrent(auth, "propose")
func (p *PBAC) AssertCanByCurrent(action string) {
	addr := std.CurrentRealm().Address()
	p.AssertCan(addr, action)
}

// AssertCanByPrevious panics if the previous realm is not authorized to perform
// an action.
//
// Example: authz.AssertCanByPrevious(auth, "propose")
func (p *PBAC) AssertCanByPrevious(action string) {
	addr := std.PreviousRealm().Address()
	p.AssertCan(addr, action)
}

// DoByCurrent implements the Actionable interface which is usually used in
// asynchronous patterns; allowing to start an ownership with a static
// memberset, and later switching to a DAO. A.k.a., progressive decentralization.
func (p *PBAC) DoByCurrent(action string, fn func() error) error {
	p.AssertCanByCurrent(action)
	return fn()
}

// DoByPrevious implements the Actionable interface.
func (p *PBAC) DoByPrevious(action string, fn func() error) error {
	p.AssertCanByPrevious(action)
	return fn()
}

// PBACPrivateConfig
//

// PBACPrivateConfig is the private configuration for the PBAC.
// It should not be exposed.
type PBACPrivateConfig struct {
	memberships map[string]Membership // action -> Membership
}

// Set sets the Membership for the specified action.
func (p *PBACPrivateConfig) Set(action string, membership Membership) {
	p.memberships[action] = membership
}

// Del removes the Membership for the specified action if it exists.
func (p *PBACPrivateConfig) Del(action string) {
	delete(p.memberships, action)
}

// MembershipFor returns the private Membership for the specified action.
func (p *PBACPrivateConfig) MembershipFor(action string) Membership {
	return p.memberships[action]
}
