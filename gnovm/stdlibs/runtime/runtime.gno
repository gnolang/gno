// Package runtime allows access to the GnoVM's runtime information.
package runtime

func NewReadOnlyBanker() chain.Banker {
}

func NewLimitedSendBanker() chain.Banker {
}

func NewSendBanker() chain.Banker {
}

func NewIssueBanker() chain.Banker {
}

func AssertOriginCall() {
}

func CallerRealm() chain.Realm {
}

func CurrentRealm() chain.Realm {
}

func Deposits() chain.Coins {
}

func ChainID() string {
}

func Height() uint64 {
}

func Emit(eventType string, attrKeyValues ...string) {
}
