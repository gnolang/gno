package commondao

// MemberGroupingOption configures member groupings.
type MemberGroupingOption func(MemberGrouping)

// SECURITY XXX: UseStorageFactory captures a caller-supplied factory
// that returns the MemberStorage used to track who is a DAO member.
// A malicious factory can return a MemberStorage impl that lies about
// membership — always-true (everyone is a member, votes pass trivially)
// or always-false (legitimate members are locked out). This isn't a
// cur-leak; it's a behavior-substitution risk on the data layer the
// DAO's voting and proposal logic depends on.
// Fix: document that the factory must come from trusted code (a
// hardcoded constructor in the DAO realm itself), and never accept
// a UseStorageFactory option from external-realm input. Or restrict
// to a canonical-impl allowlist (type-switch on known MemberStorage
// types) in commondao itself.
//
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
