// On-chain project management utilities.
// This library enables efficient project management for evaluation DAOs, governance modules, community pool management, launchpads, or general transparency.
package pm

// ITask is a completion unit that is associated with IOwner and can include sub-tasks.
// It is typically utilized for issues, epics, milestones, cards, and stories.
type ITask interface {
	// attributes

	ID() string
	Title() string
	Kind() stringOrEnum
	State() Enum
	Progress() int
	IsBlocked() bool

	// relationships

	Author() IOwner
	Parent() IOwner
	Assignees() []IOwner
	Reviewers() []IOwner
	Dependencies() ITask
}

// An IOwner contains ITask or other IOwner entities and can be applied to folders, repositories, teams, organizations, and users.
// It has no definitive completion status and can be archived but always serves as a container.
type IOwner interface {
	// attributes

	ID() string
	Title() string
	Kind() stringOrEnum

	// relationships

	Tasks() []ITask
}

// ITopic tags and labels ITask and IOwner entities for grouping related topics based on arbitrary information.
// These tags may be used with weighted attributes in the future.
type ITopic interface {
	ID() string
	Title() string
	Kind() stringOrEnum
}
