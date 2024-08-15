package membstore

import (
	"errors"

	"std"

	"gno.land/p/demo/avl"
)

var (
	ErrAlreadyMember        = errors.New("address is already a member")
	ErrMissingMember        = errors.New("address is not a member")
	ErrInvalidAddressUpdate = errors.New("invalid address update")
	ErrNotGovDAO            = errors.New("caller not correct govdao instance")
)

// maxRequestMembers is the maximum number of
// paginated members that can be requested
const maxRequestMembers = 50

type Option func(*MembStore)

// WithInitialMembers initializes the member store
// with an initial member list
func WithInitialMembers(members []Member) Option {
	return func(store *MembStore) {
		for _, m := range members {
			store.members.Set(m.Address.String(), m)
			store.totalVotingPower += m.VotingPower
		}
	}
}

// WithDAOPkgPath initializes the member store
// with a dao package path guard
func WithDAOPkgPath(daoPkgPath string) Option {
	return func(store *MembStore) {
		store.daoPkgPath = daoPkgPath
	}
}

// MembStore implements the dao.MembStore abstraction
type MembStore struct {
	daoPkgPath       string    // active dao pkg path, if any
	members          *avl.Tree // std.Address -> Member
	totalVotingPower uint64    // cached value for quick lookups
}

// NewMembStore creates a new member store
func NewMembStore(opts ...Option) *MembStore {
	m := &MembStore{
		members:          avl.NewTree(), // empty set
		daoPkgPath:       "",            // no dao guard
		totalVotingPower: 0,
	}

	// Apply the options
	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *MembStore) AddMember(member Member) error {
	if !m.isCallerDAORealm() {
		return ErrNotGovDAO
	}

	// Check if the member exists
	if m.IsMember(member.Address) {
		return ErrAlreadyMember
	}

	// Add the member
	m.members.Set(member.Address.String(), member)

	// Update the total voting power
	m.totalVotingPower += member.VotingPower

	return nil
}

func (m *MembStore) UpdateMember(address std.Address, member Member) error {
	if !m.isCallerDAORealm() {
		return ErrNotGovDAO
	}

	// Get the member
	oldMember, err := m.Member(address)
	if err != nil {
		return err
	}

	// Check if this is a removal request
	if member.VotingPower == 0 {
		m.members.Remove(address.String())

		// Update the total voting power
		m.totalVotingPower -= oldMember.VotingPower

		return nil
	}

	// Check that the member wouldn't be
	// overwriting an existing one
	isAddressUpdate := address != member.Address
	if isAddressUpdate && m.IsMember(member.Address) {
		return ErrInvalidAddressUpdate
	}

	// Remove the old member info
	// in case the address changed
	if address != member.Address {
		m.members.Remove(address.String())
	}

	// Save the new member info
	m.members.Set(member.Address.String(), member)

	// Update the total voting power
	difference := member.VotingPower - oldMember.VotingPower
	m.totalVotingPower += difference

	return nil
}

func (m *MembStore) IsMember(address std.Address) bool {
	_, exists := m.members.Get(address.String())

	return exists
}

func (m *MembStore) Member(address std.Address) (Member, error) {
	member, exists := m.members.Get(address.String())
	if !exists {
		return Member{}, ErrMissingMember
	}

	return member.(Member), nil
}

func (m *MembStore) Members(offset, count uint64) []Member {
	// Calculate the left and right bounds
	if count < 1 || offset >= uint64(m.members.Size()) {
		return []Member{}
	}

	// Limit the maximum number of returned members
	if count > maxRequestMembers {
		count = maxRequestMembers
	}

	// Gather the members
	members := make([]Member, 0)
	m.members.IterateByOffset(
		int(offset),
		int(count),
		func(_ string, val interface{}) bool {
			member := val.(Member)

			// Save the member
			members = append(members, member)

			return false
		})

	return members
}

func (m *MembStore) Size() int {
	return m.members.Size()
}

// TotalPower returns the total voting power
// of the member store
func (m *MembStore) TotalPower() uint64 {
	return m.totalVotingPower
}

// isCallerDAORealm returns a flag indicating if the
// current caller context is the active DAO Realm.
// We need to include a dao guard, even if the
// executor guarantees it, because
// the API of the member store is public and callable
// by anyone who has a reference to the member store instance.
func (m *MembStore) isCallerDAORealm() bool {
	return m.daoPkgPath == "" || std.CurrentRealm().PkgPath() == m.daoPkgPath
}
