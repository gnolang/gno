# basedao: Membership and Role Management for DAOs

`basedao` is a gnolang package that extends the DAOkit framework with comprehensive membership and role management capabilities. It serves as the foundation for most DAO implementation.

## 1. Core Components

Provides three main types for managing DAO membership and roles:

```go
// DAO member with assigned roles
type Member struct { 
    Address string    
    Roles   []string  
}

// Contains metadata about DAO roles
type RoleInfo struct { 
    Name        string 
    Description string 
    Color       string 
}

// Central component for managing members and roles
type MembersStore struct { 
    Roles   *avl.Tree 
    Members *avl.Tree 
}
```

### MembersStore Usage

**Creating and initializing**:
```go
// Create with initial data
store := basedao.NewMembersStore(
    []basedao.RoleInfo{
        {Name: "admin", Description: "Administrators", Color: "#329175"},
        {Name: "treasurer", Description: "Treasury management", Color: "#F3D3BC"},
    },
    []basedao.Member{
        {Address: "g1admin...", Roles: []string{"admin"}},
        {Address: "g1user...", Roles: []string{}},
    },
)

// Create empty store for dynamic management
store := basedao.NewMembersStore(nil, nil)
```

**Common operations**:
```go
// Member operations
store.AddMember("g1new...", []string{"moderator"})
store.RemoveMember("g1former...")
isMember := store.IsMember("g1user...")
count := store.MembersCount()

// Role operations  
store.AddRole(basedao.RoleInfo{Name: "secretary", Description: "Records keeper", Color: "#4A90E2"})
store.RemoveRole("obsolete-role")
store.AddRoleToMember("g1user...", "secretary")
store.RemoveRoleFromMember("g1user...", "moderator")

// Query operations
hasRole := store.HasRole("g1user...", "admin")
memberRoles := store.GetMemberRoles("g1user...")
roleMembers := store.GetMembersWithRole("admin")
adminCount := store.CountMembersWithRole("admin")
membersWithoutRoles := store.GetMembersWithoutRole()
roleInfo := store.RoleInfo("admin")  // Get role metadata
membersJSON := store.GetMembersJSON() // Export as JSON
```

## 2. Built-in Actions

Provides built-in actions for common DAO operations. Each action has a unique type identifier:

### Action Type Constants
```go
const ActionAddMemberKind = "gno.land/p/samcrew/basedao.AddMember"
const ActionRemoveMemberKind = "gno.land/p/samcrew/basedao.RemoveMember"  
const ActionAssignRoleKind = "gno.land/p/samcrew/basedao.AssignRole"
const ActionUnassignRoleKind = "gno.land/p/samcrew/basedao.UnassignRole"
const ActionEditProfileKind = "gno.land/p/samcrew/basedao.EditProfile"
const ActionChangeDAOImplementationKind = "gno.land/p/samcrew/basedao.ChangeDAOImplementation"
```

### Creating Actions
```go
// Add a member with roles
action := basedao.NewAddMemberAction(&basedao.ActionAddMember{
    Address: address("g1newmember..."),
    Roles:   []string{"moderator", "treasurer"},
})

// Remove member
action := basedao.NewRemoveMemberAction(address("g1member..."))

// Assign role to member
action := basedao.NewAssignRoleAction(&basedao.ActionAssignRole{
    Address: address("g1member..."),
    Role:    "admin",
})

// Remove role from member
action := basedao.NewUnassignRoleAction(&basedao.ActionUnassignRole{
    Address: address("g1member..."),
    Role:    "moderator",
})

// Edit DAO profile
action := basedao.NewEditProfileAction(
    [2]string{"DisplayName", "My Updated DAO Name"},
    [2]string{"Bio", "An improved description of our DAO"},
    [2]string{"Avatar", "https://example.com/new-logo.png"},
)
```

## 3. Core Governance Interface

Allow DAO's members to interact through proposals and voting.

### Creating Proposals
```go
// Create a new proposal for members to vote on
// Returns its proposal ID
func Propose(req daokit.ProposalRequest) uint64 {...}

type ProposalRequest struct {
	Title       string 
	Description string 
	Action      Action
}
```

**Example - Adding a new member:**
```go
addMemberAction := basedao.NewAddMemberAction(&basedao.ActionAddMember{
    Address: "g1alice...",
    Roles:   []string{"treasurer"},
})

proposal := daokit.ProposalRequest{
    Title:       "Add new treasurer",
    Description: "Proposal to add Alice as treasurer for better fund management",
    Action:      addMemberAction,
}
proposalID := Propose(proposal)
```

#### Voting on Proposals
```go
// Cast your vote on an active proposal
// Available vote: VoteYes, VoteNo, VoteAbstain
func Vote(proposalID uint64, vote daocond.Vote) {...}

Vote(1, daocond.VoteYes) // Vote yes on proposal #1
Vote(2, daocond.VoteNo) // Vote no on proposal #2
Vote(3, daocond.VoteAbstain) // Vote abstain on proposal #3
```

#### Executing Proposals  
```go
// Execute a proposal that has passed its voting requirements
func Execute(proposalID uint64) {...}

Execute(1) // Execute proposal #1 -- only works if it has enough votes
```

#### Instant Execution

Skip the voting process and execute a proposal immediately if you have the required permissions:

```go
// This performs: Propose() -> Vote(VoteYes) -> Execute()
proposalID := daokit.InstantExecute(DAO, proposal)
```

Useful for admin actions, migrations, and emergency procedures.

### Rendering DAO Information

**Built-in render paths**:
- `/` - Main DAO overview with basic info
- `/members` - Member list with roles and permissions
- `/proposals` - Proposal list with their status and voting progress  
- `/proposals/{id}` - Detailed view of a proposal with vote breakdown
- `/config` - DAO configuration and governance rules
- `/roles` - Available roles and their descriptions

You can add or overwrite renders by providing a custom `RenderFn` in the DAO configuration.

## 4. Creating a DAO with basedao

### 4.1 Basic DAO Setup

```go
package my_dao

import (
    "gno.land/p/samcrew/basedao"
    "gno.land/p/samcrew/daocond"
    "gno.land/p/samcrew/daokit"
)

var (
	DAO        daokit.DAO          // External interface for DAO interaction
	daoPrivate *basedao.DAOPrivate // Full access to internal DAO state
)

func init() {
    // Set up roles
    roles := []basedao.RoleInfo{
        {Name: "admin", Description: "Administrators", Color: "#329175"},
        {Name: "member", Description: "Regular members", Color: "#21577A"},
    }

    // Add initial members
    members := []basedao.Member{
        {Address: "g1admin...", Roles: []string{"admin"}},
        {Address: "g1user1...", Roles: []string{"member"}},
        {Address: "g1user2...", Roles: []string{"member"}},
    }

    store := basedao.NewMembersStore(roles, members)

    // Require 60% of members to approve proposals
    condition := daocond.MembersThreshold(0.6, store.IsMember, store.MembersCount)

    // Create the DAO
    DAO, daoPrivate = basedao.New(&basedao.Config{
        Name:             "My DAO",
        Description:      "A simple DAO example",
        Members:          store,
        InitialCondition: condition,
    })
}

// Create a new Proposal to be voted on
// To execute this function, you must use a MsgRun (maketx run)
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
func Propose(req daokit.ProposalRequest) {
	DAO.Propose(req)
}

// Allows DAO members to cast their vote on a specific proposal
func Vote(proposalID uint64, vote daocond.Vote) {
    DAO.Vote(proposalID, vote)
}

// Triggers the implementation of a proposal's actions
func Execute(proposalID uint64) {
	DAO.Execute(proposalID)
}

// Render generates a UI representation of the DAO's state
func Render(path string) string {
	return DAO.Render(path)
}
```

### 4.2 Configuration Options

```go
type Config struct {
	// Basic DAO information
	Name        string
	Description string
	ImageURI    string

	// Core components
	Members *MembersStore

	// Feature toggles
	NoDefaultHandlers  bool // Skips registration of default management actions (add/remove members, etc.)
	NoDefaultRendering bool // Skips setup of default web UI rendering routes
	NoCreationEvent    bool // Skips emitting the DAO creation event

	// Governance configuration
	InitialCondition daocond.Condition // Default condition for all built-in actions, defaults to 60% member majority

	// Profile integration (optional)
	SetProfileString ProfileStringSetter // Function to update profile fields (DisplayName, Bio, Avatar)
	GetProfileString ProfileStringGetter // Function to retrieve profile fields for members

	// Advanced customization hooks
	SetImplemFn       SetImplemRaw      // Function called when DAO implementation changes via governance
	MigrationParamsFn MigrationParamsFn // Function providing parameters for DAO upgrades
	RenderFn          RenderFn          // Rendering function for Gnoweb
	CrossFn           daokit.CrossFn    // Cross-realm communication function for multi-realm DAOs
	CallerID          CallerIDFn        // Custom function to identify the current caller, defaults to realmid.Previous

	// Internal configuration
	PrivateVarName string // Name of the private DAO variable for member querying extensions
}
```

## 5. DAO Upgrades and Migration

Supports upgrading DAO implementations through governance proposals, allowing DAOs to evolve over time.

### 5.1 Configuration for Upgrades

```go
DAO, daoPrivate = basedao.New(&basedao.Config{
    // ... other config
    MigrationParamsFn: func() []any { return nil }, // Parameters passed to migration function

    SetImplemFn:      setImplem,           
    CrossFn:          crossFn,             
})

// Update DAO variables after migration
func setImplem(newLocalDAO daokit.DAO, newDAO daokit.DAO) {
    localDAO, DAO = newLocalDAO, newDAO
}

// Necessary due to crossing constraint
func crossFn(_ realm, callback func()) {
	callback()
}
```

### 5.2 Migration Process

```go
// Migration function signature
type MigrateFn = func(prev *DAOPrivate, params []any) daokit.DAO

// Parameters function signature  
type MigrationParamsFn = func() []any

// 1. Define migration function
// params contains data from MigrationParamsFn - use for config, settings, etc.
func migrateTo2_0(prev *basedao.DAOPrivate, params []any) daokit.DAO {
    // Preserve existing member store
    memberStore := prev.Members
    
    // Add new roles for v2.0
    memberStore.AddRole(basedao.RoleInfo{
        Name: "auditor", 
        Description: "Financial oversight",
    })
    
    // Create new DAO with enhanced features
    newLocalDAO, newPrivate := basedao.New(&basedao.Config{
        Name:             prev.InitialConfig.Name + " v2.0",
        Description:      "Upgraded DAO with audit capabilities",
        Members:          memberStore,
        InitialCondition: prev.InitialConfig.InitialCondition,
        // ... other configuration
    })
    
    return newLocalDAO
}

// 2. Create and submit upgrade proposal
action := basedao.NewChangeDAOImplementationAction(migrateTo2_0)
proposal := daokit.ProposalRequest{
    Title:       "Upgrade to DAO v2.0",
    Description: "Adds auditor role and enhanced governance",
    Action:      action,
}
proposalID := DAO.Propose(proposal)

// 3. Execute Migration
DAO.Execute(proposalID)

// Alternatively, you can use InstantExecute to skip the voting process
// if you have sufficient permissions to execute the action directly
daokit.InstantExecute(DAO, proposal) 
```

## 6. Event System

Events are emitted when actions happen in your DAO. This helps track activities.

```go
// DAO creation (only if NoCreationEvent is false)
chain.Emit("BaseDAOCreated")

// Member management
chain.Emit("BaseDAOAddMember", "address", memberAddr)
chain.Emit("BaseDAORemoveMember", "address", memberAddr)
```

## 7. Membership Extension

Allows other packages and realms to check if an address is a member of your DAO.

### Usage

```go
import "gno.land/p/samcrew/basedao"

// Check if someone is a DAO member
ext := basedao.MustGetMembersViewExtension(dao)
if ext.IsMember("g1user...") {
    // User is a member
}
```

### Example: Member-Only action

```go
package my_content

import (
    "gno.land/p/samcrew/basedao"
    "gno.land/r/some/dao"
)

func Post(title, content string) {
    caller := std.PrevRealm().Addr()
    ext := basedao.MustGetMembersViewExtension(dao.DAO)
    
    if !ext.IsMember(caller.String()) {
        panic("Only DAO members can post")
    }
    
    createPost(title, content)
}
```

The extension is automatically registered when you create a DAO with `basedao.New()`.

---

*Part of the daokit framework for building decentralized autonomous organizations in gnolang.*