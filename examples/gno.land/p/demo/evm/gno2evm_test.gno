package evm

import "testing"

func TestGno2EvmMyAddressOK(t *testing.T) {
	_, err := GnoAddressToUint256("g18dgpzqxnhmrsemnadngg403f8clf5l82jalxza")
	if err != nil {
		t.Error(err)
	}
}

func TestGno2EvmMyAddress20Bytes(t *testing.T) {
	num, _ := GnoAddressToUint256("g18dgpzqxnhmrsemnadngg403f8clf5l82jalxza")
	b := num.Bytes()
	l := len(b)
	if l != 20 {
		t.Errorf("expected 20 bytes, got %d", l)
	}
}