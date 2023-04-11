// project-management library.
package pm

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
	Parent() ITask
	Assignees() []IOwner
	Reviewers() []IOwner
	Dependencies() ITask
}

type IOwner interface {
	// attributes

	ID() string
	Title() string
	Kind() stringOrEnum

	// relationships

	Tasks() []ITask
}

type ITopic interface {
	ID() string
	Title() string
	Kind() stringOrEnum
}
