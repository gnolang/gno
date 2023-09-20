# DAO Framework Quick Start Guide

## Overview

In this guide, we'll go over how to set up and configure a Decentralized Autonomous Organization (DAO) using the Gno DAO framework. We'll cover the core components that make up a DAO, walk you through the process of creating your first DAO, and provide code examples to help you get started.

## Table of Contents

1. [Theoretical Foundations](#theoretical-foundations)
    - [Core Interface](#core-interface)
    - [Voting Module](#voting-module)
    - [Proposal Modules](#proposal-modules)
    - [Message Handlers](#message-handlers)
2. [Practical Implementation](#practical-implementation)
    - [Setting Up Your Workspace](#setting-up-your-workspace)
    - [Creating the Voting Module](#creating-the-voting-module)
    - [Creating the Proposal Module](#creating-the-proposal-module)
    - [Registering Message Handlers](#registering-message-handlers)
    - [Creating the DAO Core](#creating-the-dao-core)

---

## Theoretical Foundations

### Core Interface

The `IDAOCore` interface ties all the other components together and offers the main entry points for interacting with the DAO.

**Interface Definition:**

```go
type IDAOCore interface {
	Render(path string) string

	VotingModule() IVotingModule
	ProposalModules() []ActivableProposalModule
	ActiveProposalModuleCount() int
	Registry() *MessagesRegistry

	UpdateVotingModule(newVotingModule IVotingModule)
	UpdateProposalModules(toAdd []IProposalModule, toDisable []int)
}
```

A default implementation is provided in the package `gno.land/p/demo/teritori/dao_core`, and custom implementations are generally not required.

### A voting module

### Voting Module

The `gno.land/p/demo/teritori/dao_interfaces.IVotingModule` interface defines how voting power is allocated to addresses within the DAO.

**Interface Definition:**

```go
type IVotingModule interface {
	Info() ModuleInfo
	ConfigJSON() string
	Render(path string) string
	VotingPowerAtHeight(address std.Address, height int64) (power uint64)
	TotalPowerAtHeight(height int64) uint64
}
```

There is only one implementation currently, `gno.land/p/demo/teritori/dao_voting_group`, it's a wrapper around a gno group, providing a membership-based voting power definition

### Proposal Modules

A proposal module (`gno.land/p/demo/teritori/dao_interfaces.IProposalModule`) is responsible for:
- Receiving proposals, the proposal type is defined by the module
- Managing the proposals lifecycle
- Tallying votes, the vote type is defined by the module and the associated voting power is queried from the voting module
- Executing proposals once they are passed

**Interface Definition:**
```go
type IProposalModule interface {
	Core() IDAOCore
	Info() ModuleInfo
	ConfigJSON() string
	Render(path string) string
	Execute(proposalID int)
	VoteJSON(proposalID int, voteJSON string)
	ProposeJSON(proposalJSON string) int
	ProposalsJSON(limit int, startAfter string, reverse bool) string
	ProposalJSON(proposalID int) string
}
```

There is only one implementation currently, `gno.land/p/demo/teritori/dao_proposal_single`, providing a yes/no/abstain vote model with quorum and threshold

### Message handlers

Proposals actions are encoded as objects implementing `gno.land/p/demo/teritori/dao_interfaces.ExecutableMessage`
```go
type ExecutableMessage interface {
	ujson.JSONAble
	ujson.FromJSONAble

	String() string
	Type() string
}
```

They are deserialized and executed by message handlers implementing `gno.land/p/demo/teritori/dao_interfaces.MessageHandler`
```go
type MessageHandler interface {
	Execute(message ExecutableMessage)
	MessageFromJSON(ast *ujson.JSONASTNode) ExecutableMessage
	Type() string
}
```

Message handlers are registered at core creation and new message handlers can be registered via proposals to extend the DAO capabilities

## Practical Implementation

### Setting Up Your Workspace

Sooo, let's create a new realm

```
git clone https://github.com/TERITORI/gno.git gno-dao-tutorial
cd gno-dao-tutorial
git checkout teritori-unified
mkdir examples/gno.land/r/demo/my_dao
```

### Creating the Voting Module

We will start by instantiating a voting module

1. **Initialize the Factory**

Modules instantiation uses the factory pattern in case the module needs to access the core

`examples/gno.land/r/demo/my_dao/my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/teritori/dao_interfaces"
)

func init() {
    votingModuleFactory := func(core dao_interfaces.IDAOCore) {
   
    }
}
```

2. **Create the Group**

First we need to create a group that will be queried by the module. We use the teritori fork of the groups realm since the upstream groups can't be created by a non-EOA

`examples/gno.land/r/demo/my_dao/my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/teritori/dao_interfaces"
    "gno.land/p/demo/teritori/groups" // <- new
)

func init() {
    var groupID groups.GroupID

    votingModuleFactory := func(core dao_interfaces.IDAOCore) {
        groupID = groups.CreateGroup("my_dao_voting_group") // <- new
    }
}
```

We need to keep a reference to the group ID to instantiate it's message handlers later

3. **Add Initial Members**

`examples/gno.land/r/demo/my_dao/my_dao.gno`
```go
func init() {
    votingModuleFactory := func(core dao_interfaces.IDAOCore) {
        groupID = groups.CreateGroup("my_dao_voting_group")
        groups.AddMember(groupID, "your-address", 1, "") // <- new
        // repeat for any other initial members you want in the DAO
    }
}
```

4. **Instantiate the voting module**

`examples/gno.land/r/demo/my_dao/my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/teritori/dao_interfaces"
    "gno.land/p/demo/teritori/groups"
    "gno.land/p/demo/teritori/dao_voting_group" // <- new
)

func init() {
    votingModuleFactory := func(core dao_interfaces.IDAOCore) dao_interfaces.IVotingModule {
        groupID = groups.CreateGroup("my_dao_voting_group")
        groups.AddMember(groupID, "your-address", 1, "")
        return dao_voting_group.NewVotingGroup(groupID) // <- new
    }
}
```

Now let's create a proposal module

### Creating the proposal module

1. **Initialize the Factory**

`examples/gno.land/r/demo/my_dao/my_dao.gno`
```go
func init() {
    votingModuleFactory := func(core dao_interfaces.IDAOCore) {
        // ...
    }
    proposalModuleFactories := []dao_interfaces.ProposalModuleFactory{
        func(core dao_interfaces.IDAOCore) dao_interfaces.IProposalModule {

        },
    }
}
```

2. **Configure and instantiate the Proposal Module**

`examples/gno.land/r/demo/my_dao/my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/teritori/dao_interfaces"
    "gno.land/p/demo/teritori/groups"
    "gno.land/p/demo/teritori/dao_voting_group"
    "gno.land/p/demo/teritori/dao_proposal_single" // <- new
)

func init() {
    // ...
    var proposalModule *dao_proposal_single.DAOProposalSingle
    proposalModuleFactories := []dao_interfaces.ProposalModuleFactory{
        func(core dao_interfaces.IDAOCore) dao_interfaces.IProposalModule {
            tt := dao_proposal_single.PercentageThresholdPercent(100) // 1% threshold
			tq := dao_proposal_single.PercentageThresholdPercent(100) // 1% quorum
			proposalModule = dao_proposal_single.NewDAOProposalSingle(core, &dao_proposal_single.DAOProposalSingleOpts{
				MaxVotingPeriod: time.Hour * 24 * 42,
				Threshold: &dao_proposal_single.ThresholdThresholdQuorum{
					Threshold: &tt,
					Quorum:    &tq,
				},
			})
            return proposalModule
        },
    }
}
```

We need to keep a reference to the group ID to instantiate it's message handlers later

### Registering Message Handlers

Add message handlers to allow your DAO to perform specific actions when proposals are executed.

`examples/gno.land/r/demo/my_dao/my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/teritori/dao_interfaces"
    "gno.land/p/demo/teritori/groups"
    "gno.land/p/demo/teritori/dao_voting_group"
    "gno.land/p/demo/teritori/dao_proposal_single"
)

func init() {
    // ...
    messageHandlersFactories := []dao_interfaces.MessageHandlerFactory{
        // Allow to manage the voting group
        func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
            return groups.NewAddMemberHandler(groupID)
        },
        func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
            return groups.NewDeleteMemberHandler(groupID)
        },
        // Allow to update the proposal module settings
        func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
            return dao_proposal_single.NewUpdateSettingsHandler(proposalModule)
        },
    }
}
```

### Creating the DAO Core

Now we can create the actual DAO

```go
package my_dao

import (
    "gno.land/p/demo/teritori/dao_interfaces"
    "gno.land/p/demo/teritori/groups"
    "gno.land/p/demo/teritori/dao_voting_group"
    "gno.land/p/demo/teritori/dao_proposal_single"
    "gno.land/p/demo/teritori/dao_core" // <- new
)

var (
    daoCore dao_interfaces.IDAOCore // <- new
)

func init() {
    // ...
    messageHandlersFactories := []dao_interfaces.MessageHandlerFactory{
        // ...
    }
    daoCore = dao_core.NewDAOCore(votingModuleFactory, proposalModuleFactories, messageHandlersFactories) // <- new
}
```

We also need to expose the DAO methods in the realm

```go
func init() {
    // ...
}

func Render(path string) string {
	return daoCore.Render(path)
}

func VoteJSON(moduleIndex int, proposalID int, voteJSON string) {
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	if !module.Enabled {
		panic("proposal module is not enabled")
	}
	module.Module.VoteJSON(proposalID, voteJSON)
}

func Execute(moduleIndex int, proposalID int) {
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	if !module.Enabled {
		panic("proposal module is not enabled")
	}
	module.Module.Execute(proposalID)
}

func ProposeJSON(moduleIndex int, proposalJSON string) int {
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	if !module.Enabled {
		panic("proposal module is not enabled")
	}
	return module.Module.ProposeJSON(proposalJSON)
}

func getProposalsJSON(moduleIndex int, limit int, startAfter string, reverse bool) string {
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	return module.Module.ProposalsJSON(limit, startAfter, reverse)
}

func getProposalJSON(moduleIndex int, proposalIndex int) string {
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	return module.Module.ProposalJSON(proposalIndex)
}
```

## Deploying and Interacting with Your DAO

*TODO: Add instructions for deploying and interacting with the DAO.*

---

That's it! You've successfully created your first DAO using the Gno DAO framework. To expand its capabilities, you can register additional message handlers or even create new modules if you feel bold