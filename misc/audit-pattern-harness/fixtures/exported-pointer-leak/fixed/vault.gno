package vault

type Vault struct {
	Balance int
}

var privateVault = Vault{Balance: 10}

// Fixed: return a value copy, not a pointer alias to package state.
func GetVault() Vault {
	return privateVault
}
