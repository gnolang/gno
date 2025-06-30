package authz

import (
	"std"
	
	"gno.land/p/moul/addrset"
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
	pbac := PBAC{
		cfg: PBACPrivateConfig{
			memberships: make(map[string]Membership),
		},
	}
	return &pbac, &pbac.cfg
}

// New creates a new PBAC instance and returns it as Authority interface.
// This provides backward compatibility with the test examples.
func New() (Authority, *PBACPrivateConfig) {
	pbac, cfg := NewPBAC()
	return pbac, cfg
}

// Authority interface defines the methods for authorization control.
type Authority interface {
	Can(addr std.Address, action string) bool
	CurrentCan(action string) bool
	PreviousCan(action string) bool
	AssertCan(addr std.Address, action string)
	AssertCurrentCan(action string)
	AssertPreviousCan(action string)
	MembershipFor(action string) Membership
}

// Ensure PBAC implements Authority
var _ Authority = (*PBAC)(nil)

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

// CurrentCan checks if the current realm is authorized to perform an action.
//
// Example: if auth.CurrentCan("propose") {}
func (p *PBAC) CurrentCan(action string) bool {
	addr := std.CurrentRealm().Address()
	return p.Can(addr, action)
}

// PreviousCan checks if the previous realm is authorized to perform an
// action.
//
// Example: if auth.PreviousCan("propose") {}
func (p *PBAC) PreviousCan(action string) bool {
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

// AssertCurrentCan panics if the current realm is not authorized to perform
// an action.
//
// Example: auth.AssertCurrentCan("propose")
func (p *PBAC) AssertCurrentCan(action string) {
	addr := std.CurrentRealm().Address()
	p.AssertCan(addr, action)
}

// AssertPreviousCan panics if the previous realm is not authorized to perform
// an action.
//
// Example: auth.AssertPreviousCan("propose")
func (p *PBAC) AssertPreviousCan(action string) {
	addr := std.PreviousRealm().Address()
	p.AssertCan(addr, action)
}

// DoByCurrent implements the Actionable interface which is usually used in
// asynchronous patterns; allowing to start an ownership with a static
// memberset, and later switching to a DAO. A.k.a., progressive decentralization.
func (p *PBAC) DoByCurrent(action string, fn func() error) error {
	p.AssertCurrentCan(action)
	return fn()
}

// DoByPrevious implements the Actionable interface.
func (p *PBAC) DoByPrevious(action string, fn func() error) error {
	p.AssertPreviousCan(action)
	return fn()
}

// PBACPrivateConfig
//

// PBACPrivateConfig is the private configuration for the PBAC.
// It should not be exposed.
type PBACPrivateConfig struct {
	memberships map[string]Membership // action -> Membership
	// XXX: or map[string]addrset.Set?; or both?
	// XXX: filters (dynamic And()/Or()), for instance to allow a high-level
	//      police DAO.
}

// Set sets the Membership for the specified action.
func (p *PBACPrivateConfig) Set(action string, membership Membership) {
	p.memberships[action] = membership
}

// Add adds addresses or a membership to the specified action.
func (p *PBACPrivateConfig) Add(action string, addrsOrMembership ...interface{}) {
	existing := p.memberships[action]
	
	// Handle different types of arguments
	for _, arg := range addrsOrMembership {
		switch v := arg.(type) {
		case string:
			// Treat as std.Address
			addr := std.Address(v)
			if existing == nil {
				existing = addrset.New(addr)
			} else {
				// Try to add to existing addrset, or create composite
				if existingSet, ok := existing.(*addrset.Set); ok {
					existingSet.Add(addr)
				} else {
					// Create composite membership
					existing = Or(existing, addrset.New(addr))
				}
			}
		case std.Address:
			if existing == nil {
				existing = addrset.New(v)
			} else {
				if existingSet, ok := existing.(*addrset.Set); ok {
					existingSet.Add(v)
				} else {
					existing = Or(existing, addrset.New(v))
				}
			}
		case Membership:
			if existing == nil {
				existing = v
			} else {
				existing = Or(existing, v)
			}
		}
	}
	
	p.memberships[action] = existing
}

// Del removes the Membership for the specified action if it exists.
func (p *PBACPrivateConfig) Del(action string) {
	delete(p.memberships, action)
}

// MembershipFor returns the private Membership for the specified action.
func (p *PBACPrivateConfig) MembershipFor(action string) Membership {
	membership := p.memberships[action]
	if membership == nil {
		return &nilMembership{}
	}
	return membership
}
