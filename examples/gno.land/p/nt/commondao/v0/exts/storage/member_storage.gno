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

	v, found := storage.memberGroups.Get(member.String())
	if !found {
		return nil
	}

	tree := v.(*avl.Tree)
	groups := make([]string, 0, tree.Size())
	tree.Iterate("", "", func(group string, _ any) bool {
		groups = append(groups, group)
		return false
	})
	return groups
}

// NewMemberStorage creates a new CommonDAO member storage with grouping support.
//
// This is a custom storage that automatically adds or removes members that are added
// or removed from any of the member groups. This allows for quick and inexpensive
// checks for the number of total unique storage users, including users added to groups,
// and also to iterate all of them without needing to iterate individual groups.
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

	memberGroups avl.Tree // string(address) -> *avl.Tree(group -> struct{})
	messages     message.Subscriber
}

func (s *memberStorage) handleMemberAddMsg(msg message.Message) {
	data := msg.Data.(groupMemberUpdateData)
	key := data.Member.String()
	v, _ := s.memberGroups.Get(key)
	groups, ok := v.(*avl.Tree)
	if !ok {
		// Create a new tree to track member's groups
		groups = avl.NewTree()
		s.memberGroups.Set(key, groups)

		// Add the new member to the storage
		s.MemberStorage.Add(data.Member)
	}

	// Keep track of the new member group
	groups.Set(data.Group, struct{}{})
}

func (s *memberStorage) handleMemberRemoveMsg(msg message.Message) {
	data := msg.Data.(groupMemberUpdateData)
	key := data.Member.String()
	v, found := s.memberGroups.Get(key)
	if !found {
		// Member should always be found
		return
	}

	// Remove the group from the list of groups member belongs
	groups := v.(*avl.Tree)
	groups.Remove(data.Group)

	// Remove the member from the storage when it doesn't belong to any group
	if groups.Size() == 0 {
		s.memberGroups.Remove(key)
		s.MemberStorage.Remove(data.Member)
	}
}

// Size returns the number of members in the storage.
// It also includes unique members that belong to any number of member groups.
func (s memberStorage) Size() int {
	return s.MemberStorage.Size()
}

// IterateByOffset iterates members starting at the given offset.
// The callback can return true to stop iteration.
// It also iterates unique members that belong to any of the member groups.
func (s memberStorage) IterateByOffset(offset, count int, fn commondao.MemberIterFn) bool {
	return s.MemberStorage.IterateByOffset(offset, count, fn)
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
func (s *groupMemberStorage) Add(member address) bool {
	s.messages.Publish(msgMemberAdd, groupMemberUpdateData{
		Group:  s.group,
		Member: member,
	})
	return s.MemberStorage.Add(member)
}

// Remove removes a member from the storage.
// Returns true if member was removed, or false if it was not found.
func (s *groupMemberStorage) Remove(member address) bool {
	s.messages.Publish(msgMemberRemove, groupMemberUpdateData{
		Group:  s.group,
		Member: member,
	})
	return s.MemberStorage.Remove(member)
}
