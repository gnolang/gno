package simpledao

import (
	"errors"

	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/gov/dao"
)

var (
	errAlreadyMember        = errors.New("address is already a member")
	errMissingMember        = errors.New("address is not a member")
	errInvalidAddressUpdate = errors.New("invalid address update")
)

type MembStoreOption func(*MembStore)

// WithInitialMembers initializes the simpledao member store
// with an initial member list
func WithInitialMembers(members []dao.Member) MembStoreOption {
	return func(store *MembStore) {
		for _, m := range members {
			store.members.Set(m.Address.String(), m)
		}
	}
}

// MembStore implements the dao.MembStore abstraction
type MembStore struct {
	members *avl.Tree // std.Address -> dao.Member

	totalVotingPower uint64 // cached value for quick lookups
	majorityPower    uint64 // cached value for quick lookups
}

// NewMembStore creates a new simpledao member store
func NewMembStore(opts ...MembStoreOption) *MembStore {
	m := &MembStore{
		members:          avl.NewTree(),
		totalVotingPower: 0,
		majorityPower:    0,
	}

	// Apply the options
	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *MembStore) AddMember(member dao.Member) error {
	if !isCallerGOVDAO() {
		return errNotGovDAO
	}

	// Check if the member exists
	if m.IsMember(member.Address) {
		return errAlreadyMember
	}

	// Add the member
	m.members.Set(member.Address.String(), member)

	// Update the total voting power
	m.totalVotingPower += member.VotingPower
	m.majorityPower = (2 * m.totalVotingPower) / 3

	return nil
}

func (m *MembStore) RemoveMember(address std.Address) error {
	if !isCallerGOVDAO() {
		return errNotGovDAO
	}

	// Check if the member exists
	memberRaw, exists := m.members.Get(address.String())
	if !exists {
		return errMissingMember
	}

	member := memberRaw.(dao.Member)

	// Remove the member
	m.members.Remove(address.String())

	m.totalVotingPower -= member.VotingPower
	m.majorityPower = (2 * m.totalVotingPower) / 3

	return nil
}

func (m *MembStore) UpdateMember(address std.Address, member dao.Member) error {
	if !isCallerGOVDAO() {
		return errNotGovDAO
	}

	// Check if the member exists
	memberRaw, exists := m.members.Get(address.String())
	if !exists {
		return errMissingMember
	}

	// Check that the member wouldn't be
	// overwriting an existing one
	isAddressUpdate := address != member.Address
	if isAddressUpdate && m.IsMember(member.Address) {
		return errInvalidAddressUpdate
	}

	// Remove the old member info
	// in case the address changed
	if address != member.Address {
		m.members.Remove(address.String())
	}

	// Save the new member info
	m.members.Set(member.Address.String(), member)

	return nil
}

func (m *MembStore) IsMember(address std.Address) bool {
	_, exists := m.members.Get(address.String())

	return exists
}

func (m *MembStore) GetMember(address std.Address) (dao.Member, error) {
	member, exists := m.members.Get(address.String())
	if !exists {
		return dao.Member{}, errMissingMember
	}

	return member.(dao.Member), nil
}

func (m *MembStore) GetMembers() []dao.Member {
	members := make([]dao.Member, 0, m.members.Size())

	m.members.Iterate("", "", func(_ string, val interface{}) bool {
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

// getMajorityPower returns the majority voting power
// of the member store
func (m *MembStore) getMajorityPower() uint64 {
	return m.majorityPower
}
