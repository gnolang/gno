package simpledao

import (
	"std"

	"gno.land/p/gov/dao"
)

type (
	executeDelegate   func() error
	isExpiredDelegate func() bool
)

type mockExecutor struct {
	executeFn   executeDelegate
	isExpiredFn isExpiredDelegate
}

func (m *mockExecutor) Execute() error {
	if m.executeFn != nil {
		return m.executeFn()
	}

	return nil
}

func (m *mockExecutor) IsExpired() bool {
	if m.isExpiredFn != nil {
		return m.isExpiredFn()
	}

	return false
}

type (
	membersDelegate      func(uint64, uint64) []dao.Member
	sizeDelegate         func() int
	isMemberDelegate     func(std.Address) bool
	totalPowerDelegate   func() uint64
	memberDelegate       func(std.Address) (dao.Member, error)
	addMemberDelegate    func(dao.Member) error
	updateMemberDelegate func(std.Address, dao.Member) error
)

type mockMemberStore struct {
	membersFn      membersDelegate
	sizeFn         sizeDelegate
	isMemberFn     isMemberDelegate
	totalPowerFn   totalPowerDelegate
	memberFn       memberDelegate
	addMemberFn    addMemberDelegate
	updateMemberFn updateMemberDelegate
}

func (m *mockMemberStore) Members(offset, count uint64) []dao.Member {
	if m.membersFn != nil {
		return m.membersFn(offset, count)
	}

	return nil
}

func (m *mockMemberStore) Size() int {
	if m.sizeFn != nil {
		return m.sizeFn()
	}

	return 0
}

func (m *mockMemberStore) IsMember(address std.Address) bool {
	if m.isMemberFn != nil {
		return m.isMemberFn(address)
	}

	return false
}

func (m *mockMemberStore) TotalPower() uint64 {
	if m.totalPowerFn != nil {
		return m.totalPowerFn()
	}

	return 0
}

func (m *mockMemberStore) Member(address std.Address) (dao.Member, error) {
	if m.memberFn != nil {
		return m.memberFn(address)
	}

	return dao.Member{}, nil
}

func (m *mockMemberStore) AddMember(member dao.Member) error {
	if m.addMemberFn != nil {
		return m.addMemberFn(member)
	}

	return nil
}

func (m *mockMemberStore) UpdateMember(address std.Address, member dao.Member) error {
	if m.updateMemberFn != nil {
		return m.updateMemberFn(address, member)
	}

	return nil
}
