# daokit

# 1. Introduction

A **Decentralized Autonomous Organization (DAO)** is a self-governing entity that operates through smart contracts, enabling transparent decision-making without centralized control.

`daokit` is a gnolang package for creating complex DAO models. It introduces a new framework based on conditions, composed of :
- `daokit` : Core package for building DAOs, proposals, and actions
- `basedao` : Extension with membership and role management
- `daocond`: Stateless condition engine for evaluating proposals

# 2. What is `daokit` ?

`daokit` provides a powerful condition and role-based system to build flexible and programmable DAOs.

## Key Features:
- Create proposals that include complex execution logic
- Attach rules (conditions) to each resource
- Assign roles to users to structure permissions and governance

## 2.1 Key Concepts

- **Proposal**: A request to execute a **resource**. Proposals are voted on and executed only if predefined **conditions** are met.
- **Resource**: An executable action within the DAO. Each resource is governed by a **condition**.
- **Condition**: A set of rules that determine whether a proposal can be executed.
- **Role**: Labels that assign governance power or permissions to DAO members.

**Example Use Case**: A DAO wants to create a proposal to spend money from its treasury.

**Rules**:
- `SpendMoney` is a resource with a condition requiring:
	- 50% approval from the administration board
	- Approval from the CFO

**Outcome**:
- Any user can propose to spend money
- Only board and CFO votes are considered
- The proposal executes only if the condition is satisfied

# 3. Architecture

DAOkit framework is composed of three packages:

## 3.1 [daocond](/p/samcrew/daocond)

`daocond` provides a stateless condition engine used to evaluate if a proposal should be executed.

### 3.1.1 Interface
```go
type Condition interface {
	// Eval checks if the condition is satisfied based on current votes.
	Eval(ballot Ballot) bool
	// Signal returns a value from 0.0 to 1.0 to indicate how close the condition is to being met.
	Signal(ballot Ballot) float64

	// Render returns a static human-readable representation of the condition.
	Render() string
	// RenderWithVotes returns a dynamic representation with vote context included.
	RenderWithVotes(ballot Ballot) string
}

type Ballot interface {
	// Vote allows a user to vote on a proposal.
	Vote(voter string, vote Vote)
	// Get returns the vote of a user.
	Get(voter string) Vote
	// Total returns the total number of votes.
	Total() int
	// Iterate iterates over all votes, similar as avl.Tree.Iterate.
	Iterate(fn func(voter string, vote Vote) bool)
}
```

### 3.1.2 Built-in Conditions
`daocond` provides several built-in conditions to cover common governance scenarios.

```go
// MembersThreshold requires that a specified fraction of all DAO members approve the proposal.
func MembersThreshold(threshold float64, isMemberFn func(memberId string) bool, membersCountFn func() uint64) Condition

// RoleThreshold requires that a certain percentage of members holding a specific role approve.
func RoleThreshold(threshold float64, role string, hasRoleFn func(memberId string, role string) bool, usersRoleCountFn func(role string) uint32) Condition

// RoleCount requires a fixed minimum number of members holding a specific role to approve.
func RoleCount(count uint64, role string, hasRoleFn func(memberId string, role string) bool) Condition
```

### 3.1.3 Logical Composition
You can combine multiple conditions to create complex governance rules using logical operators:

```go
// And returns a condition that is satisfied only if *all* provided conditions are met.
func And(conditions ...Condition) Condition
// Or returns a condition that is satisfied if *any* one of the provided conditions is met.
func Or(conditions ...Condition) Condition
```

**Example**:
```go
// Require both admin approval and at least one CFO
cond := daocond.And(
    daocond.RoleThreshold(0.5, "admin", hasRole, roleCount),
    daocond.RoleCount(1, "CFO", hasRole),
)
```

Conditions are stateless for flexibility and scalability.

## 3.2 daokit

`daokit` provides the core mechanics:

### 3.2.1 Core Structure:
It's the central component of a DAO, responsible for managing both available resources that can be executed and the proposals.
```go
type Core struct {
	Resources *ResourcesStore
	Proposals *ProposalsStore
}
```

### 3.2.2 DAO Interface:
The interface defines the external functions that users or other modules interact with. It abstracts the core governance flow: proposing, voting, and executing.
```go
type DAO interface {
	Propose(req ProposalRequest) uint64
	Vote(id uint64, vote daocond.Vote)
	Execute(id uint64)
}
```
> 📖 [Code Example of a Basic DAO](#4-code-example-of-a-basic-dao)

### 3.2.3 Proposal Lifecycle

Each proposal goes through the following states:

1. **Open**: 
- Initial state after proposal creation.
- Accepts votes from eligible participants.

2. **Passed**
- Proposal has gathered enough valid votes to meet the condition.
- Voting is **closed** and cannot be modified.
- The proposal is now eligible for **execution**.

3. **Executed**
- Proposal action has been successfully carried out.
- Final state — proposal can no longer be voted on or modified.


## 3.3 [basedao](/p/samcrew/basedao)

`basedao` extends `daokit` to handle members and roles management.
It handles who can participate in a DAO and what permissions they have.

### 3.3.1 Core Types
```go
type MembersStore struct {
	Roles   *avl.Tree 
	Members *avl.Tree 
}
```

### 3.3.2 Initialize the DAO
Create a `MembersStore` structure to initialize the DAO with predefined roles and members.

```go
roles := []basedao.RoleInfo{
	{Name: "admin", Description: "Administrators"},
	{Name: "finance", Description: "Handles treasury"},
}

members := []basedao.Member{
	{Address: "g1abc...", Roles: []string{"admin"}},
	{Address: "g1xyz...", Roles: []string{"finance"}},
}

store := basedao.NewMembersStore(roles, members)
```

### 3.3.3 Example Usage
```go
store := basedao.NewMembersStore(nil, nil)

// Add a role and assign it
store.AddRole(basedao.RoleInfo{Name: "moderator", Description: "Can moderate posts"})
store.AddMember("g1alice...", []string{"moderator"})

// Update role assignment
store.AddRoleToMember("g1alice...", "editor")
store.RemoveRoleFromMember("g1alice...", "moderator")

// Inspect the state
isMember := store.IsMember("g1alice...") // "Is Alice a member?"
hasRole := store.HasRole("g1alice...", "editor") // "Is Alice an editor?"
members := store.GetMembersJSON() // "All Members (JSON):"
```

### 3.3.4 Creating a DAO:

```go
func New(conf *Config) (daokit.DAO, *DAOPrivate)
```

#### 3.3.4.1 Key Structures:
- `DAOPrivate`: Full access to internal DAO state
- `daokit.DAO`: External interface for DAO interaction


### 3.3.5 Configuration:
```go
type Config struct {
	Name              string
	Description       string
	ImageURI          string
	// Use `basedao.NewMembersStore(...)` to create members and roles.
	Members           *MembersStore
	// Set to `true` to disable built-in actions like add/remove member.
	NoDefaultHandlers bool
	// Default rule applied to all built-in DAO actions.
	InitialCondition  daocond.Condition
	// Optional helpers to store profile data (e.g., from `/r/demo/profile`).
	SetProfileString  ProfileStringSetter
	GetProfileString  ProfileStringGetter
	// Set to `true` if you don’t want a "DAO Created" event to be emitted.
	NoCreationEvent   bool
}
```

# 4. Code Example of a Basic DAO

```go
package daokit_demo

import (
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
	"gno.land/r/demo/profile"
)

var (
	DAO        daokit.DAO // External interface for DAO interaction
	daoPrivate *basedao.DAOPrivate // Full access to internal DAO state
)

func init() {
	initialRoles := []basedao.RoleInfo{
		{Name: "admin", Description: "Admin is the superuser"},
		{Name: "public-relationships", Description: "Responsible of communication with the public"},
		{Name: "finance-officer", Description: "Responsible of funds management"},
	}

	initialMembers := []basedao.Member{
		{Address: "g126...zlg", Roles: []string{"admin", "public-relationships"}},
		{Address: "g1ld6...3jv", Roles: []string{"public-relationships"}},
		{Address: "g1r69...0tth", Roles: []string{"finance-officer"}},
		{Address: "g16jv...6e0r", Roles: []string{}},
	}

	memberStore := basedao.NewMembersStore(initialRoles, initialMembers)

	membersMajority := daocond.MembersThreshold(0.6, memberStore.IsMember, memberStore.MembersCount)
	publicRelationships := daocond.RoleCount(1, "public-relationships", memberStore.HasRole)
	financeOfficer := daocond.RoleCount(1, "finance-officer", memberStore.HasRole)

	// `and` and `or` use va_args so you can pass as many conditions as needed
	adminCond := daocond.And(membersMajority, publicRelationships, financeOfficer)

	DAO, daoPrivate = basedao.New(&basedao.Config{
		Name:             "Demo DAOKIT DAO",
		Description:      "This is a demo DAO built with DAOKIT",
		Members:          memberStore,
		InitialCondition: adminCond,
		GetProfileString: profile.GetStringField,
		SetProfileString: profile.SetStringField,
	})
}

func Vote(proposalID uint64, vote daocond.Vote) {
	DAO.Vote(proposalID, vote)
}

func Execute(proposalID uint64) {
	DAO.Execute(proposalID)
}

func Render(path string) string {
	return daoPrivate.Render(path)
}
```

# 5. Create Custom Resources

To add new behavior to your DAO — or to enable others to integrate your package into their own DAOs — define custom resources by implementing:

```go
type Action interface {
	Type() string // return the type of the action. e.g.: "gno.land/p/samcrew/blog.NewPost"
	String() string // return stringify content of the action
}

type ActionHandler interface {
	Type() string // return the type of the action. e.g.: "gno.land/p/samcrew/blog"
	Execute(action Action) // executes logic associated with the action
}
```
This allows DAOs to execute arbitrary logic or interact with Gno packages through governance-approved decisions.

## Steps to Add a Custom Resource:
1. Define the path of the action, it should be unique 
```go
// XXX: pkg "/p/samcrew/blog" - does not exist, it's just an example
const ActionNewPostKind = "gno.land/p/samcrew/blog.NewPost"
```

2. Create the structure type of the payload
```go
type ActionNewPost struct {
	Title string
	Content string
}
```

3. Implement the action and handler
```go
func NewPostAction(title, content string) daokit.Action {
	// def: daoKit.NewAction(kind: String, payload: interface{})
	return daokit.NewAction(ActionNewPostKind, &ActionNewPost{
		Title:   title,
		Content: content,
	})
}

func NewPostHandler(blog *Blog) daokit.ActionHandler {
	// def: daoKit.NewActionHandler(kind: String, payload: func(interface{}))
	return daokit.NewActionHandler(ActionNewPostKind, func(payload interface{}) {
		action, ok := payload.(*ActionNewPost)
		if !ok {
			panic(errors.New("invalid action type"))
		}
		blog.NewPost(action.Title, action.Content)
	})
}
```

4. Register the resource
```go
resource := daokit.Resource{
    Condition: daocond.NewRoleCount(1, "CEO", daoPrivate.Members.HasRole),
    Handler: blog.NewPostHandler(blog),
}
daoPrivate.Core.Resources.Set(&resource)
```
