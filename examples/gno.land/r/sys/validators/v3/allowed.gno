package validators

import (
	"chain"
	"chain/runtime"
	"strings"

	"gno.land/p/nt/bptree/v0"
	"gno.land/r/gov/dao"
)

// allowedRealms holds the whitelist of realm paths allowed to modify the validator set.
var allowedRealms = bptree.NewBPTree32()

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
// a validator to the set. The operator must already be registered in
// valoperCache with KeepRunning=true.
func AddValidator(cur realm, change ValoperChange) {
	assertCallerIsAllowed()
	if change.Power == 0 {
		panic("AddValidator requires Power > 0; use RemoveValidator for removals")
	}
	exec := newValoperChangeExecutor([]ValoperChange{change})
	if err := exec.Execute(cross); err != nil {
		panic(err)
	}
}

// RemoveValidator is a public method that allows whitelisted realms to remove
// a validator from the set, looked up by operator address.
func RemoveValidator(cur realm, op address) {
	assertCallerIsAllowed()
	exec := newValoperChangeExecutor([]ValoperChange{{OperatorAddress: op, Power: 0}})
	if err := exec.Execute(cross); err != nil {
		panic(err)
	}
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
//
// Each path may appear AT MOST ONCE across add and remove combined;
// duplicates within either list, and any add ∩ remove overlap, are
// rejected at create-time.
func NewPropAllowedRealmUpdateRequest(add []string, remove []string, title, description string) dao.ProposalRequest {
	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}

	if len(add) == 0 && len(remove) == 0 {
		panic("proposal must add or remove at least one realm")
	}

	seen := make(map[string]bool, len(add)+len(remove))
	for _, r := range add {
		if seen[r] {
			panic("duplicate realm in proposal: " + r)
		}
		seen[r] = true
	}
	for _, r := range remove {
		if seen[r] {
			panic("duplicate realm in proposal: " + r)
		}
		seen[r] = true
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
			chain.Emit("AllowedRealmAdded", "realm", r)
		}
		for _, r := range remove {
			allowedRealms.Remove(r)
			chain.Emit("AllowedRealmRemoved", "realm", r)
		}
		return nil
	}

	e := dao.NewSimpleExecutor(callback, "")
	return dao.NewProposalRequest(title, desc.String(), e)
}
