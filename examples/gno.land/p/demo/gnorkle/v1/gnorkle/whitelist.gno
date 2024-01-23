package gnorkle

type Whitelist interface {
	ClearAddresses()
	AddAddresses(addresses []string)
	RemoveAddress(address string)
	HasDefinition() bool
	HasAddress(address string) bool
}

func (i *Instance) ClearWhitelist(feedID string) {
	if feedID == "" {
		i.whitelist.ClearAddresses()
		return
	}

	feedWithWhitelist := i.getFeedWithWhitelist(feedID)
	feedWithWhitelist.ClearAddresses()
}

func (i *Instance) AddToWhitelist(feedID string, addresses []string) {
	if feedID == "" {
		i.whitelist.AddAddresses(addresses)
		return
	}

	feedWithWhitelist := i.getFeedWithWhitelist(feedID)
	feedWithWhitelist.AddAddresses(addresses)
}

func (i *Instance) RemoveFromWhitelist(feedID string, address string) {
	if feedID == "" {
		i.whitelist.RemoveAddress(address)
		return
	}

	feedWithWhitelist := i.getFeedWithWhitelist(feedID)
	feedWithWhitelist.RemoveAddress(address)
}

// TODO: test this.

// addressWhiteListed returns true if:
// - the feed has a white list and the address is whitelisted, or
// - the feed has no white list and the instance has a white list and the address is whitelisted, or
// - the feed has no white list and the instance has no white list.
func addressIsWhitelisted(instanceWhitelist, feedWhitelist Whitelist, address string, instanceWhitelistedOverride *bool) bool {
	// A feed whitelist takes priority, so it will return false if the feed has a whitelist and the caller is
	// not a part of it. An empty whitelist defers to the instance whitelist.
	if feedWhitelist != nil {
		if feedWhitelist.HasDefinition() && !feedWhitelist.HasAddress(address) {
			return false
		}

		// Getting to this point means that one of the following is true:
		// - the feed has no defined whitelist (so it can't possibly have the address whitelisted)
		// - the feed has a defined whitelist and the caller is a part of it
		//
		// In this case, we can be sure that the boolean indicating whether the feed has this address whitelisted
		// is equivalent to the boolean indicating whether the feed has a defined whitelist.
		if feedWhitelist.HasDefinition() {
			return true
		}
	}

	if instanceWhitelistedOverride != nil {
		return *instanceWhitelistedOverride
	}

	// We were unable able to determine whether this address is allowed after looking at the feed whitelist,
	// so fall back to the instance whitelist. A complete absence of values in the instance whitelist means
	// that the instance has no whitelist so we can return true because everything is allowed by default.
	if instanceWhitelist == nil || !instanceWhitelist.HasDefinition() {
		return true
	}

	// The instance whitelist is defined so it the address is present then it is allowed.
	return instanceWhitelist.HasAddress(address)
}
