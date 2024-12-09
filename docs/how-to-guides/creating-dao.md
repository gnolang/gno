---
id: creating-dao
---

# How to Create a DAO

## Overview

This guide will show you how to write a simple [**DAO**](https://en.wikipedia.org/wiki/Decentralized_autonomous_organization) [realm](../concepts/realms.md) in [Gno](../concepts/gno-language.md). For actually deploying the realm, please see the
[deployment](deploy.md) guide.

We'll cover the core components that make up a DAO, walk you through the process of creating your first DAO, and provide code examples to help you get started.

## Theoretical Foundations

### Core Interface

The `IDAOCore` interface ties all the other components together and offers the main entry points for interacting with the DAO.

**Interface Definition:**

```go
type ActivableProposalModule struct {
	Enabled bool
	Module  IProposalModule
}

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

A default implementation is provided in the package `gno.land/p/demo/dao_maker/dao_core`, and custom implementations are generally not required.

### Voting Module

The `gno.land/p/demo/dao_maker/dao_interfaces.IVotingModule` interface defines how voting power is allocated to addresses within the DAO.

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

There is only one implementation currently, `gno.land/p/demo/dao_maker/dao_voting_group`, providing a membership-based voting power definition.

### Proposal Modules

A proposal module (`gno.land/p/demo/dao_maker/dao_interfaces.IProposalModule`) is responsible for:
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

There is only one implementation currently, `gno.land/p/demo/dao_maker/dao_proposal_single`, providing a yes/no/abstain vote model with quorum and threshold.

### Message handlers

Proposal actions are encoded as objects implementing `ExecutableMessage` found under `gno.land/p/demo/dao_maker/dao_interfaces`.
```go
type ExecutableMessage interface {
	ToJSON() *json.Node
	FromJSON(ast *json.Node)

	String() string
	Type() string
}
```

They are unmarshalled and executed by message handlers implementing `gno.land/p/demo/dao_maker/dao_interfaces.MessageHandler`.
```go
type MessageHandler interface {
	Execute(message ExecutableMessage)
	Instantiate() ExecutableMessage
	Type() string
}
```

Message handlers are registered at core creation and new message handlers can be registered via proposals to extend the DAO capabilities.

## Practical Implementation
In this section, we will showcase how to implement your own DAO using the `dao_maker` package suite.

### Setting Up Your Workspace

#### Setup the Tooling

To setup your tooling, see [Getting Started: Local Setup](../getting-started/local-setup.md).

#### Create a new Gno module

- Create a new directory and move into it: `mkdir my-gno-dao && cd my-gno-dao`
- Initialize the gno module: `gno mod init gno.land/r/<your_namespace>/my_dao`

### Creating the Voting Module

We will start by instantiating a voting module.

1. **Initialize the Factory**

Modules instantiation uses the factory pattern in case the module needs to access the core.

`my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/dao_maker/dao_interfaces"
)

func init() {
    votingModuleFactory := func(core dao_interfaces.IDAOCore) {
   
    }
}
```

2. **Instantiate the module**

`my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/dao_maker/dao_interfaces"
    "gno.land/p/demo/dao_maker/dao_voting_group" // <- new
)

func init() {
    var group *dao_voting_group.VotingGroup // <- new

    votingModuleFactory := func(core dao_interfaces.IDAOCore) {
        group = dao_voting_group.NewVotingGroup() // <- new
    }
}
```

We need to keep a reference to the module to instantiate its message handlers later.

3. **Add Initial Members and return the module**

`my_dao.gno`
```go
func init() {
    votingModuleFactory := func(core dao_interfaces.IDAOCore) {
        group = dao_voting_group.NewVotingGroup()
        group.SetMemberPower("your-address", 1) // <- new
        // repeat for any other initial members you want in the DAO
        return group // <- new
    }
}
```

Now let's create a proposal module.

### Creating the proposal module

1. **Initialize the Factory**

`my_dao.gno`
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

`my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/dao_maker/dao_interfaces"
    "gno.land/p/demo/dao_maker/dao_voting_group"
    "gno.land/p/demo/dao_maker/dao_proposal_single" // <- new
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

We also need to keep a reference to the module to instantiate it's message handlers later.

### Registering Message Handlers

Add message handlers to allow your DAO to perform specific actions when proposals are executed.

`my_dao.gno`
```go
package my_dao

import (
    "gno.land/p/demo/dao_maker/dao_interfaces"
    "gno.land/p/demo/dao_maker/dao_voting_group"
    "gno.land/p/demo/dao_maker/dao_proposal_single"
)

func init() {
    // ...
    messageHandlersFactories := []dao_interfaces.MessageHandlerFactory{
        // Allow to manage the voting group
        func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
            return group.UpdateMembersHandler()
        },
        // Allow to update the proposal module settings
        func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
            return dao_proposal_single.NewUpdateSettingsHandler(proposalModule)
        },
    }
}
```

### Creating the DAO Core

Now we can create the actual DAO.

```go
package my_dao

import (
    "gno.land/p/demo/dao_maker/dao_interfaces"
    "gno.land/p/demo/dao_maker/dao_voting_group"
    "gno.land/p/demo/dao_maker/dao_proposal_single"
    "gno.land/p/demo/dao_maker/dao_core" // <- new
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

We also need to expose the DAO methods in the realm.

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

## Conclusion

That's it! You've successfully created your first DAO using the Gno DAO framework. To expand its capabilities, you can register additional message handlers or even create new modules if you feel bold.