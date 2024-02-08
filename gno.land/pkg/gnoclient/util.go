package gnoclient

func (cfg BaseTxCfg) validateBaseTxConfig() error {
	if cfg.GasWanted < 0 {
		return ErrInvalidGasWanted
	}
	if cfg.GasFee < "" {
		return ErrInvalidGasFee
	}

	return nil
}

func (msg MsgCall) validateMsgCall() error {
	if msg.PkgPath == "" {
		return ErrEmptyPkgPath
	}
	if msg.FuncName == "" {
		return ErrEmptyFuncName
	}
	return nil
}
