---
id: write-simple-dapp
---

# How to write a simple dApp on Gno.land

## Overview

This guide will show you how to write a complete dApp that combines both a package and a realm.
Our app will allow any user to create a poll, and subsequently vote
YAY or NAY for any poll that has not exceeded the voting deadline.

## Prerequisites

- **Text editor**

## Defining dApp functionality

Our dApp will consist of a Poll package, which will handle all things related to the Poll struct,
and a Poll Factory realm, which will handle the user-facing functionality and rendering.

For simplicity, we will define the functionality in plain text, and leave comments explaining the code.

### Poll Package

- Defines a `Poll` struct
- Defines a `NewPoll` constructor
- Defines `Poll` field getters
- Defines a `Vote` function
- Defines a `HasVoted` check method
- Defines a `VoteCount` getter method

[embedmd]:# (../assets/how-to-guides/write-simple-dapp/poll-1.gno go)
```go
package poll

import (
	"std"

	"gno.land/p/demo/avl"
)

// Main struct
type Poll struct {
	title       string
	description string
	deadline    int64     // block height
	voters      *avl.Tree // addr -> yes / no (bool)
}

// Getters
func (p Poll) Title() string {
	return p.title
}

func (p Poll) Description() string {
	return p.description
}

func (p Poll) Deadline() int64 {
	return p.deadline
}

func (p Poll) Voters() *avl.Tree {
	return p.voters
}

// Poll instance constructor
func NewPoll(title, description string, deadline int64) *Poll {
	return &Poll{
		title:       title,
		description: description,
		deadline:    deadline,
		voters:      avl.NewTree(),
	}
}

// Vote Votes for a user
func (p *Poll) Vote(voter std.Address, vote bool) {
	p.Voters().Set(string(voter), vote)
}

// HasVoted vote: yes - true, no - false
func (p *Poll) HasVoted(address std.Address) (bool, bool) {
	vote, exists := p.Voters().Get(string(address))
	if exists {
		return true, vote.(bool)
	}
	return false, false
}

// VoteCount Returns the number of yay & nay votes
func (p Poll) VoteCount() (int, int) {
	var yay int

	p.Voters().Iterate("", "", func(key string, value interface{}) bool {
		vote := value.(bool)
		if vote == true {
			yay = yay + 1
		}
	})
	return yay, p.Voters().Size() - yay
}
```

A few remarks:

- We are using the `std` library for accessing blockchain-related functionality and types, such as `std.Address`.
- Since the `map` data type is not deterministic in Go, we need to use the AVL tree structure, defined
  under `p/demo/avl`.
  It behaves similarly to a map; it maps a key of type `string` onto a value of any type - `interface{}`.
- We are importing the `p/demo/avl` package directly from on-chain storage, which can be accessed through the
  path `gno.land/`.
  As of October 2023, you can find already-deployed packages & libraries which provide additional Gno functionality in
  the [monorepo](https://github.com/gnolang/gno), under the `examples/gno.land` folder.

:::info
After testing the `Poll` package, we need to deploy it in order to use it in our realm.
Check out the [deployment](deploy.md) guide to learn how to do this.
:::

### Poll Factory Realm

Moving on, we can create the Poll Factory realm.

The realm will contain the following functionality:

- An exported `NewPoll` method, to allow users to create polls
- An exported `Vote` method, to allow users to pledge votes for any active poll
- A `Render` function to display the realm state

[embedmd]:# (../assets/how-to-guides/write-simple-dapp/poll-2.gno go)
```go
package poll

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/poll"
	"gno.land/p/demo/ufmt"
)

// state variables
var (
	polls         *avl.Tree // id -> Poll
	pollIDCounter int
)

func init() {
	polls = avl.NewTree()
	pollIDCounter = 0
}

// NewPoll - Creates a new Poll instance
func NewPoll(title, description string, deadline int64) string {
	// get block height
	if deadline <= std.GetHeight() {
		return "Error: Deadline has to be in the future."
	}

	// convert int ID to string used in AVL tree
	id := ufmt.Sprintf("%d", pollIDCounter)
	p := poll.NewPoll(title, description, deadline)

	// add new poll in avl tree
	polls.Set(id, p)

	// increment ID counter
	pollIDCounter = pollIDCounter + 1

	return ufmt.Sprintf("Successfully created poll #%s!", id)
}

// Vote - vote for a specific Poll
// yes - true, no - false
func Vote(pollID int, vote bool) string {
	// get txSender
	txSender := std.GetOrigCaller()

	id := ufmt.Sprintf("%d", pollID)
	// get specific Poll from AVL tree
	pollRaw, exists := polls.Get(id)

	if !exists {
		return "Error: Poll with specified doesn't exist."
	}

	// cast Poll into proper format
	poll, _ := pollRaw.(*poll.Poll)

	voted, _ := poll.HasVoted(txSender)
	if voted {
		return "Error: You've already voted!"
	}

	if poll.Deadline() <= std.GetHeight() {
		return "Error: Voting for this poll is closed."
	}

	// record vote
	poll.Vote(txSender, vote)

	// update Poll in tree
	polls.Set(id, poll)

	if vote == true {
		return ufmt.Sprintf("Successfully voted YAY for poll #%s!", id)
	}
	return ufmt.Sprintf("Successfully voted NAY for poll #%s!", id)
}
```

With that we have written the core functionality of the realm, and all that is left is
the [Render function](http://localhost:3000/explanation/realms).
Its purpose is to help us display the state of the realm in Markdown, by formatting the state into a string buffer:

Add this library:
```go
import (
	...
	"bytes"
)
```

[embedmd]:# (../assets/how-to-guides/write-simple-dapp/poll-3.gno go)
```go
func Render(path string) string {
	var b bytes.Buffer

	b.WriteString("# Polls!\n\n")

	if polls.Size() == 0 {
		b.WriteString("### No active polls currently!")
		return b.String()
	}
	polls.Iterate("", "", func(key string, value interface{}) bool {

		// cast raw data from tree into Poll struct
		p := value.(*poll.Poll)
		ddl := p.Deadline()

		yay, nay := p.VoteCount()
		yayPercent := 0
		nayPercent := 0

		if yay+nay != 0 {
			yayPercent = yay * 100 / (yay + nay)
			nayPercent = nay * 100 / (yay + nay)
		}

		b.WriteString(
			ufmt.Sprintf(
				"## Poll #%s: %s\n",
				key, // poll ID
				p.Title(),
			),
		)

		dropdown := "<details>\n<summary>Poll details</summary><br>"

		b.WriteString(dropdown + "Description: " + p.Description())

		b.WriteString(
			ufmt.Sprintf("<br>Voting until block: %d<br>Current vote count: %d",
				p.Deadline(),
				p.Voters().Size()),
		)

		b.WriteString(
			ufmt.Sprintf("<br>YAY votes: %d (%d%%)", yay, yayPercent),
		)
		b.WriteString(
			ufmt.Sprintf("<br>NAY votes: %d (%d%%)</details>", nay, nayPercent),
		)

		dropdown = "<br><details>\n<summary>Vote details</summary>"
		b.WriteString(dropdown)

		p.Voters().Iterate("", "", func(key string, value interface{}) bool {

			voter := key
			vote := value.(bool)

			if vote == true {
				b.WriteString(
					ufmt.Sprintf("<br>%s voted YAY!", voter),
				)
			} else {
				b.WriteString(
					ufmt.Sprintf("<br>%s voted NAY!", voter),
				)
			}
			return false
		})

		b.WriteString("</details>\n\n")
		return false
	})
	return b.String()
}
```

## Conclusion

That's it 🎉

You have successfully built a simple but fully-fledged dApp using Gno!
Now you're ready to conquer new, more complex dApps in Gno.
