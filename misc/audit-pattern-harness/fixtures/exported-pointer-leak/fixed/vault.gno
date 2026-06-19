package vault

type Vault struct {
	Balance int
}

var privateVault = Vault{Balance: 10}

func GetVault() Vault {
	return privateVault
}
