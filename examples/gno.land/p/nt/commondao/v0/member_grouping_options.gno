package commondao

// MemberGroupingOption configures member groupings.
type MemberGroupingOption func(MemberGrouping)

// UseStorageFactory assigns a custom member storage creation function to the grouping.
// Creation function is called each time a member group is added, with the name of the
// group as the only argument, to create a storage where the new group stores its members.
func UseStorageFactory(fn func(group string) MemberStorage) MemberGroupingOption {
	if fn == nil {
		panic("storage factory function must not be nil")
	}

	return func(g MemberGrouping) {
		grouping, ok := g.(*memberGrouping)
		if !ok {
			panic("storage factory not supported by member grouping")
		}

		grouping.createStorage = fn
	}
}

// WithGroups creates multiple members groups.
// To use a custom member storage factory to create the groups make sure that this option
// comes after the `UseStorageFactory()` option, otherwise groups are created using the
// default factory which is `commondao.NewMemberStorage()`.
func WithGroups(names ...string) MemberGroupingOption {
	return func(g MemberGrouping) {
		for _, name := range names {
			if _, err := g.Add(name); err != nil {
				panic(err)
			}
		}
	}
}
