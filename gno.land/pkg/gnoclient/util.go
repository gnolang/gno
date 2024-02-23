package gnoclient

import "github.com/gnolang/gno/tm2/pkg/std"

func (cfg BaseTxCfg) validateBaseTxConfig() error {
	if cfg.GasWanted <= 0 {
		return ErrInvalidGasWanted
	}
	if cfg.GasFee == "" {
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

func (msg MsgSend) validateMsgSend() error {
	if msg.ToAddress.IsZero() {
		return ErrInvalidToAddress
	}
	_, err := std.ParseCoins(msg.Send)
	if err != nil {
		return ErrInvalidSendAmount
	}
	return nil
}
