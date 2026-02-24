package validators

import (
	"chain/runtime"
	"strings"

	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
)

func assertCallerIsAllowed() {
	callerPath := runtime.PreviousRealm().PkgPath()
	if callerPath == "" {
		panic("caller must be a realm, not a user")
	}

	if !IsAllowedRealm(callerPath) {
		panic("realm " + callerPath + " is not allowed to modify the validator set")
	}
}

// AddValidator is a public method that allows whitelisted realms to add
// a validator to the set.
func AddValidator(cur realm, validator validators.Validator) {
	assertCallerIsAllowed()
	addValidator(validator)
}

// RemoveValidator is a public method that allows whitelisted realms to remove
// a validator from the set.
func RemoveValidator(cur realm, addr address) {
	assertCallerIsAllowed()
	removeValidator(addr)
}

// IsAllowedRealm returns true if the given realm path is in the whitelist.
func IsAllowedRealm(pkgPath string) bool {
	return allowedRealms.Has(pkgPath)
}

// GetAllowedRealms returns the list of realm paths that are allowed to modify the validator set.
func GetAllowedRealms() []string {
	realms := make([]string, 0, allowedRealms.Size())
	allowedRealms.Iterate("", "", func(key string, _ any) bool {
		realms = append(realms, key)
		return false
	})
	return realms
}

// NewPropAllowedRealmUpdateRequest creates a new proposal request that updates
// the set of realms allowed to modify the validator set.
func NewPropAllowedRealmUpdateRequest(add []string, remove []string, title, description string) dao.ProposalRequest {
	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}

	if len(add) == 0 && len(remove) == 0 {
		panic("proposal must add or remove at least one realm")
	}

	var desc strings.Builder
	desc.WriteString(description)
	if len(description) > 0 {
		desc.WriteString("\n\n")
	}

	desc.WriteString("## Allowed Realm Updates\n")
	for _, r := range add {
		desc.WriteString("- add: " + r + "\n")
	}
	for _, r := range remove {
		desc.WriteString("- remove: " + r + "\n")
	}

	callback := func(cur realm) error {
		for _, r := range add {
			allowedRealms.Set(r, true)
		}
		for _, r := range remove {
			allowedRealms.Remove(r)
		}
		return nil
	}

	e := dao.NewSimpleExecutor(callback, "")
	return dao.NewProposalRequest(title, desc.String(), e)
}
