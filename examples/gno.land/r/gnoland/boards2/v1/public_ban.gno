package boards2

import (
	"std"
	"strings"
	"time"

	"gno.land/p/nt/avl"
)

// Constants for different banning periods.
const (
	BanDay  = uint(24)
	BanWeek = BanDay * 7
	BanYear = BanDay * 365
)

// Ban bans a user from a board for a period of time.
// Only invited guest members and external users can be banned.
// Banning board owners, admins and moderators is not allowed.
func Ban(_ realm, boardID BoardID, user std.Address, hours uint, reason string) {
	assertAddressIsValid(user)

	if hours == 0 {
		panic("ban period in hours is required")
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		panic("ban reason is required")
	}

	board := mustGetBoard(boardID)
	caller := std.PreviousRealm().Address()
	until := time.Now().Add(time.Minute * 60 * time.Duration(hours))
	args := Args{boardID, user, until, reason}
	board.perms.WithPermission(cross, caller, PermissionUserBan, args, func(realm, Args) {
		// When banning invited members make sure they are guests, otherwise
		// disallow banning. Only guest or external users can be banned.
		board := mustGetBoard(boardID)
		if board.perms.HasUser(user) && !board.perms.HasRole(user, RoleGuest) {
			panic("owner, admin and moderator banning is not allowed")
		}

		banned, found := getBannedUsers(boardID)
		if !found {
			banned = avl.NewTree()
			gBannedUsers.Set(boardID.Key(), banned)
		}

		banned.Set(user.String(), until)

		std.Emit(
			"UserBanned",
			"bannedBy", caller.String(),
			"boardID", boardID.String(),
			"user", user.String(),
			"until", until.Format(time.RFC3339),
			"reason", reason,
		)
	})
}

// Unban unbans a user from a board.
func Unban(_ realm, boardID BoardID, user std.Address, reason string) {
	assertAddressIsValid(user)

	board := mustGetBoard(boardID)
	caller := std.PreviousRealm().Address()
	args := Args{boardID, user, reason}
	board.perms.WithPermission(cross, caller, PermissionUserUnban, args, func(realm, Args) {
		banned, found := getBannedUsers(boardID)
		if !found || !banned.Has(user.String()) {
			panic("user is not banned")
		}

		banned.Remove(user.String())

		std.Emit(
			"UserUnbanned",
			"bannedBy", caller.String(),
			"boardID", boardID.String(),
			"user", user.String(),
			"reason", reason,
		)
	})
}

// IsBanned checks if a user is banned from a board.
func IsBanned(boardID BoardID, user std.Address) bool {
	banned, found := getBannedUsers(boardID)
	return found && banned.Has(user.String())
}

func assertAddressIsValid(addr std.Address) {
	if !addr.IsValid() {
		panic("invalid address: " + addr.String())
	}
}

func assertUserIsNotBanned(boardID BoardID, user std.Address) {
	banned, found := getBannedUsers(boardID)
	if !found {
		return
	}

	v, found := banned.Get(user.String())
	if !found {
		return
	}

	until := v.(time.Time)
	if time.Now().Before(until) {
		panic(user.String() + " is banned until " + until.Format(dateFormat))
	}
}
