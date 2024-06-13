package main

type homeDirectory struct {
	homeDir     string
	genesisFile string
}

func (h homeDirectory) Path() string       { return h.homeDir }
func (h homeDirectory) ConfigDir() string  { return h.Path() + "/config" }
func (h homeDirectory) ConfigFile() string { return h.ConfigDir() + "/config.toml" }

func (h homeDirectory) GenesisFilePath() string {
	if h.genesisFile != "" {
		return h.genesisFile
	}
	return h.Path() + "/genesis.json"
}

func (h homeDirectory) SecretsDir() string     { return h.Path() + "/secrets" }
func (h homeDirectory) SecretsNodeKey() string { return h.SecretsDir() + "/" + defaultNodeKeyName }
func (h homeDirectory) SecretsValidatorKey() string {
	return h.SecretsDir() + "/" + defaultValidatorKeyName
}
func (h homeDirectory) SecretsValidatorState() string {
	return h.SecretsDir() + "/" + defaultValidatorStateName
}
