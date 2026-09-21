package registry

import "errors"

var (
	mainAddress   address
	backupAddress address

	ErrInvalidAddr  = errors.New("Ivan's registry: Invalid address")
	ErrUnauthorized = errors.New("Ivan's registry: Unauthorized")
)

func init() {
	mainAddress = "g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x"
	backupAddress = "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"
}

func MainAddress() address {
	return mainAddress
}

func BackupAddress() address {
	return backupAddress
}

func SetMainAddress(cur realm, addr address) error {
	assertAuthorized(0, cur)

	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	mainAddress = addr
	return nil
}

func SetBackupAddress(cur realm, addr address) error {
	assertAuthorized(0, cur)

	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	backupAddress = addr
	return nil
}

// assertAuthorized checks that the caller realm (rlm.Previous()) is the
// main or backup address. rlm must be the caller's live cur — the
// IsCurrent gate makes the helper safe-by-construction so a future
// non-crossing call site can't silently bypass auth by passing a
// stashed or sibling-frame realm value. See
// docs/resources/gno-security.md (Class-2 designation forgery).
func assertAuthorized(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic(ErrUnauthorized)
	}
	caller := rlm.Previous().Address()
	isAuthorized := caller == mainAddress || caller == backupAddress

	if !isAuthorized {
		panic(ErrUnauthorized)
	}
}
