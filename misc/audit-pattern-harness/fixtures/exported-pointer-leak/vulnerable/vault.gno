package vault

type Vault struct {
	Balance int
}

// Vulnerable: exported pointer gives callers a direct alias to mutable state.
var PublicVault *Vault = &Vault{Balance: 10}

// Vulnerable: exported pointer getter leaks the same package-level object.
func GetVault() *Vault {
	return PublicVault
}
