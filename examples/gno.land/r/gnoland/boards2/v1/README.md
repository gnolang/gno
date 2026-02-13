# Boards: Decentralized Discussion Forums on Gno.land

Boards is a forum realm for Gno.land with robust role-based access control, moderation tools, and a clean public API. It serves two roles: a proof-of-concept that demonstrates building real dApps on Gno.land, and a practical tool for decentralized social coordination on the internet. It features pluggable permission systems, on-chain governance with irreversible commitments, moderation that balances openness with safety, and an event-driven architecture that enables integrations and observability. For communities, it provides transparent governance structures, on-chain decision-making, configurable moderation, permanent auditable records, and flexible role management that bridges on-chain governance with human dynamics.

Access (v1): Boards operates invite-only access during beta. Creating boards and other privileged actions require realm membership. How to request/receive an invite: [TO ADD].

## Background and Evolution

The original Boards realm (`r/demo/boards`) created by Jae Kwon showcased decentralized forums on Gno.land: simple boards, threads, and posts, persisted on-chain and rendered via `Render()`. It validated the programming model, UI templating, and basic interaction loops.

The latest Boards implementation builds on that foundation to enable full social coordination at scale:
- Adds member invitations, role changes, and removals
- Formalizes moderation (flagging thresholds, hide-on-threshold, freezing at multiple scopes, bans)
- Adds governance controls (irreversible realm lock, notices, help)
- Improves reposting across boards and clarifies render behaviors for hidden/frozen content

## Key Features

- **Role-based access control** with pluggable permission engine (realm- and board-scoped)
- **Member management**: invitations, role changes, removals
- **Moderation**: per-board flag thresholds, hide-on-threshold, freezing (board/thread/reply), bans
- **Governance**: irreversible realm lock (with optional member-lock), realm/board notice/help
- **Content**: threads and nested replies, repost across boards
- **Safety rails**: validators for board names and privileged actions; creator-only reply edits
- **Observability**: rich events emitted for key actions
- **Extensibility**: swap permissions, add validators, link actions via `txlink`

## Quickstart

### Prerequisites

- A funded Gno.land account (address) and access to a signing tool (e.g., `gnokey`)
- Access is invite-only (see notice above); realm membership is required for privileged actions (invitation process: [TO ADD])
- The realm path where Boards is deployed (e.g., `gno.land/r/gnoland/boards2/v1`)
- Ability to submit `Call` transactions with appropriate gas/fees

### Basic Operations

1. **Discover or create a board** (requires realm membership for creation)
   - Resolve by name: `GetBoardIDFromName(name)`
   - Or create: `CreateBoard(name, listed)` (requires `board:create` permission)

2. **Post content**
   - Create a thread: `CreateThread(boardID, title, body)`
   - Reply to a thread: `CreateReply(boardID, threadID, 0, body)`

3. **Moderate**
   - Flag problematic content: `FlagThread(...)` / `FlagReply(...)`
   - Freeze to pause activity: `FreezeBoard(...)`, `FreezeThread(...)`
   - Ban repeat offenders: `Ban(boardID, user, hours, reason)`

4. **Manage members**
   - Invite or accept requests: `InviteMember(...)`, `AcceptInvite(...)`
   - Change roles: `ChangeMemberRole(...)`

5. **Govern**
   - Adjust thresholds: `SetFlaggingThreshold(boardID, n)`
   - Lock realm (irreversible): `LockRealm(lockRealmMembers)`

## Overview

Boards operates on two permission scopes that are fundamental to understand:

**Realm-level operations** (`boardID=0`): Control who can create boards, manage realm members, set notices, and lock the realm. Uses `gPerms` (global realm permissions).

**Board-level operations** (`boardID!=0`): Control activities within a specific board - posting, moderation, member management for that board. Each board has its own `Permissions` instance.

This dual-scope design allows realm governance (who can create boards) to be separate from board governance (how individual boards operate).

- Boards organizes content into boards → threads → replies
- Each board has its own permissions (roles, thresholds, ban-list, freeze state)
- Realm-level governance can be locked permanently and independently from member modifications
- Moderation leverages per-board flagging thresholds, freezing, and bans
- Permissions are enforced via a pluggable `Permissions` interface; the default implementation is `BasicPermissions` backed by a DAO member set

### Key Packages and Files

- Public API: `examples/gno.land/r/gnoland/boards2/v1/public.gno`
- Board model and helpers: `examples/gno.land/r/gnoland/boards2/v1/board.gno`
- Post model and helpers: `examples/gno.land/r/gnoland/boards2/v1/post.gno`
- Permissions: `examples/gno.land/r/gnoland/boards2/v1/permissions.gno`, `permissions_basic.gno`, `permissions_validators.gno`
- Moderation (flag/ban/freeze): `public_flag.gno`, `flag.gno`, `public_ban.gno`, `public_freeze.gno`
- Realm lock: `public_lock.gno`
- Invitations: `public_invite.gno`
- Rendering helpers/URIs: `render.gno`, `uris_post.gno`

## Roles and Permissions

Authorization model controls who can do what, and where. Boards uses a pluggable permission engine to enforce capabilities at realm and board scopes. Use the default mapping, refine with validators, or swap the engine with `SetPermissions` to fit governance.

- **Roles**: `owner` (superuser), `admin`, `moderator`, `guest`
- **Permissions** are verb:resource pairs (for example, `thread:create`, `board:freeze`)
- **Scope**: realm-level controls (for example, `realm:lock`) vs board-level controls
- **Validators**: additional business rules on sensitive operations (for example, inviting owners, username ownership checks)
- **Membership**: DAO-backed; non-members cannot invoke protected actions even if roles are mapped

### Roles

- **`owner`**: Super role with all permissions; used for both realm and boards
- **`admin`**: Board creation, member management, full moderation capabilities
- **`moderator`**: Content moderation and user management
- **`guest`** (empty string): Default/lightweight member role for basic content creation

### Permissions (Selected)

- **Realm-level**: `realm:help`, `realm:lock`, `realm:notice`, `permissions:update` (also applicable to boards)
- **Board-level creation**: `board:create`, `board:rename`, `board:freeze`, `board:flagging-update`
- **Threads**: `thread:create`, `thread:edit`, `thread:delete`, `thread:flag`, `thread:freeze`, `thread:repost`
- **Replies**: `reply:create`, `reply:delete`, `reply:flag`, `reply:freeze`
- **Members**: `member:invite`, `member:invite-remove`, `member:remove`, `role:change`
- **Users (moderation)**: `user:ban`, `user:unban`

See `permissions.gno` for the complete list.

### Default Role → Permission Mapping

**Realm permissions initialization** (`boards.gno:initRealmPermissions`, `permissions_basic.gno:createBasicPermissions`):
- `owner`: super role (all permissions)
- `admin`: `board:create`

**Board permissions** (`board.gno:createBasicBoardPermissions`):
- `owner`: super role (all board permissions)
- `admin`: `board:rename`, `board:flagging-update`, `member:invite`, `member:invite-remove`, `member:remove`, `thread:create`, `thread:edit`, `thread:delete`, `thread:repost`, `thread:flag`, `thread:freeze`, `reply:create`, `reply:delete`, `reply:flag`, `reply:freeze`, `role:change`, `user:ban`, `user:unban`
- `moderator`: `thread:create`, `thread:edit`, `thread:repost`, `thread:flag`, `reply:create`, `reply:flag`, `user:ban`, `user:unban`
- `guest`: `thread:create`, `thread:repost`, `reply:create`

**Critical Implementation Details**:
- `owner` is set as a super role via `BasicPermissions.SetSuperRole`, bypassing per-permission mapping checks
- `BasicPermissions.WithPermission` enforces BOTH permission AND DAO membership: `if !bp.HasPermission(user, p) || !bp.dao.Members().Has(user)` - users must be DAO members AND have the permission
- `SetUserRoles` automatically adds users to the DAO if they don't exist: `if !bp.HasUser(user) { bp.dao.Members().Add(user) }`

### Permission Validators

Validators apply at both realm and board scope, registered during permission system initialization:

**Realm-level validators** (registered in `initRealmPermissions`):
- `PermissionBoardCreate` → `validateBoardCreate`
- `PermissionMemberInvite` → `validateMemberInvite` 
- `PermissionRoleChange` → `validateRoleChange`

**Board-level validators** (registered in `createBasicBoardPermissions`):
- `PermissionBoardRename` → `validateBoardRename`
- `PermissionMemberInvite` → `validateMemberInvite`
- `PermissionRoleChange` → `validateRoleChange`

**Validator Logic**:
- `validateBoardCreate/Rename`: Checks `std.Address(name).IsValid()` to reject addresses as names; calls `users.ResolveName(name)` and ensures caller owns any matching registered username
- `validateMemberInvite`: Prevents non-owners from inviting owners: `if role == RoleOwner && !perms.HasRole(caller, RoleOwner)`
- `validateRoleChange`: Additional role transition constraints

### Role-Permission Matrix (Common Actions)

| Action | Guest | Moderator | Admin | Owner |
|---|---|---|---|---|
| Create Board |  |  | ✓ (realm) | ✓ |
| Rename Board |  |  | ✓ | ✓ |
| Freeze Board |  |  | ✓ | ✓ |
| Set Flag Threshold |  |  | ✓ | ✓ |
| Invite Member |  |  | ✓ | ✓ |
| Remove/Revoke Invite |  |  | ✓ | ✓ |
| Change Member Role |  |  | ✓ | ✓ |
| Create Thread | ✓ | ✓ | ✓ | ✓ |
| Edit Thread |  | ✓ | ✓ | ✓ |
| Delete Thread |  |  | ✓ | ✓ |
| Repost Thread | ✓ | ✓ | ✓ | ✓ |
| Flag Thread |  | ✓ | ✓ | ✓ |
| Freeze Thread |  |  | ✓ | ✓ |
| Create Reply | ✓ | ✓ | ✓ | ✓ |
| Edit Reply | (creator only) | (creator only) | (creator only) | (creator only) |
| Delete Reply |  |  | ✓ | ✓ |
| Flag Reply |  | ✓ | ✓ | ✓ |
| Freeze Reply |  |  | ✓ | ✓ |
| Ban/Unban User |  | ✓ | ✓ | ✓ |

**Notes**:
- Owners are super users; they implicitly have all permissions
- Reply edit is creator-only regardless of role
- Exact permissions are defined in code; this table summarizes defaults

## Membership and Invitations

Access model (v1): realm-level access is invite-only during beta. Board-level invitations are supported via the public API. Specific steps for requesting an invite to the realm: [TO ADD].

Membership is the gateway to capabilities. You can proactively invite members with roles, or accept self-service requests. Role changes elevate or limit powers, and removals revoke membership. When the realm's member set is locked (see Governance), these flows halt by design.

Invitations ensure boards aren't flooded; requests let community members self-onboard. Accepting an invite request adds the user as a member without a role by default; promote if warranted. Removing a member demotes them to external user; their public content persists (immutable history), but they lose special powers. Realm member-lock prevents further changes to the member graph once finalized.

### Public Methods

(`public_invite.gno`, `public.gno`):

- **`InviteMember(boardID, user, role)`**: Invite a user to the realm or a specific board; role optional; requires `member:invite` and passes validator (owners inviting owners)
- **`RequestInvite(boardID)`**: External user requests to join a board; stored in per-board `gInviteRequests`; idempotent
- **`AcceptInvite(boardID, user)`**: Accepts a stored invite request; requires `member:invite` and adds the user as a member without a role by default
- **`RevokeInvite(boardID, user)`**: Revokes an existing invite request; requires `member:invite-remove`
- **`RemoveMember(boardID, member)`**: Remove a member (see `public.gno`)
- **`ChangeMemberRole(boardID, member, role)`**: Update a member's role; requires `role:change` and passes validators

Realm can optionally lock member modifications (see Locking below).

## Moderation

Moderation provides safety controls. Flags provide signal; thresholds automatically hide content; freezing pauses activity at the chosen scope; bans restrict abusive users. Realm owners retain limited controls in frozen contexts for remediation.

### Flagging and Hiding Content

Per-board flagging threshold controls when a post gets hidden:
- `SetFlaggingThreshold(boardID, threshold)` (admin) updates threshold; default is `1`
- Actual threshold lookup in `flag.gno:getFlaggingThreshold` backed by `gFlaggingThresholds`

**Actions**:
- `FlagThread(boardID, threadID, reason)` requires `thread:flag` (unless realm owner)
- `FlagReply(boardID, threadID, replyID, reason)` requires `reply:flag` (unless realm owner)

**Behavior**:
- When flags reach the threshold, the item becomes `Hidden`
- Realm owners can hide with a single flag and can flag even when the board is frozen
- Double-flagging by the same user is rejected; exceeding threshold panics (defensive invariants)

### Freezing

Freeze prevents edits/deletes and new activity at the chosen scope:

**Actions**:
- `FreezeBoard(boardID)` / `UnfreezeBoard(boardID)` require `board:freeze`
- `FreezeThread(boardID, threadID)` / `UnfreezeThread(...)` require `thread:freeze`
- `FreezeReply(boardID, threadID, replyID)` / `UnfreezeReply(...)` require `reply:freeze`

**Checks**:
- Board freeze blocks thread/reply changes; thread freeze blocks reply changes; reply freeze blocks that reply only
- Freeze helpers assert caller permissions and current non-frozen state where applicable

### Banning Users

- `Ban(boardID, user, hours, reason)` and `Unban(boardID, user, reason)` require `user:ban`/`user:unban`
- Only `guest` members or external users can be banned; owners/admins/moderators cannot be banned
- Bans are enforced via `assertUserIsNotBanned` which blocks posting and editing actions

## Realm Governance and Locking

Locking is irreversible and can optionally lock the member set. Use only when necessary because it changes allowed actions across the realm.

`LockRealm(lockRealmMembers)` (requires `realm:lock`) permanently locks the realm:
- If `lockRealmMembers` is false, realm content mutability is locked but members can still be updated in a subsequent call
- If true, both realm and member modifications are locked. A locked state is irreversible
- Status checks: `IsRealmLocked()`, `AreRealmMembersLocked()`; internal guards prevent state changes when locked

## Boards, Threads, Replies — CRUD

This is the operational core. Creating, editing, deleting, and reposting all respect moderation and governance constraints. Input validation protects the UI and users. Think of CRUD as the "how," governed by Roles/Membership (who) and filtered by Moderation/Governance (when).

### Boards

- **`CreateBoard(name, listed)`** (realm) — Validates name and permissions (`board:create`), reserves a unique `BoardID`, creates per-board permissions for the creator, and indexes by name and id
- **`RenameBoard(name, newName)`** — Requires `board:rename` and passes validators; preserves aliases (old names remain resolvable)
- **`GetBoardIDFromName(name)`** — Lookup by name

### Threads

- **`CreateThread(boardID, title, body)`** — Requires permission and non-banned caller; asserts non-empty title/body; returns `PostID`
- **`EditThread(boardID, threadID, title, body)`** — Creator or `thread:edit` can update; body may be empty only for reposts
- **`DeleteThread(boardID, threadID)`** — Creator can delete; otherwise `thread:delete`; realm owners may bypass some checks even when realm is locked
- **`CreateRepost(boardID, threadID, title, body, destinationBoardID)`** — Copies a thread to another board; requires `thread:repost` on destination

### Replies

- **`CreateReply(boardID, threadID, replyID, body)`** — Creates a reply to a thread or another reply (`replyID`=0 for top-level reply)
- **`EditReply(boardID, threadID, replyID, body)`** — Only the reply creator can edit; asserts non-empty body and non-frozen ancestors
- **`DeleteReply(boardID, threadID, replyID)`** — Creator or users with `reply:delete` (see `public.gno`)

### Input Validation Highlights

- **Board names**: 3–50 characters; start with a letter; use letters, numbers, underscore, or dash; cannot be a valid address. If the name matches a registered username, only that username’s owner can create the board.
- **Thread titles**: Max 100 characters.
- **Reply bodies**: Max 300 characters. Lines starting with headings, horizontal rules, or blockquotes (for example, "#", "---", ">") are not allowed to prevent UI breakage.
- **All inputs**: Leading and trailing whitespace is trimmed before validation

## Data Model (Key Structures)

A shared mental model helps contributors reason about behavior and state. `Board` owns threads and configuration; `Post` models both threads and replies; global indexes enable efficient lookups and pagination. These structures inform both the API and rendering.

### Board (`board.gno`)

- **Fields**: `ID`, `Name`, `Aliases`, `Creator`, `Readonly`, `perms`, `threads`, `postsCtr`, `createdAt`
- **Methods**: iteration helpers, post/thread accessors, ID/key helpers

### Post (`post.gno`)

- Threads and replies are both `Post` with IDs; threads have `PostID == ThreadID`
- **Key fields include**: `ID`, `ThreadID`, `ParentID`, `Creator`, `Title`, `Body`, `Readonly`, `Hidden`, reply trees, reposts, flags, timestamps
- **Methods**: `AddReply`, `DeleteReply`, `Repost`, `Flag`, `FlagsCount`, render helpers

### Global Indexes (`boards.gno`)

- `gBoardsByID`, `gBoardsByName`, `gListedBoardsByID`
- `gInviteRequests` (per board), `gBannedUsers` (per board)
- `gFlaggingThresholds` (per board)
- `gLocked` (realm and member lock status)

## Permission System Architecture

Under the hood, the `Permissions` interface powers RBAC. `BasicPermissions` is the default, DAO-backed and validator-augmented. Because it's swappable, you can evolve governance without rewriting business logic—register new validators or drop in a new engine.

`Permissions` interface (`permissions.gno`) defines capability checks and member iteration:
- `HasRole(addr, role)`, `HasPermission(addr, perm)`
- `WithPermission(addr, perm, args, func(args))` — membership + permission + validator gating before executing callback
- `SetUserRoles`, `RemoveUser`, `HasUser`, `UsersCount`, `IterateUsers`

Default implementation: `BasicPermissions` (`permissions_basic.gno`)
- Internals: per-role permission mapping, per-user role mapping, validator registry, and DAO-backed membership (`commondao`)
- Super role support via `SetSuperRole(RoleOwner)`
- Validator registration via `ValidateFunc(permission, fn)`

**Realm vs Board Permission Architecture**:

**Realm permissions** (`gPerms` global):
- Initialized in `init()` with hardcoded owners (@devx, @moul) as super users
- Controls board creation (`board:create` for admins), realm locking, notices, permission engine swapping
- Uses single `BasicPermissions` instance with `commondao.New()`

**Board permissions** (per-board):
- Created fresh for each board in `CreateBoard` via `createBasicBoardPermissions(caller)`
- Board creator becomes `owner` (super user) of that specific board
- Controls all board-specific operations (posting, moderation, member management within that board)
- Each board has independent DAO membership, roles, and validators

**Permission Engine Swapping**:
- `SetPermissions(0, newPerms)`: Replace realm-level permission engine (`gPerms`)
- `SetPermissions(boardID, newPerms)`: Replace specific board's permission engine (`board.perms`)

## Rendering and UI Composition

Actions are first-class links. `txlink.Realm` (`gRealmLink`) wires UI controls directly to public calls, and URI helpers build safe, contextual links. Hidden/frozen content renders intentionally to preserve context without amplifying abuse.

**Rendering System Details**:
- The UI generates callable URLs for public actions
- URI helpers build contextual action links (create/edit/delete/flag)
- Routing maps paths like: `""` (board list), `"{board}"`, `"{board}/{thread}"`, `"{board}/{thread}/{reply}"`
- Controls appear/disappear based on permission checks and freeze state
- Hidden content shows `"⚠ Reply is hidden as it has been flagged as inappropriate"` while preserving thread structure
- Frozen content removes Edit/Delete/Reply links but keeps Flag/View links

### Gnoweb Integration, Capabilities, and Limitations

Boards renders through gnoweb components/templates and inherits their behaviors:

Capabilities:
- Server-side rendering using `gno.land/pkg/gnoweb/components` (views, layouts, templates)
- Markdown-to-HTML rendering with safe link construction via `txlink.Realm`
- Path routing and pagination via `mux.Router` and `pager`
- Permission-aware UI: controls render only when the caller has capability

Dependencies:
- Templates and components from gnoweb: headers, layouts, UI partials
- `md`, `mdtable`, and helper utilities used by views
- Stable path routing assumptions: index, board, thread, reply

Current constraints (subject to change as gnoweb evolves):
- No SPA/JS interactivity; actions are executed via transaction links (Call URLs)
- No live updates; users refresh to see new state
- Rendering is markdown-first; complex rich text is constrained
- Hidden/frozen content intentionally limits visible interaction affordances
- URI construction depends on realm path stability; moving the realm path requires updating links

These are current capabilities focused on determinism, auditability, and clear permission boundaries. A separate frontend can provide a richer UX over the same public API and events.

## API Reference

### Realm-Level Functions

- **`SetHelp(content)`**: Update realm help content
- **`SetPermissions(0, permissions)`**: Replace realm permission engine
- **`SetRealmNotice(message)`**: Set global realm notice (requires `thread:create` permission)
- **`LockRealm(lockRealmMembers)`**: Irreversibly lock realm

### Board Management

- **`CreateBoard(name, listed)`**: Create new board
- **`RenameBoard(name, newName)`**: Rename existing board
- **`GetBoardIDFromName(name)`**: Resolve board ID by name
- **`FreezeBoard(boardID)`** / **`UnfreezeBoard(boardID)`**: Control board mutability
- **`SetFlaggingThreshold(boardID, threshold)`** / **`GetFlaggingThreshold(boardID)`**: Manage content moderation

### Thread Operations

- **`CreateThread(boardID, title, body)`**: Create new discussion thread
- **`EditThread(boardID, threadID, title, body)`**: Modify existing thread
- **`DeleteThread(boardID, threadID)`**: Remove thread and all replies
- **`CreateRepost(boardID, threadID, title, body, destinationBoardID)`**: Cross-post thread
- **`FlagThread(boardID, threadID, reason)`**: Flag thread for review

### Reply Operations

- **`CreateReply(boardID, threadID, replyID, body)`**: Create reply to thread or reply
- **`EditReply(boardID, threadID, replyID, body)`**: Modify reply (creator only)
- **`DeleteReply(boardID, threadID, replyID)`**: Remove reply
- **`FlagReply(boardID, threadID, replyID, reason)`**: Flag reply for review
- **`FreezeReply(boardID, threadID, replyID)`** / **`UnfreezeReply(...)`**: Control reply mutability

### Member Management

- **`InviteMember(boardID, user, role)`**: Invite user with role (`boardID=0` for realm members, `boardID!=0` for board members)
- **`RequestInvite(boardID)`**: Request board membership (only for specific boards, not realm)
- **`AcceptInvite(boardID, user)`**: Accept pending invitation
- **`RevokeInvite(boardID, user)`**: Remove pending invitation
- **`ChangeMemberRole(boardID, member, role)`**: Update member role
- **`RemoveMember(boardID, member)`**: Remove member

### Moderation

- **`Ban(boardID, user, hours, reason)`**: Temporarily ban user
- **`Unban(boardID, user, reason)`**: Remove user ban

## Usage Examples

### Owner Walkthrough

1. Create board: `CreateBoard("governance", true)`
2. Set moderation threshold: `SetFlaggingThreshold(boardID, 3)`
3. Invite initial team: `InviteMember(boardID, adminAddr, "admin")`
4. Lock realm for stability: `LockRealm(false)` (content only initially)
5. Later lock members: `LockRealm(true)` (full lock)

### Moderator Walkthrough

1. Review pending invites in UI or via events
2. Accept worthy requests: `AcceptInvite(boardID, userAddr)`
3. Promote active contributors: `ChangeMemberRole(boardID, userAddr, "moderator")`
4. During incidents: `FlagThread(boardID, threadID, "harassment")`, `FreezeThread(boardID, threadID)`
5. Ban problematic users: `Ban(boardID, userAddr, 24, "spam")`

### Guest Walkthrough

1. Request access: `RequestInvite(boardID)`
2. Once accepted, participate: `CreateThread(boardID, "Hello", "Introduction post")`
3. Engage: `CreateReply(boardID, threadID, 0, "Welcome message")`
4. Build reputation for potential promotion

## Setup and Deployment

This realm is deployed like any Gno.land realm package. Boards logic is already deployed; developers and users can create and use boards against the current public API. There are no runtime migrations described here.


Future versions: If a new version is published under a new path (for example, `.../v2`), it will be documented with clear notes on compatibility and behavior changes. Existing boards under the current version will continue to function as-is on their deployed path.

Beta status: Gno.land is currently in beta; certain behaviors and interfaces may evolve. Where capabilities are “to be determined,” this README documents the current behavior and will be updated as standards are finalized.

**Running in a Gno.land environment**:
- Deploy the realm package to a path like `r/gnoland/boards2/v1`
- Default realm owners are hardcoded: `g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq` (@devx) and `g1manfred47kzduec920z88wfr64ylksmdcedlf5` (@moul)
- Adapt `initRealmPermissions` in `boards.gno:init()` for different deployments
- Interact via gnoweb or by submitting `Call` transactions to public functions

**Local development tips**:
- Use the file-tests in this directory (see `z_*_filetest.gno`) to understand invariants and expected behaviors
- Explore `docs/users/example-boards.md` for general board usage patterns

## Operational Runbook

Use this when you're on-call or responding to incidents. It strings together common actions (adjust threshold, freeze, ban, unlock paths) into safe sequences, acknowledging constraints from Roles, Moderation, and Governance.

### Routine Operations

- **Adjust flagging threshold**: `SetFlaggingThreshold(boardID, n)`; audit with `GetFlaggingThreshold`
- **Freeze during incidents**: `FreezeBoard` then unfreeze once resolved; for targeted control, freeze threads/replies
- **Ban cycle**: temporary bans using `Ban(..., hours, reason)`; review and `Unban` when appropriate
- **Membership hygiene**: handle `RequestInvite` queue; promote/demote with `ChangeMemberRole`

### Critical Actions

- **Lock realm** (irreversible): `LockRealm(true)` to freeze both content and membership; `LockRealm(false)` to preserve ability to modify members post-lock
- **Owner intervention**: realm owners can still flag content in frozen boards and hide with one flag

## Security and Best Practices

Validators help prevent privilege escalation; reply-edit limits reduce impersonation risk; freezing and bans are targeted controls. Favor least privilege and explicit roles, and prefer configuration (permissions) over code changes for routine policy updates.

- Realm locking is irreversible. Plan governance and migration paths before invoking `LockRealm`
- Board freezing prevents mutations; realm owners retain certain emergency capabilities (e.g., flagging) to mitigate abuse
- Validators enforce name ownership and restrict privilege escalation (e.g., owners inviting owners only)
- Bans apply only to guests/external users; use role changes to demote inappropriate staff before banning if needed
- Threshold-based flagging can be tuned per board; set sane defaults to avoid abuse
- Permissions are membership-gated; ensure DAO membership maintenance remains consistent with your governance process

## Troubleshooting and FAQ

Start here when something "should work but doesn't." Each answer calls back to the controlling dimension—permissions, moderation, or governance—so you can quickly identify which lever to adjust.

**Q: I get "unauthorized" when calling a function.**
- **DAO membership**: `BasicPermissions.WithPermission` requires BOTH permission AND DAO membership (`!bp.dao.Members().Has(user)`)
- **Permission scope**: Ensure you're checking the right scope (realm `gPerms` vs board `board.perms`)
- **Role mapping**: Verify the role has the permission in the specific scope
- **Lock state**: Check if realm/board is locked or frozen
- **Validator rejection**: Validators can reject even with correct permissions (e.g., non-owners inviting owners, username conflicts)

**Q: Why can't I use a username as a board name?**
- Board name validation calls `users.ResolveName(name)` and checks `user.Addr() != std.PreviousRealm().Address()`
- If the name matches a registered user, only that user can create a board with that name
- This prevents impersonation and namespace conflicts

**Q: Reply edit fails even though I'm moderator.**
- Replies are creator-edit only by design

**Q: Why is my reply hidden after flags?**
- It reached the board's flagging threshold. Moderators/Admins/Owners can review and adjust thresholds
- Check whether hide was triggered by realm owner single-flag

**Q: Board operations fail with "realm is locked"**
- Check `IsRealmLocked()` - realm locking prevents most mutations
- Only a limited set of operations (for example, owner flagging) work in locked realms

**Q: Member operations fail unexpectedly**
- Check `AreRealmMembersLocked()` - member locking prevents invitation/role changes
- Verify the caller has appropriate permissions for the specific operation

## Extensibility

Boards is intended to be composed and extended. Swap the permission engine, add validators, or layer additional UI controls—without destabilizing core flows. New capabilities generally arrive as new functions/permissions.

- Swap the permission engine at runtime (`SetPermissions`) to integrate custom role models or external ACLs
- Register additional validators for new permissions with `ValidateFunc`
- Extend rendering to surface custom moderation controls or analytics

## Versioning and Contributing

Versioning: [TO ADD — determined post‑launch].
