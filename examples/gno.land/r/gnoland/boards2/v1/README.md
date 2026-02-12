# Boards2

Boards2 is a social discussion forum for open communication and community-driven conversations.

Users can start discussions by creating or reposting threads and then submitting comments or replies to
other user comments.

Discussions happen within different boards, where each board is an independent self managed community.

By default Boards2 allows users to create two types of boards, one is the invite only board where only
invited users can create threads and comment, and where non invited users can only read the content and
discussions; The other type of board is the open board where any user with a registered Gno.land username
and a specific amount of GNOT in their account can create threads, repost and comment.

## Open Boards Quick Start

If you don’t have a registered Gno.land username, or are new to Gno.land in general, the quick start guide
below can help you get started quickly.

What you need to create threads and start commenting within open boards is a registered username and having
a specific amount of GNOT in your Gno.land user account, which by default initially is 3000 GNOT. This
initial GNOT amount could be changed over time to a different amount, so this requirement can change.

### How To Get a Gno.land Address

To use Boards2 you'll need:

1. A Gno.land address, and
2. Some GNOT tokens to register a Gno.land username and to be able to interact with Boards2

You can quickly setup your account using [Adena] or any Gno.land compatible wallet by following these steps:

- Download [Adena], or a Gno.land compatible wallet
- Once installed, you have to create a new account or add an existing one following wallet's instructions
- If you don't have GNOT you will need to use a faucet to get some, if the network allows it

For testing networks you can use the official [Faucet Hub] to receive GNOT in your account.

### How to Register Your Username

If you don't have it, to register a new Gno.land username visit the [users realm] and follow the steps
described there.

Once you register a username it will appear on all your threads, posts and comments.

### How to Start Using Open Boards

Once you have a username and the required GNOT amount in your account you can start commenting, creating and
reposting threads within any open board.

To comment and engage on an open board discussion visit a thread and click on the "Comment" link. You can
also reply to any of the thread's comments by clicking on the "Reply" link.

To create threads, visit an open board and then click on the "Create Thread" link, there you will have to
enter a title and some content for the thread body.

Thread and comments content can be written as plaintext, or Markdown if you want to format the content so
it's rendered as rich text.

You can also repost any thread, even the ones from invite only boards, into any open board. To do so visit
the thread you want to repost and click on the "Repost" link at the bottom of the thread, there you will have
to enter the open board where you want the repost to be created, a title for the thread repost and optionally
also some content to render at the top of the repost. The optional content can also be written as plaintext
or Markdown, like threads.

After your thread, repost or comment is created, you can easily share the link with others so they can join
the discussion!

## Boards

Boards2 realm enables the creation of different communities though independent boards.

When a board is created, and independetly of the board type, it initially has a single "owner" member
assigned by default, which is the user that creates it. The member is called "owner" because by default it
has the `owner` role, which grants all permissions within that board.

Members of a board with the `owner` or `admin` role, independently of the board type, can invite other
members, or otherwise users can request being invited to be a member by visiting the board and clicking the
"Request Invite" link. Requested invites can be accepted or revoked using these public realm functions:

```go
// AcceptInvite accepts a board invite request.
func AcceptInvite(_ realm, boardID boards.ID, user address)

// RevokeInvite revokes a board invite request
func RevokeInvite(_ realm, boardID boards.ID, user address)
```

There are four possible roles that invited users can have when they are members of a board:
- `owner`: Grants all available permissions
- `admin`: Grants basic, moderator and advanced permissions, like being able to rename boards, add or remove
   members, or change their role.
- `moderator`: Grants basic and moderation related permissions, like being able to ban or unban users, or
  flag content.
- `guest`: Grants basic permissions that allow creating threads, reposting and commenting.

No roles or number of members is enforced for boards, so technically a board can be updated to have no
members, or for example boards could exists without any "owner" if all members with `owner` role are removed
from it.

Other custom user defined roles can exists on top of the default ones though [custom board] implementations.

### Custom Board Permissions

TODO

### Boards Governance

By default boards are created with an undelying DAO, so each new board is linked to an independent DAO which
is used to organize members by role, and can also be used to update boards in a permissionless manner.

Current Boards2 realm implementation doesn't run proposals, but some of the mechanics that are currently
implemented through roles and permissions will rely on DAO proposals to actually execute changes. **Default
support for proposals is going to be implemented in upcoming Boards2 versions**.

Right now is possible to integrate with the underlying DAO and change the default board mechanics to rely on
proposals using a [custom board] implementation, by creating a new realm that imports and uses the [custom
permissions] realm, which exposes the DAO. The new realm can then be used to replace the default board
permissions, which right now can only be done by a Boards2 realm `owner` using a public realm function:

```go
// SetPermissions sets a permissions implementation for boards2 realm or a board
func SetPermissions(_ realm, boardID boards.ID, p boards.Permissions)
```

## Moderation

### Flagging

Threads and comments are moderated by flagging, which requires the `moderator`, `admin` or `owner` roles.

A reason is required each time content is flagged by a member. Content is replaced by a feedback message
and a link to the list of flagging reasons given by moderators when a moderation flagging threshold is
reached. By default the threshold is of a single flag.

> Right now is not possible to show the content of a thread or comment that has been hidden because of
> moderation, but future Boards2 versions might implement a way to handle moderation disputes and allow
> restoring the thread or comment content.

> Boards2 realm `owners` are allowed to moderate content with a single flag within any board at this point,
> but this might be changed to work though a DAO proposal.

Each board's `owner` or `admin` members are free to change the flagging threshold within a single board to a
greater value using a public realm function:

```go
// SetFlaggingThreshold sets the number of flags required to hide a thread or comment
func SetFlaggingThreshold(_ realm, boardID boards.ID, threshold int)
```

### Banning

Members with the `moderator`, `admin` or `owner` roles are the only ones that are allowed to ban or unban
a user within a board.

Users can be banned with a reason for any number of hours. Within this period banned users are not allowed
to interact or make any changes.

Only invited guest members and open board users can be banned, banning board owners, admins and moderators is
not allowed.

Banning and unbanning can be done by calling these public realm functions:

```go
// Ban bans a user from a board for a period of time
func Ban(_ realm, boardID boards.ID, user address, hours uint, reason string)

// Unban unbans a user from a board
func Unban(_ realm, boardID boards.ID, user address, reason string)
```

## Freezing

Boards2 realm allows `owner` or `admin` members of a board to freeze the board or any of its threads.
Freezing makes the board or thread readonly, disallowing any changes or additions until unfrozen.

The following public realm function can be called for freezing:

```go
// FreezeBoard freezes a board so no more threads and comments can be created or modified
func FreezeBoard(_ realm, boardID boards.ID)

// UnfreezeBoard removes frozen status from a board
func UnfreezeBoard(_ realm, boardID boards.ID)

// FreezeThread freezes a thread so thread cannot be replied, modified or deleted
func FreezeThread(_ realm, boardID, threadID boards.ID)

// UnfreezeThread removes frozen status from a thread
func UnfreezeThread(_ realm, boardID, threadID boards.ID)
```


[users realm]: https://gno.land/r/gnoland/users/v1
[custom permissions]: https://gno.land/r/gnoland/boards2/v1/permissions/
[custom board]: #custom-board-permissions
[Adena]: https://www.adena.app/
[Faucet Hub]: https://faucet.gno.land/
