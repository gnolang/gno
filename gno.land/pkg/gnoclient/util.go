package gnoclient

func (cfg BaseTxCfg) validateBaseTxConfig() error {
	if cfg.GasWanted <= 0 {
		return ErrInvalidGasWanted
	}
	if cfg.GasFee == "" {
		return ErrInvalidGasFee
	}

	return nil
}
