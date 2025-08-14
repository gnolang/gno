package pausablenft

import (
	"gno.land/p/demo/grc/grc721remake"
	"gno.land/p/demo/ufmt"
)

// Hook Definition
type PausableHook struct {
	paused bool
}

func (h *PausableHook) GetHookType() grc721remake.HookType {
	return grc721remake.HookUniversal
}

func (h *PausableHook) GetHookTime() grc721remake.HookTime {
	return grc721remake.HookBefore
}

func (h *PausableHook) OnAny(nft *grc721remake.NFT, operation string) error {
	if h.paused {
		return ufmt.Errorf("all operations are paused")
	}
	return nil
}
