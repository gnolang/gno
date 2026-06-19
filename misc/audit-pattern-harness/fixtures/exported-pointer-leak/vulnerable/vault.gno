package vault

type Vault struct {
	Balance int
}

var PublicVault *Vault = &Vault{Balance: 10}

func GetVault() *Vault {
	return PublicVault
}
