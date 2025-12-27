package grc721remake

import (
	"std"

	"gno.land/p/demo/avl"
)

// HookType represents the type of operation a hook is associated with
type HookType string

const (
	HookTransfer       HookType = "transfer"
	HookApproval       HookType = "approval"
	HookApprovalForAll HookType = "approvalForAll"
	HookMint           HookType = "mint"
	HookBurn           HookType = "burn"
	HookUniversal      HookType = "universal"
)

// HookTime represents when a hook should be executed
type HookTime string

const (
	HookBefore HookTime = "before"
	HookAfter  HookTime = "after"
)

// Hook is the base interface that all hooks must implement
type Hook interface {
	GetHookType() HookType
	GetHookTime() HookTime
}

// TransferHook is implemented by hooks that want to intercept transfer operations
type TransferHook interface {
	Hook
	OnTransfer(nft *NFT, from, to std.Address, tid TokenID) error
}

// ApprovalHook is implemented by hooks that want to intercept approval operations
type ApprovalHook interface {
	Hook
	OnApproval(nft *NFT, owner, approved std.Address, tid TokenID) error
}

// ApprovalForAllHook is implemented by hooks that want to intercept approvalForAll operations
type ApprovalForAllHook interface {
	Hook
	OnApprovalForAll(nft *NFT, owner, operator std.Address, approved bool) error
}

// MintHook is implemented by hooks that want to intercept mint operations
type MintHook interface {
	Hook
	OnMint(nft *NFT, to std.Address, tid TokenID) error
}

// BurnHook is implemented by hooks that want to intercept burn operations
type BurnHook interface {
	Hook
	OnBurn(nft *NFT, tid TokenID) error
}

// UniversalHook is implemented by hooks that want to intercept all operations
type UniversalHook interface {
	Hook
	OnAny(nft *NFT, operation string) error
}

// HookRegistry manages a collection of hooks using a tree structure
type HookRegistry struct {
	hooks *avl.Tree
}

// NewHookRegistry creates a new hook registry
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: avl.NewTree(),
	}
}

// RegisterHookWithKey adds a hook to the registry with a specified key
func (reg *HookRegistry) RegisterHook(key string, hook Hook) {
	reg.hooks.Set(key, hook)
}

// GetHook retrieves a hook by its key
func (reg *HookRegistry) GetHook(key string) (Hook, bool) {
	value, exists := reg.hooks.Get(key)
	if !exists {
		return nil, false
	}
	return value.(Hook), true
}

// TriggerTransferHooks executes all transfer hooks of the specified time
func (reg *HookRegistry) TriggerTransferHooks(nft *NFT, time HookTime, from, to std.Address, tid TokenID) error {
	var foundError error = nil

	reg.hooks.Iterate("", "", func(key string, value any) bool {
		hook := value.(Hook)
		if hook.GetHookTime() != time {
			return false
		}

		if hook.GetHookType() == HookTransfer {
			if transferHook, ok := hook.(TransferHook); ok {
				if err := transferHook.OnTransfer(nft, from, to, tid); err != nil {
					foundError = err
					return true
				}
			}
		} else if hook.GetHookType() == HookUniversal {
			if universalHook, ok := hook.(UniversalHook); ok {
				if err := universalHook.OnAny(nft, "transfer"); err != nil {
					foundError = err
					return true
				}
			}
		}
		return false
	})

	return foundError
}

// TriggerApprovalHooks executes all approval hooks of the specified time
func (reg *HookRegistry) TriggerApprovalHooks(nft *NFT, time HookTime, owner, approved std.Address, tid TokenID) error {
	var foundError error = nil

	reg.hooks.Iterate("", "", func(key string, value any) bool {
		hook := value.(Hook)
		if hook.GetHookTime() != time {
			return false
		}

		if hook.GetHookType() == HookApproval {
			if approvalHook, ok := hook.(ApprovalHook); ok {
				if err := approvalHook.OnApproval(nft, owner, approved, tid); err != nil {
					foundError = err
					return true
				}
			}
		} else if hook.GetHookType() == HookUniversal {
			if universalHook, ok := hook.(UniversalHook); ok {
				if err := universalHook.OnAny(nft, "approval"); err != nil {
					foundError = err
					return true
				}
			}
		}
		return false
	})

	return foundError
}

// TriggerApprovalForAllHooks executes all approvalForAll hooks of the specified time
func (reg *HookRegistry) TriggerApprovalForAllHooks(nft *NFT, time HookTime, owner, operator std.Address, approved bool) error {
	var foundError error = nil

	reg.hooks.Iterate("", "", func(key string, value any) bool {
		hook := value.(Hook)
		if hook.GetHookTime() != time {
			return false
		}

		if hook.GetHookType() == HookApprovalForAll {
			if approvalForAllHook, ok := hook.(ApprovalForAllHook); ok {
				if err := approvalForAllHook.OnApprovalForAll(nft, owner, operator, approved); err != nil {
					foundError = err
					return true
				}
			}
		} else if hook.GetHookType() == HookUniversal {
			if universalHook, ok := hook.(UniversalHook); ok {
				if err := universalHook.OnAny(nft, "approvalForAll"); err != nil {
					foundError = err
					return true
				}
			}
		}
		return false
	})

	return foundError
}

// TriggerMintHooks executes all mint hooks of the specified time
func (reg *HookRegistry) TriggerMintHooks(nft *NFT, time HookTime, to std.Address, tid TokenID) error {
	var foundError error = nil

	reg.hooks.Iterate("", "", func(key string, value any) bool {
		hook := value.(Hook)
		if hook.GetHookTime() != time {
			return false
		}

		if hook.GetHookType() == HookMint {
			if mintHook, ok := hook.(MintHook); ok {
				if err := mintHook.OnMint(nft, to, tid); err != nil {
					foundError = err
					return true
				}
			}
		} else if hook.GetHookType() == HookUniversal {
			if universalHook, ok := hook.(UniversalHook); ok {
				if err := universalHook.OnAny(nft, "mint"); err != nil {
					foundError = err
					return true
				}
			}
		}
		return false
	})

	return foundError
}

// TriggerBurnHooks executes all burn hooks of the specified time
func (reg *HookRegistry) TriggerBurnHooks(nft *NFT, time HookTime, tid TokenID) error {
	var foundError error = nil

	reg.hooks.Iterate("", "", func(key string, value any) bool {
		hook := value.(Hook)
		if hook.GetHookTime() != time {
			return false
		}

		if hook.GetHookType() == HookBurn {
			if burnHook, ok := hook.(BurnHook); ok {
				if err := burnHook.OnBurn(nft, tid); err != nil {
					foundError = err
					return true
				}
			}
		} else if hook.GetHookType() == HookUniversal {
			if universalHook, ok := hook.(UniversalHook); ok {
				if err := universalHook.OnAny(nft, "burn"); err != nil {
					foundError = err
					return true
				}
			}
		}
		return false
	})

	return foundError
}
