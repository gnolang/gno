# Gno Moderation DAO

## Overview

The general flow is

- Users flag harmful content
- Moderators (members of the moderation DAO) can review the list of most flagged content
- A moderator creates a proposal to ban some content
- Moderators vote on the ban proposal, the tallying process is fully customizable using proposal modules, see the [DAO framework documentation](./DAO_TUTORIAL.md) for details
- Once the proposal is passed, it can be executed and the content is removed

A Gno moderation DAO is comprised of two main parts

- The Gno DAO framework, that allows to create generic DAOs. You can read about it in details [here](./DAO_TUTORIAL.md)
- A generic [flagging package](/examples/gno.land/p/demo/teritori/flags_index/) integrated with the target content realm, it allows user to flag harmful content. It stores flagged ids and allow to query most flagged content

There is currently two integrations, one for a [fork of the boards realm](./modboards), meant to be used in gnoweb + cli and one for the [teritori social feed](./social_feeds) meant to be used by a browser user

## Integration

In this section we will see how decentralize moderation was integrated in the boards realm

### Flagging

First we need to instantiate the flags index somewhere

```go
import (
    "gno.land/p/demo/teritori/flags_index"
)

flags := flags_index.NewFlagsIndex()
```

In the boards integration, this object is stored as a member of the board object.

The flags index works on string IDs and we need to map content IDs to flag ID which are strings, here is how it's done for boards:

```go
func getFlagID(threadID PostID, postID PostID) flags_index.FlagID {
	return flags_index.FlagID(threadID.String() + "-" + postID.String())
}
```

In this case there is a flags index for each board so we don't need to put the board ID in the flag ID.

Then we need to expose the flagging method on the realm, using the flags index's `Flag(flagID FlagID, flaggerID string)` method. Here's how it's done for boards:

```go
func FlagPost(boardID BoardID, threadID PostID, postID PostID) {
	// check that the target exists
	board := getBoard(boardID)
	if board == nil {
		panic("board not exist")
	}
	thread := board.GetThread(threadID)
	if thread == nil {
		panic("thread not exist")
	}
	if postID != threadID {
		post := thread.GetReply(postID)
		if post == nil {
			panic("post not exist")
		}
	}

	// flag
	board.flags.Flag(getFlagID(threadID, postID), std.PrevRealm().Addr().String())
}
```

Now we need a way to show the most flagged content, in the boards integration, this is done in the `Render` function

```go
func Render(path string) string {
	// (...)
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// (...)
	} else if len(parts) == 2 {
		name := parts[0]
		boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		board := boardI.(*Board)

		if parts[1] == "flags" {
			// /r/demo/modboards:BOARD_NAME/flags
			return board.RenderFlags(1000, 0) // TODO: pagination
		}
        // (...)
    }
    // (...)
}
```

We also need to clear flags when content is removed so it's also removed from the flagged content list, for boards this is done in the `DeletePost` routine

```go
func DeletePost(bid BoardID, threadid, postid PostID, reason string) {
	// (...)
	board.flags.ClearFlagCount(getFlagID(threadid, postid))
}
```

### Moderation

Now that the moderators can review the most flagged content, we need to enable the moderation flow

1. **Create a DAO**

The first thing we need to do is setup a DAO, we won't go into the details here but you can follow the [DAO framework quickstart guide](./DAO_TUTORIAL.md) to learn more about it

2. **Extend the DAO to delete content**

Once we have a DAO, we need to allow the DAO to delete content via an ExecutableMessage and a MessageHandler, for boards this is done like so:

```go
type ExecutableMessageDeletePost struct {
	dao_interfaces.ExecutableMessage

	BoardID  BoardID
	ThreadID PostID
	PostID   PostID
	Reason   string
}

func (msg ExecutableMessageDeletePost) Type() string {
	return "gno.land/r/demo/teritori/modboards.DeletePost"
}

func (msg *ExecutableMessageDeletePost) String() string {
	// Code excluded for brevity, this method is used in gnoweb renders of the DAO
}

func (msg *ExecutableMessageDeletePost) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "boardId", Value: msg.BoardID},
		{Key: "threadId", Value: msg.ThreadID},
		{Key: "postId", Value: msg.PostID},
		{Key: "reason", Value: msg.Reason},
	})
}

type DeletePostHandler struct {
	dao_interfaces.MessageHandler
}

func NewDeletePostHandler() *DeletePostHandler {
	return &DeletePostHandler{}
}

func (h *DeletePostHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageDeletePost)
	DeletePost(msg.BoardID, msg.ThreadID, msg.PostID, msg.Reason)
}

func (h DeletePostHandler) Type() string {
	return ExecutableMessageDeletePost{}.Type()
}

func (h *DeletePostHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageDeletePost{}
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "boardId", Value: &msg.BoardID},
		{Key: "threadId", Value: &msg.ThreadID},
		{Key: "postId", Value: &msg.PostID},
		{Key: "reason", Value: &msg.Reason},
	})
	return msg
}
```

This handler can be registered at DAO instanciation like so:

```go
	messageHandlersFactories := []dao_interfaces.MessageHandlerFactory{
		// (...)
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return modboards.NewDeletePostHandler()
		},
		// (...)
	}
```

Or later via a proposal

3. **Create a delete proposal**

Now that we have a DAO that can delete content, moderators can create delete proposals using the `ProposeJSON(moduleIndex int, proposalJSON string) int` function of DAO realms, with cli, a local node and the single choice proposal module, this looks like this:

```bash
gnokey maketx call -pkgpath "gno.land/r/demo/teritori/boards_moderation_dao" -func "ProposeJSON" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "0" -args '{"title": "Ban content", "description": "", "messages": [{"type": "gno.land/r/demo/teritori/modboards.DeletePost", "payload": {"boardId": 1, "threadId": 1, "postId": 1, "reason": "Does not comply with community guidelines"}}]}' -remote "localhost:26657" wallet-name
```

4. **Vote on the proposal**

We can then vote on this proposal, for example:

```bash
gnokey maketx call -pkgpath "gno.land/r/demo/teritori/boards_moderation_dao" -func "VoteJSON" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "0" -args "0" -args '{"vote": 0, "rationale": "This indeed does not comply"}' -remote "localhost:26657" wallet-name
```

5. **Execute proposal**

If the proposal is passed, we can then execute the proposal to remove the content

```bash
gnokey maketx call -pkgpath "gno.land/r/demo/teritori/boards_moderation_dao" -func "Execute" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "0" -args "0" -remote "localhost:26657" wallet-name
```

Although we show an example using CLI, this flow can be done through UI and this is how it's currently done for the teritori social feed moderation POC. For the boards POC, the flow is made slightly easier than raw cli using the gnoweb pre-fill help system

### Further improvements

We would like to do the following improvements next:

- Improve how proposals are linked to content, currently we do this inefficiently in the UI by searching through all proposals
- Automate the proposal creation process, for example when the flag count for a content exceeds a particular threshold
- Creating incentives for flagging and moderation, for example, flagging content could require a deposit that is transfered to the moderation DAO if the content is not considered harmful to prevent flagging spam. Another nice-to-have incentive would be that you need to make a deposit to start posting and if some of your content get banned, this deposit is transfered to the moderation DAO and you need to make a new deposit to post again. We could go even further and put an appeal system in place that would use a justice DAO and would penalize moderators that wrongly banned an abiding content. Of course these are just ideas and we would need to think about the game theoretics of all these potential incentives.