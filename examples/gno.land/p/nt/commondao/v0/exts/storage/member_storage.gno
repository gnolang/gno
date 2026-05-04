package storage

import (
	"gno.land/p/jeronimoalbi/message"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/commondao/v0"
)

const (
	msgMemberAdd    message.Topic = "MemberAdd"
	msgMemberRemove               = "MemberRemove"
)

// GetMemberGroups returns the groups that a member belongs to.
// It returns no groups if member storage was not created using this package.
func GetMemberGroups(s commondao.MemberStorage, member address) []string {
	storage, ok := s.(*memberStorage)
	if !ok {
		return nil
	}

	v, found := storage.uniqueMembers.Get(member.String())
	if !found {
		return nil
	}

	tree, ok := v.(*avl.Tree)
	if !ok {
		// Member doesn't belong to any group
		return nil
	}

	groups := make([]string, 0, tree.Size())
	tree.Iterate("", "", func(group string, _ any) bool {
		groups = append(groups, group)
		return false
	})
	return groups
}

// NewMemberStorage creates a new CommonDAO member storage with grouping support.
//
// This is a custom storage that automatically keeps track of unique members that are
// added or removed from any of the member groups or base storage. This allows for quick
// and inexpensive checks for the number of total unique storage users and also to
// iterate all of them without needing to iterate individual groups.
func NewMemberStorage() commondao.MemberStorage {
	// Create a new broker to allow storages to publish and subscribe to messages
	// It is used to add/remove users from the storage each time one or more groups change.
	broker := message.NewBroker()

	// Define a factory for creating custom member storages when groups are created.
	// Custom storage publishes when a member is added or removed from a group.
	innerFactory := func(group string) commondao.MemberStorage {
		return &groupMemberStorage{
			MemberStorage: commondao.NewMemberStorage(),
			messages:      broker,
			group:         group,
		}
	}

	// Create a member storage that automatically adds or removes members each time
	// a member group changes. This allows the storage to keep all members within
	// the same "root" storage for easier iteration.
	storage := &memberStorage{
		MemberStorage: commondao.NewMemberStorageWithGrouping(
			commondao.UseStorageFactory(innerFactory),
		),
		messages: broker,
	}

	// Subscribe to messages published by member groups
	storage.messages.Subscribe(msgMemberAdd, storage.handleMemberAddMsg)
	storage.messages.Subscribe(msgMemberRemove, storage.handleMemberRemoveMsg)
	return storage
}

type memberStorage struct {
	commondao.MemberStorage

	uniqueMembers avl.Tree // string(address) -> *avl.Tree(group -> struct{})
	messages      message.Subscriber
}

func (s *memberStorage) handleMemberAddMsg(msg message.Message) {
	data := msg.Data.(groupMemberUpdateData)
	key := data.Member.String()
	v, _ := s.uniqueMembers.Get(key)
	groups, ok := v.(*avl.Tree)
	if !ok {
		// Create a new tree to track member's groups
		groups = avl.NewTree()

		// Keep track of unique members
		s.uniqueMembers.Set(key, groups)
	}

	// Keep track of the new member group
	groups.Set(data.Group, struct{}{})
}

func (s *memberStorage) handleMemberRemoveMsg(msg message.Message) {
	data := msg.Data.(groupMemberUpdateData)
	key := data.Member.String()
	v, found := s.uniqueMembers.Get(key)
	if !found {
		return
	}

	// Remove the group from the list of groups member belongs
	groups := v.(*avl.Tree)
	groups.Remove(data.Group)

	// Remove the member from the unique storage when it doesn't belong
	// to any group and it doesn't exists in the base storage. Members
	// can be removed from groups but still exists as ungrouped.
	if groups.Size() == 0 && !s.MemberStorage.Has(data.Member) {
		s.uniqueMembers.Remove(data.Member.String())
	}
}

// Size returns the number of unique members in the storage.
// Members are only counted once, even if the same address is assigned to many groups.
func (s memberStorage) Size() int {
	return s.uniqueMembers.Size()
}

// TotalSize returns the number of unique members in the storage.
// Members are only counted once, even if the same address is assigned to many groups.
func (s memberStorage) TotalSize() int {
	return s.uniqueMembers.Size()
}

// Has checks if a member exists in the storage.
func (s memberStorage) Has(member address) bool {
	return s.uniqueMembers.Has(member.String())
}

// IterateByOffset iterates unique members starting at the given offset.
// The callback can return true to stop iteration.
func (s memberStorage) IterateByOffset(offset, count int, fn commondao.MemberIterFn) bool {
	return s.uniqueMembers.IterateByOffset(offset, count, func(key string, _ any) bool {
		return fn(address(key))
	})
}

// Add adds a member to the storage.
// Returns true if the member is added, or false if it already existed.
func (s *memberStorage) Add(member address) bool {
	added := s.MemberStorage.Add(member)
	if added && !s.uniqueMembers.Has(member.String()) {
		s.uniqueMembers.Set(member.String(), nil)
	}
	return added
}

// Remove removes a member from the storage.
// Returns true if member was removed, or false if it was not found.
func (s *memberStorage) Remove(member address) bool {
	removed := s.MemberStorage.Remove(member)
	if !removed {
		return false
	}

	// Check if the member being removed from the base storage is a member of a group
	v, _ := s.uniqueMembers.Get(member.String())
	groups, ok := v.(*avl.Tree)
	if ok && groups.Size() > 0 {
		return removed
	}

	// Remove unique member when it has no groups
	s.uniqueMembers.Remove(member.String())
	return removed
}

// groupMemberUpdateData defines a data type for group member updates.
type groupMemberUpdateData struct {
	Group  string
	Member address
}

// groupMemberStorage defines a member storage for member groups.
// This type of storage publishes messages when a member is added or removed from a group.
type groupMemberStorage struct {
	commondao.MemberStorage

	group    string
	messages message.Publisher
}

// Add adds a member to the storage.
// Returns true if the member is added, or false if it already existed.
// It panics when storage is sealed.
func (s *groupMemberStorage) Add(member address) bool {
	added := s.MemberStorage.Add(member)
	if added {
		s.messages.Publish(msgMemberAdd, groupMemberUpdateData{
			Group:  s.group,
			Member: member,
		})
	}
	return added
}

// Remove removes a member from the storage.
// Returns true if member was removed, or false if it was not found.
// It panics when storage is sealed.
func (s *groupMemberStorage) Remove(member address) bool {
	removed := s.MemberStorage.Remove(member)
	if removed {
		s.messages.Publish(msgMemberRemove, groupMemberUpdateData{
			Group:  s.group,
			Member: member,
		})
	}
	return removed
}
