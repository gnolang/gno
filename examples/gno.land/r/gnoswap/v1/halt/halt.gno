package halt

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/access"
)

var errHalted = errors.New("halted")

var govAddr, _ = access.GetAddress(access.ROLE_GOVERNANCE)

// OperationType defines the structure for operation configuration
type OperationType struct {
	Type        OpType
	Name        string
	Description string
	IsContract  bool
}

const (
	// Special mode for beta mainnet (no withdrawals)
	// Issue: https://github.com/gnoswap-labs/gnoswap/issues/517
	LvMainnetSafeMode LevelID = 10
)

var haltManager *Manager

const (
	enabled  = true
	disabled = false
)

func init() {
	// Create a default manager first
	haltManager = DefaultManager()

	// Create custom MainnetSafeMode level that disables withdrawals for beta mainnet
	operations := make(map[OpType]bool)
	operations[OpTypePool] = enabled
	operations[OpTypePosition] = enabled
	operations[OpTypeProtocolFee] = enabled
	operations[OpTypeRouter] = enabled
	operations[OpTypeStaker] = enabled
	operations[OpTypeLaunchpad] = enabled
	operations[OpTypeGovernance] = enabled
	operations[OpTypeGovStaker] = enabled
	operations[OpTypeXGns] = enabled
	operations[OpTypeEmission] = enabled
	operations[OpTypeWithdraw] = disabled

	mainnetSafeMode := NewHaltLevel(
		LvMainnetSafeMode,
		"MainnetSafeMode",
		"Safe mode for post-mainnet (no withdrawals)",
		operations,
	)

	haltManager.RegisterHaltLevel(mainnetSafeMode)

	// Set initial level to MainnetSafeMode for right after mainnet
	if err := haltManager.SetCurrentLevel(LvMainnetSafeMode); err != nil {
		panic(ufmt.Sprintf("Failed to set initial halt level: %v", err))
	}
}

// GetHalt returns the current halt status.
func GetHalt() bool {
	currentLevel := haltManager.Level().Level()
	return currentLevel == LvCompleteHalt
}

// IsHalted checks if the requested operations are allowed in the current halt level.
// If not allowed, it returns an error.
//
// Example Usage:
//
// For checking if only one operation is halted:
//
//	err := halt.IsHalted(halt.OpTypeSwap)
//
// For checking if multiple operations are halted:
//
//	err := halt.IsHalted(halt.OpTypeSwap, halt.OpTypeLiquidity /* and more */)
//
// Parameters:
// - opType (halt.OpType): Optional operation type to check. If not provided, checks if system is completely halted.
func IsHalted(opType ...OpType) error {
	// if no operation specified, check if system is completely halted
	if len(opType) == 0 {
		if GetHalt() {
			return errHalted
		}
		return nil
	}

	// check each operation(s)
	for _, oo := range opType {
		op := NewOperation(oo, string(oo), string(oo)+" operation")
		if !haltManager.Level().IsOperationAllowed(op) {
			return ufmt.Errorf("Operation '%s' is currently disabled in halt level: %s", oo, haltManager.Status(oo))
		}
	}

	return nil
}

// SetHaltByAdmin allows an admin to set the halt level of GnoSwap.
// Only an admin can execute this function.
//
// Parameters:
// - haltLevel (halt.LevelID): The new halt level to set.
//
// Returns:
// - error: nil if successful, otherwise an error.
func SetHaltLevelByAdmin(cur realm, haltLevel LevelID) error {
	caller := std.PreviousRealm().Address()
	if err := access.AdminOnly(caller); err != nil {
		return ufmt.Errorf("only admin can set halt level, called from %s", caller)
	}
	return setHaltLevel(haltLevel)
}

// SetHaltByAdmin maintains backward compatibility with the old API.
// It maps true to CompleteHalt and false to NoHalt.
//
// While the halt mechanism should fundamentally be controlled by governance,
// there needs to be functionality that allows admin intervention
// in special cases such as emergencies.
//
// Parameters:
// - halt (bool): true to halt, false to unhalt.
//
// Returns:
// - error: nil if successful, otherwise an error.
//
// TODO: Should we only allow EmergencyHalt or CompleteHalt?
func SetHaltByAdmin(cur realm, halt bool) error {
	caller := getPrevAddr()
	if err := access.AdminOnly(caller); err != nil {
		return ufmt.Errorf("only admin can set halt, called from %s", caller)
	}

	if halt {
		return setHaltLevel(LvCompleteHalt)
	}
	return setHaltLevel(LvNoHalt)
}

// SetHalt allows both admin and governance to set the halt status of GnoSwap.
// If no specific level is provided, it defaults to CompleteHalt when halting.
//
// Parameters:
// - halt (bool): true to halt, false to unhalt.
// - level (halt.LevelID): Optional halt level. Only NoHalt, EmergencyHalt and CompleteHalt are allowed.
//
// Returns:
// - error: nil if successful, otherwise an error.
func SetHalt(cur realm, halt bool, level ...LevelID) error {
	caller := getPrevAddr()
	adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
	govAddr, _ := access.GetAddress(access.ROLE_GOVERNANCE)
	if caller != adminAddr && caller != govAddr {
		return ufmt.Errorf("only admin or governance can set halt, called from %s", caller)
	}

	// If no level specified, use defaults
	if len(level) == 0 {
		if !halt {
			return setHaltLevel(LvNoHalt)
		}
		return setHaltLevel(LvCompleteHalt)
	}

	// Validate the specified level
	haltLevel := level[0]
	switch haltLevel {
	case LvNoHalt, LvEmergencyHalt, LvCompleteHalt:
		return setHaltLevel(haltLevel)
	default:
		return ufmt.Errorf("only NoHalt, EmergencyHalt and CompleteHalt levels are allowed")
	}
}

// SetHaltLevel allows the governance contract to set the halt level of GnoSwap.
// Only the governance contract can execute this function.
//
// Parameters:
// - haltLevel (halt.LevelID): The new halt level to set.
//
// Returns:
// - error: nil if successful, otherwise an error.
func SetHaltLevel(cur realm, haltLevel LevelID) error {
	caller := getPrevAddr()
	adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
	govAddr, _ := access.GetAddress(access.ROLE_GOVERNANCE)
	if caller == adminAddr || caller == govAddr {
		return setHaltLevel(haltLevel)
	}
	return ufmt.Errorf("only admin or governance can set halt level, called from %s", caller)
}

// setHaltLevel updates the halt level to the specified value.
// This is an internal function that should only be called by privileged functions.
//
// Parameters:
// - haltLevel (halt.LevelID): The new halt level to set.
//
// Returns:
// - error: nil if successful, otherwise an error.
func setHaltLevel(haltLevel LevelID) error {
	if err := haltManager.SetCurrentLevel(haltLevel); err != nil {
		return ufmt.Errorf("invalid halt level: %v", err)
	}

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"setHaltLevel",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"level", strconv.Itoa(int(haltLevel)),
		"levelName", haltManager.Level().Name(),
	)
	return nil
}

// SetOperationStatus allows governance to enable/disable specific operations.
// This provides granular control over protocol functions.
//
// Parameters:
// - opType (halt.OpType): The operation type to configure.
// - allowed (bool): Whether the operation should be allowed.
//
// Returns:
// - error: nil if successful, otherwise an error.
func SetOperationStatus(cur realm, opType OpType, allowed bool) error {
	caller := getPrevAddr()
	if err := access.GovernanceOnly(caller); err != nil {
		return ufmt.Errorf("only governance(%s) can set operation status, called from %s", govAddr, caller)
	}
	return setOperationStatus(opType, allowed)
}

// SetOperationStatusByAdmin allows an admin to enable/disable specific operations.
//
// Parameters:
// - opType (halt.OpType): The operation type to configure.
// - allowed (bool): Whether the operation should be allowed.
//
// Returns:
// - error: nil if successful, otherwise an error.
func SetOperationStatusByAdmin(cur realm, opType OpType, allowed bool) error {
	caller := getPrevAddr()
	if err := access.AdminOnly(caller); err != nil {
		return ufmt.Errorf("only admin can set operation status, called from %s", caller)
	}

	return setOperationStatus(opType, allowed)
}

// setOperationStatus updates whether a specific operation is allowed.
// This is an internal function that should only be called by privileged functions.
//
// Parameters:
// - opType (halt.OpType): The operation type to configure.
// - allowed (bool): Whether the operation should be allowed.
//
// Returns:
// - error: nil if successful, otherwise an error.
func setOperationStatus(opType OpType, allowed bool) error {
	if err := haltManager.SetOperationStatus(opType, allowed); err != nil {
		return ufmt.Errorf("failed to set operation status: %v", err)
	}

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"setOperationStatus",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"operation", string(opType),
		"allowed", strconv.FormatBool(allowed),
	)
	return nil
}

// EnableWithdrawalsForTestnet enables withdrawals in MainnetSafeMode.
// This should be called after ensuring stability on beta mainnet.
//
// Returns:
// - error: nil if successful, otherwise an error.
func EnableWithdrawalsForTestnet(cur realm) error {
	caller := getPrevAddr() // TODO: no need to check governance?
	if err := access.AdminOnly(caller); err != nil {
		return ufmt.Errorf("only admin can enable withdrawals, called from %s", caller)
	}

	// Only proceed if we're in MainnetSafeMode
	if haltManager.Level().Level() != LvMainnetSafeMode {
		return nil
	}

	err := setOperationStatus(OpTypeWithdraw, true)
	if err != nil {
		return err
	}

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"enableWithdrawals",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
	)

	return nil
}

// GetHaltManager returns the halt manager instance.
// This is useful for external packages that need to interact directly with the manager.
//
// Returns:
// - *halt.Manager: The halt manager instance.
func GetHaltManager() *Manager {
	return haltManager
}

// GetCurrentHaltLevel returns the current halt level.
//
// Returns:
// - halt.LevelID: The current halt level ID.
func GetCurrentHaltLevel() LevelID {
	return haltManager.Level().Level()
}

// EnableContractOperations allows enabling contract-specific operations.
// Only an admin can execute this function.
//
// Parameters:
// - contract (halt.OpType): The contract operation type to enable.
//
// Returns:
// - error: nil if successful, otherwise an error.
func EnableContractOperations(cur realm, contract OpType) error {
	caller := getPrevAddr()
	if err := access.AdminOnly(caller); err != nil {
		return ufmt.Errorf("only admin can enable contract operations, called from %s", caller)
	}

	// Only proceed if we're in ContractSafeMode
	if haltManager.Level().Level() != LvContractHalt {
		return nil
	}

	err := setOperationStatus(contract, true)
	if err != nil {
		return err
	}

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"enableContract",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"contract", string(contract),
	)

	return nil
}

// DisableContractOperations allows disabling contract-specific operations.
// Only an admin can execute this function.
//
// Parameters:
// - contract (halt.OpType): The contract operation type to disable.
//
// Returns:
func DisableContractOperations(cur realm, contract OpType) error {
	caller := getPrevAddr()
	if err := access.AdminOnly(caller); err != nil {
		return ufmt.Errorf("only admin can disable contract operations, called from %s", caller)
	}

	err := setOperationStatus(contract, false)
	if err != nil {
		return err
	}

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"disableContract",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"contract", string(contract),
	)

	return nil
}

// IsContractHalted checks if specific contract operations are halted
func IsContractHalted(contract OpType) bool {
	// if contract is halted, IsHalted will return an error
	// so we need to check if the error is not nil
	err := IsHalted(contract)
	return err != nil
}

// GetContractStatus returns the current status for a specific contract
func GetContractStatus(contract OpType) bool {
	return haltManager.Level().IsOperationAllowed(NewOperation(contract, string(contract), ""))
}
