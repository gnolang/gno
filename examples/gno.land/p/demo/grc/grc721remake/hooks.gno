package grc721remake

import (
	"std"
)

// HookType represents the type of operation a hook is associated with
type HookType string

const (
	HookTransfer       HookType = "transfer"
	HookApproval       HookType = "approval"
	HookApprovalForAll HookType = "approvalForAll"
	HookMint           HookType = "mint"
	HookBurn           HookType = "burn"
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
	OnTransfer(from, to std.Address, tid TokenID) error
}

// ApprovalHook is implemented by hooks that want to intercept approval operations
type ApprovalHook interface {
	Hook
	OnApproval(owner, approved std.Address, tid TokenID) error
}

// ApprovalForAllHook is implemented by hooks that want to intercept approvalForAll operations
type ApprovalForAllHook interface {
	Hook
	OnApprovalForAll(owner, operator std.Address, approved bool) error
}

// MintHook is implemented by hooks that want to intercept mint operations
type MintHook interface {
	Hook
	OnMint(to std.Address, tid TokenID) error
}

// BurnHook is implemented by hooks that want to intercept burn operations
type BurnHook interface {
	Hook
	OnBurn(tid TokenID) error
}

// HookRegistry manages a collection of hooks
type HookRegistry struct {
	hooks []Hook
}

// NewHookRegistry creates a new hook registry
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make([]Hook, 0),
	}
}

// RegisterHook adds a hook to the registry
func (reg *HookRegistry) RegisterHook(hook Hook) {
	reg.hooks = append(reg.hooks, hook)
}

// TriggerTransferHooks executes all transfer hooks of the specified time
func (reg *HookRegistry) TriggerTransferHooks(time HookTime, from, to std.Address, tid TokenID) error {
	for _, hook := range reg.hooks {
		if hook.GetHookType() == HookTransfer && hook.GetHookTime() == time {
			if transferHook, ok := hook.(TransferHook); ok {
				if err := transferHook.OnTransfer(from, to, tid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// TriggerApprovalHooks executes all approval hooks of the specified time
func (reg *HookRegistry) TriggerApprovalHooks(time HookTime, owner, approved std.Address, tid TokenID) error {
	for _, hook := range reg.hooks {
		if hook.GetHookType() == HookApproval && hook.GetHookTime() == time {
			if approvalHook, ok := hook.(ApprovalHook); ok {
				if err := approvalHook.OnApproval(owner, approved, tid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// TriggerApprovalForAllHooks executes all approvalForAll hooks of the specified time
func (reg *HookRegistry) TriggerApprovalForAllHooks(time HookTime, owner, operator std.Address, approved bool) error {
	for _, hook := range reg.hooks {
		if hook.GetHookType() == HookApprovalForAll && hook.GetHookTime() == time {
			if approvalForAllHook, ok := hook.(ApprovalForAllHook); ok {
				if err := approvalForAllHook.OnApprovalForAll(owner, operator, approved); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// TriggerMintHooks executes all mint hooks of the specified time
func (reg *HookRegistry) TriggerMintHooks(time HookTime, to std.Address, tid TokenID) error {
	for _, hook := range reg.hooks {
		if hook.GetHookType() == HookMint && hook.GetHookTime() == time {
			if mintHook, ok := hook.(MintHook); ok {
				if err := mintHook.OnMint(to, tid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// TriggerBurnHooks executes all burn hooks of the specified time
func (reg *HookRegistry) TriggerBurnHooks(time HookTime, tid TokenID) error {
	for _, hook := range reg.hooks {
		if hook.GetHookType() == HookBurn && hook.GetHookTime() == time {
			if burnHook, ok := hook.(BurnHook); ok {
				if err := burnHook.OnBurn(tid); err != nil {
					return err
				}
			}
		}
	}
	return nil
} 
