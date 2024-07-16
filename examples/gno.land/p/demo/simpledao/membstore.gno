package simpledao

import (
	"errors"

	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/gov/dao"
)

var (
	ErrAlreadyMember        = errors.New("address is already a member")
	ErrMissingMember        = errors.New("address is not a member")
	ErrInvalidAddressUpdate = errors.New("invalid address update")
)

type MembStoreOption func(*MembStore)

// WithInitialMembers initializes the simpledao member store
// with an initial member list
func WithInitialMembers(members []dao.Member) MembStoreOption {
	return func(store *MembStore) {
		for _, m := range members {
			store.members.Set(m.Address.String(), m)

			store.totalVotingPower += m.VotingPower
		}
	}
}

// MembStore implements the dao.MembStore abstraction
type MembStore struct {
	members          *avl.Tree // std.Address -> dao.Member
	totalVotingPower uint64    // cached value for quick lookups
}

// NewMembStore creates a new simpledao member store
func NewMembStore(opts ...MembStoreOption) *MembStore {
	m := &MembStore{
		members:          avl.NewTree(),
		totalVotingPower: 0,
	}

	// Apply the options
	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *MembStore) AddMember(member dao.Member) error {
	if !isCallerGOVDAO() {
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

func (m *MembStore) UpdateMember(address std.Address, member dao.Member) error {
	if !isCallerGOVDAO() {
		return ErrNotGovDAO
	}

	// Get the member
	oldMember, err := m.GetMember(address)
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

func (m *MembStore) GetMember(address std.Address) (dao.Member, error) {
	member, exists := m.members.Get(address.String())
	if !exists {
		return dao.Member{}, ErrMissingMember
	}

	return member.(dao.Member), nil
}

func (m *MembStore) GetMembers(offset, count uint64) []dao.Member {
	// Calculate the left and right bounds
	if count < 1 || offset >= uint64(m.members.Size()) {
		return []dao.Member{}
	}

	// Limit the maximum number of returned members
	if count > maxRequestMembers {
		count = maxRequestMembers
	}

	// Gather the members
	members := make([]dao.Member, 0)
	m.members.IterateByOffset(
		int(offset),
		int(count),
		func(_ string, val interface{}) bool {
			member := val.(dao.Member)

			// Save the member
			members = append(members, member)

			return false
		})

	return members
}

func (m *MembStore) Size() int {
	return m.members.Size()
}

// getTotalPower returns the total voting power
// of the member store
func (m *MembStore) getTotalPower() uint64 {
	return m.totalVotingPower
}

// We need to include a govdao guard here, even if the
// executor guarantees it, because
// the API of the member store is public and callable
// by anyone who has a reference to the member store instance.
const daoPkgPath = "gno.land/r/gov/dao/v2"

// isCallerGOVDAO returns a flag indicating if the
// current caller context is the active GOVDAO
func isCallerGOVDAO() bool {
	return std.CurrentRealm().PkgPath() == daoPkgPath
}
