// Package authz provides simple and flexible authorization control for
// privileged actions.
package authz

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/addrset"
)

// XXX: try to make the api looking like english phrases:
//          if auth.CurrentCan("propose") {}
//          auth.AssertCurrentCan("propose")
//          xxx.Do("propose", func() {})

// Authority is a simple authorization system that allows to authorize
// access control to privileged actions.
// It's a safe object that can be exported and used by other packages.
// It only provides synchronous authorization checks.
type Authority struct {
	config PrivateConfig
}

// PrivateConfig is the private configuration of the Authority.
// It isn't a safe object and is intended to be kept unexported.
type PrivateConfig struct {
	policies avl.Tree // action -> addrset.Set
}

// SafeConfig is a wrapper for PrivateConfig that provides a safe-object that
// can be exported safely.
// It will allow any admin to add/remove policies.
//
// Expected usage:
//
//	var Auth, privcfg = authz.NewWithCurrent()
//	var Config = privcfg.SafeConfig()
type SafeConfig struct {
	pc PrivateConfig
}

const (
	// Anyone is a special address that can be used to approve any caller
	// address.
	Anyone = std.Address("g1anyone")

	// Anything is a special action that can be used to approve any action.
	Anything = "_anything_"
)

// New creates a new empty Authority.
func New() (*Authority, PrivateConfig) {
	auth := &Authority{}
	return auth, auth.config
}

// NewWithCurrentAdmin creates a new Authority with the current realm as
// an admin.
func NewWithCurrentAdmin() (*Authority, PrivateConfig) {
	auth, config := New()
	config.Add(Anything, std.CurrentRealm().Address())
	return auth, config
}

// NewWithPreviousAdmin creates a new Authority with the previous realm as
// an admin.
func NewWithPreviousAdmin() (*Authority, PrivateConfig) {
	auth, config := New()
	config.Add(Anything, std.PreviousRealm().Address())
	return auth, config
}

// Can checks if the caller is authorized to perform an action.
func (a Authority) Can(addr std.Address, action string) bool {
	return a.config.can(addr, action)
}

// CanByCurrent checks if the current realm is authorized to perform an action.
func (a Authority) CanByCurrent(action string) bool {
	return a.Can(std.CurrentRealm().Address(), action)
}

// CanByPrevious checks if the previous realm is authorized to perform an
// action.
func (a Authority) CanByPrevious(action string) bool {
	return a.Can(std.PreviousRealm().Address(), action)
}

// AssertCan doesn't panic if the caller is authorized to perform an action.
func (a Authority) AssertCan(addr std.Address, action string) {
	if !a.Can(addr, action) {
		panic("unauthorized")
	}
}

// AssertCanByCurrent doesn't panic if the current realm is authorized to
// perform an action.
func (a Authority) AssertCanByCurrent(action string) {
	a.AssertCan(std.CurrentRealm().Address(), action)
}

// AssertCanByPrevious doesn't panic if the previous realm is authorized to
// perform an action.
func (a Authority) AssertCanByPrevious(action string) {
	a.AssertCan(std.PreviousRealm().Address(), action)
}

// can checks if the caller is authorized to perform an action.
func (c PrivateConfig) can(addr std.Address, action string) bool {
	for _, lookup := range []string{action, Anything} {
		value, ok := c.policies.Get(lookup)
		if !ok {
			continue
		}
		addrs := value.(*addrset.Set)
		if addrs.Has(addr) || addrs.Has(Anyone) {
			return true
		}
	}
	return false
}

// Add adds a new policy to the authority.
func (c *PrivateConfig) Add(action string, members ...std.Address) {
	var policy *addrset.Set
	value, ok := c.policies.Get(action)
	if ok {
		policy = value.(*addrset.Set)
	} else {
		policy = &addrset.Set{}
	}
	for _, member := range members {
		policy.Add(member)
	}
	c.policies.Set(action, policy)
}

// Remove removes a policy from the authority.
func (c *PrivateConfig) Remove(action string, members ...std.Address) {
	value, ok := c.policies.Get(action)
	if !ok {
		return
	}
	policy := value.(*addrset.Set)
	for _, member := range members {
		policy.Remove(member)
	}
	// c.policies.Set(action, policy)
	if policy.Size() == 0 {
		c.policies.Remove(action)
	}
}

// RemoveAll removes all policies for a given action.
func (c *PrivateConfig) RemoveAll(action string) {
	c.policies.Remove(action)
}

// SafeConfig returns a safe-object that can be exported safely.
// It will allow any admin to add/remove policies.
func (c *PrivateConfig) SafeConfig() *SafeConfig {
	return &SafeConfig{pc: *c}
}

// Add lets an admin add a new policy to the Authority's PrivateConfig.
func (c *SafeConfig) Add(action string, members ...std.Address) {
	isCallerAdmin := c.pc.can(std.CurrentRealm().Address(), Anything)
	if isCallerAdmin {
		c.pc.Add(action, members...)
	}
}

// Remove lets an admin remove a policy from the Authority's PrivateConfig.
func (c *SafeConfig) Remove(action string, members ...std.Address) {
	isCallerAdmin := c.pc.can(std.CurrentRealm().Address(), Anything)
	if isCallerAdmin {
		c.pc.Remove(action, members...)
	}
}

// RemoveAll lets an admin remove all policies for a given action from the
// Authority's PrivateConfig.
func (c *SafeConfig) RemoveAll(action string) {
	isCallerAdmin := c.pc.can(std.CurrentRealm().Address(), Anything)
	if isCallerAdmin {
		c.pc.RemoveAll(action)
	}
}

/*
func (c *PrivateConfig) GetActionsForAddr(addr std.Address) []string {
	actions := []string{}
	c.policies.Iterate("", "", func(key string, value any) bool {
		if value.(*addrset.Set).Has(addr) {
			actions = append(actions, key)
		}
		return false
	})
	return actions
}
func (c *PrivateConfig) GetAddrsForAction(action string) *addrset.Set {
	value, ok := c.policies.Get(action)
	if !ok {
		return nil
	}
	return value.(*addrset.Set)
}
*/

/*
// the simple case
func Add(...) {
    crossing()
    if authz.CanByPrevious(...) { queue.Add(...) }
}

// auto setup case
func makeDAO(...) {
   ...
   dao.SetAuthorizer(authz)
}

// auto setup async thing with private setup
func makeSuggestionsBox(...) box {
  box, privauth := makeBox(...)
  box.Authorizer() // returns the public part
  privauth.
      Add(Anyone(), "propose").
      Add(std.CurrentRealm().Address(), "admin").
      Add(authz, "admin")
  return box
}
box := makeSuggestionsBox(...)
box.Suggest(suggestion)

*/

/*
// Authorizer is the main wrapper object that handles authority management.
// It is configured with a replaceable Authority implementation.
type Authorizer struct {
	auth Authority
	Can(caller std.Address, actionName string, args ...any) bool
	CanByCurrent(actionName string, args ...any) bool
	CanByPrevious(actionName string, args ...any) bool
	GetPolicyForAction(actionName string) string
}

// Authority represents an entity that can authorize privileged actions.
// It is implemented by MemberAuthority, ContractAuthority, AutoAcceptAuthority,
// and DroppedAuthority.
type Authority interface {
	// Authorize executes a privileged action if the caller is authorized
	// Additional args can be provided for context (e.g., for proposal creation)
	// Authorize(caller std.Address, title string, action PrivilegedAction, args ...any) error

	// Can checks if the caller is authorized to perform an action.
	// Additional args can be provided for context (e.g., for proposal creation).
	Can(caller std.Address, actionName string, args ...any) bool

	// String returns a human-readable description of the authority
	String() string
}

// PrivilegedAction defines a function that performs a privileged action.
type PrivilegedAction func() error

// PrivilegedActionHandler is called by contract-based authorities to handle
// privileged actions.
type PrivilegedActionHandler func(title string, action PrivilegedAction) error

// NewWithCurrent creates a new Authorizer with the auth realm's address as authority
func NewWithCurrent() *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(std.CurrentRealm().Address()),
	}
}

// NewWithPrevious creates a new Authorizer with the previous realm's address as authority
func NewWithPrevious() *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(std.PreviousRealm().Address()),
	}
}

// NewWithCurrent creates a new Authorizer with the auth realm's address as authority
func NewWithMembers(addrs ...std.Address) *Authorizer {
	return &Authorizer{
		auth: NewMemberAuthority(addrs...),
	}
}

// NewWithOrigin creates a new Authorizer with the origin caller's address as
// authority.
// This is typically used in the init() function.
func NewWithOrigin() *Authorizer {
	origin := std.OriginCaller()
	previous := std.PreviousRealm()
	if origin != previous.Address() {
		panic("NewWithOrigin() should be called from init() where std.PreviousRealm() is origin")
	}
	return &Authorizer{
		auth: NewMemberAuthority(origin),
	}
}

// NewWithAuthority creates a new Authorizer with a specific authority
func NewWithAuthority(authority Authority) *Authorizer {
	return &Authorizer{
		auth: authority,
	}
}

// Authority returns the auth authority implementation
func (a *Authorizer) Authority() Authority {
	return a.auth
}

// Transfer changes the auth authority after validation
func (a *Authorizer) Transfer(caller std.Address, newAuthority Authority) error {
	// Ask auth authority to validate the transfer
	return a.auth.Authorize(caller, "transfer_authority", func() error {
		a.auth = newAuthority
		return nil
	})
}

// DoByCurrent executes a privileged action by the auth realm.
func (a *Authorizer) DoByCurrent(title string, action PrivilegedAction, args ...any) error {
	current := std.CurrentRealm()
	caller := current.Address()
	return a.auth.Authorize(caller, title, action, args...)
}

// DoByPrevious executes a privileged action by the previous realm.
func (a *Authorizer) DoByPrevious(title string, action PrivilegedAction, args ...any) error {
	previous := std.CurrentRealm()
	caller := previous.Address()
	return a.auth.Authorize(caller, title, action, args...)
}

// String returns a string representation of the auth authority
func (a *Authorizer) String() string {
	authStr := a.auth.String()

	switch a.auth.(type) {
	case *MemberAuthority:
	case *ContractAuthority:
	case *AutoAcceptAuthority:
	case *droppedAuthority:
	default:
		// this way official "dropped" is different from "*custom*: dropped" (autoclaimed).
		return ufmt.Sprintf("custom_authority[%s]", authStr)
	}
	return authStr
}

// MemberAuthority is the default implementation using addrset for member
// management.
type MemberAuthority struct {
	members addrset.Set
}

func NewMemberAuthority(members ...std.Address) *MemberAuthority {
	auth := &MemberAuthority{}
	for _, addr := range members {
		auth.members.Add(addr)
	}
	return auth
}

func (a *MemberAuthority) Authorize(caller std.Address, title string, action PrivilegedAction, args ...any) error {
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

// AddMember adds a new member to the authority
func (a *MemberAuthority) AddMember(caller std.Address, addr std.Address) error {
	return a.Authorize(caller, "add_member", func() error {
		a.members.Add(addr)
		return nil
	})
}

// AddMembers adds a list of members to the authority
func (a *MemberAuthority) AddMembers(caller std.Address, addrs ...std.Address) error {
	return a.Authorize(caller, "add_members", func() error {
		for _, addr := range addrs {
			a.members.Add(addr)
		}
		return nil
	})
}

// RemoveMember removes a member from the authority
func (a *MemberAuthority) RemoveMember(caller std.Address, addr std.Address) error {
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
func (a *MemberAuthority) Has(addr std.Address) bool {
	return a.members.Has(addr)
}

// ContractAuthority implements async contract-based authority
type ContractAuthority struct {
	contractPath    string
	contractAddr    std.Address
	contractHandler PrivilegedActionHandler
	proposer        Authority // controls who can create proposals
}

func NewContractAuthority(path string, handler PrivilegedActionHandler) *ContractAuthority {
	return &ContractAuthority{
		contractPath:    path,
		contractAddr:    std.DerivePkgAddr(path),
		contractHandler: handler,
		proposer:        NewAutoAcceptAuthority(), // default: anyone can propose
	}
}

// NewRestrictedContractAuthority creates a new contract authority with a
// proposer restriction.
func NewRestrictedContractAuthority(path string, handler PrivilegedActionHandler, proposer Authority) Authority {
	if path == "" {
		panic("contract path cannot be empty")
	}
	if handler == nil {
		panic("contract handler cannot be nil")
	}
	if proposer == nil {
		panic("proposer cannot be nil")
	}
	return &ContractAuthority{
		contractPath:    path,
		contractAddr:    std.DerivePkgAddr(path),
		contractHandler: handler,
		proposer:        proposer,
	}
}

func (a *ContractAuthority) Authorize(caller std.Address, title string, action PrivilegedAction, args ...any) error {
	if a.contractHandler == nil {
		return errors.New("contract handler is not set")
	}

	// setup a once instance to ensure the action is executed only once
	executionOnce := once.Once{}

	// Wrap the action to ensure it can only be executed by the contract
	wrappedAction := func() error {
		current := std.CurrentRealm().Address()
		if current != a.contractAddr {
			return errors.New("action can only be executed by the contract")
		}
		return executionOnce.DoErr(func() error {
			return action()
		})
	}

	// Use the proposer authority to control who can create proposals
	return a.proposer.Authorize(caller, title+"_proposal", func() error {
		if err := a.contractHandler(title, wrappedAction); err != nil {
			return err
		}
		return nil
	}, args...)
}

func (a *ContractAuthority) String() string {
	return ufmt.Sprintf("contract_authority[contract=%s]", a.contractPath)
}

// AutoAcceptAuthority implements an authority that accepts all actions
// AutoAcceptAuthority is a simple authority that automatically accepts all
// actions.
// It can be used as a proposer authority to allow anyone to create proposals.
type AutoAcceptAuthority struct{}

func NewAutoAcceptAuthority() *AutoAcceptAuthority {
	return &AutoAcceptAuthority{}
}

func (a *AutoAcceptAuthority) Authorize(caller std.Address, title string, action PrivilegedAction, args ...any) error {
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

func (a *droppedAuthority) Authorize(caller std.Address, title string, action PrivilegedAction, args ...any) error {
	return errors.New("dropped authority: all actions are denied")
}

func (a *droppedAuthority) String() string {
	return "dropped_authority"
}

*/
