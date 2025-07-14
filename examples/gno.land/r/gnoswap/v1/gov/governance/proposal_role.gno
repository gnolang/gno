package governance

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/ufmt"

	prbac "gno.land/p/gnoswap/rbac"
	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/rbac"
)

// CheckerType defines available predefined permission checker types
type CheckerType uint8

const (
	_ CheckerType = iota
	SingleAddressChecker
	MultiAddressChecker
	AlwaysAllowChecker
	AlwaysDenyChecker
)

func checkerTypeFromString(s string) CheckerType {
	switch strings.ToLower(s) {
	case "singleaddresschecker":
		return SingleAddressChecker
	case "multiaddresschecker":
		return MultiAddressChecker
	case "alwaysallowchecker":
		return AlwaysAllowChecker
	case "alwaysdenychecker":
		return AlwaysDenyChecker
	default:
		panic("unknown checker type: " + s)
	}
}

func (c CheckerType) String() string {
	switch c {
	case SingleAddressChecker:
		return "SingleAddressChecker"
	case MultiAddressChecker:
		return "MultiAddressChecker"
	case AlwaysAllowChecker:
		return "AlwaysAllowChecker"
	case AlwaysDenyChecker:
		return "AlwaysDenyChecker"
	default:
		return "Unknown"
	}
}

// RoleChangeInfo holds information about role changes
type RoleChangeInfo struct {
	RoleName       string
	PermissionName string
	Action         RoleChangeAction
	CheckerType    CheckerType
	Addresses      []std.Address // Used for `SingleAddressChecker` and `MultiAddressChecker`
	TokenPath      string        // Used for `TokenHolderChecker`
	MinAmount      uint64        // Used for `TokenHolderChecker`
}

type RoleChangeAction uint8

const (
	_ RoleChangeAction = iota
	AddRole
	RemoveRole
	AddPermission
	RemovePermission
	UpdatePermission
)

func RoleChangeActionFromString(s string) RoleChangeAction {
	switch strings.ToLower(s) {
	case "addrole":
		return AddRole
	case "removerole":
		return RemoveRole
	case "addpermission":
		return AddPermission
	case "removepermission":
		return RemovePermission
	case "updatepermission":
		return UpdatePermission
	default:
		panic("unknown role change action: " + s)
	}
}

func (r RoleChangeAction) String() string {
	switch r {
	case AddRole:
		return "AddRole"
	case RemoveRole:
		return "RemoveRole"
	case AddPermission:
		return "AddPermission"
	case RemovePermission:
		return "RemovePermission"
	case UpdatePermission:
		return "UpdatePermission"
	default:
		return "Unknown"
	}
}

func handleRoleChangeProposal(
	cur realm,
	title string,
	desc string,
	proposer std.Address,
	execution string,
) uint64 {
	roleChangeData := strings.TrimPrefix(execution, ROLE_PREFIX)

	parts := strings.Split(roleChangeData, ":")
	if len(parts) < 5 {
		panic(addDetailToError(
			errInvalidInput,
			ufmt.Sprintf("invalid role change data: %s", roleChangeData),
		))
	}

	roleName := parts[0]
	permissionName := parts[1]
	action := RoleChangeActionFromString(parts[2])
	checkerType := checkerTypeFromString(parts[3])
	addr := ""
	if len(parts) > 4 {
		addr = parts[4]
	}

	tokenPath := ""
	if len(parts) > 5 {
		tokenPath = parts[5]
	}

	minAmount := uint64(0)
	if len(parts) > 6 {
		minAmount = parseNumber(parts[6], kindUint64).(uint64)
	}

	addresses := parseAddresses(addr)

	// Use state transition to create proposal
	transition := CreateProposalTransition(proposer, RoleChange, title, desc)
	if err := applyTransition(transition); err != nil {
		panic(err)
	}

	// Get the new proposal ID from the updated state
	newProposalId := governanceState.GetProposalCounter()
	proposal, exists := governanceState.GetProposal(newProposalId)
	if !exists {
		panic(ufmt.Sprintf("proposal(%d) not found", newProposalId))
	}

	// Set role change info
	proposal.RoleChange = RoleChangeInfo{
		RoleName:       roleName,
		PermissionName: permissionName,
		Action:         action,
		CheckerType:    checkerType,
		Addresses:      addresses,
		TokenPath:      tokenPath,
		MinAmount:      minAmount,
	}

	// Update the proposal in state with role change info
	governanceState = governanceState.WithUpdatedProposal(newProposalId, func(p ProposalInfo) ProposalInfo {
		return proposal
	})

	// Proposal already stored in state

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"ProposeRoleChange",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"title", title,
		"description", desc,
		"roleName", roleName,
		"permissionName", permissionName,
		"action", action.String(),
		"checkerType", checkerType.String(),
		"proposalId", formatUint(newProposalId),
		"quorumAmount", formatUint(proposal.QuorumAmount),
		"maxVotingWeight", proposal.MaxVotingWeight.ToString(),
		"configVersion", formatUint(proposal.ConfigVersion),
		"createdAt", formatUint(proposal.State.CreatedAt),
	)

	return newProposalId
}

func getPermissionChecker(info RoleChangeInfo) prbac.PermissionChecker {
	switch info.CheckerType {
	case SingleAddressChecker:
		if len(info.Addresses) != 1 {
			panic("SingleAddressChecker requires exactly one address")
		}
		return func(caller std.Address) error {
			if caller == info.Addresses[0] {
				return nil
			}
			return errors.New("unauthorized: not the designated address")
		}
	case MultiAddressChecker:
		if len(info.Addresses) == 0 {
			panic("MultiAddressChecker requires at least one address")
		}
		return func(caller std.Address) error {
			for _, addr := range info.Addresses {
				if caller == addr {
					return nil
				}
			}
			return errors.New("unauthorized: not in the list of authorized addresses")
		}
	case AlwaysAllowChecker:
		return func(std.Address) error {
			return nil
		}
	case AlwaysDenyChecker:
		return func(std.Address) error {
			return errors.New("unauthorized: access denied")
		}
	default:
		panic("unknown checker type")
	}
}

// parseAddresses parses a colon-separated string of addresses into an array of std.Address.
// Example input: "addr1:addr2:addr3"
func parseAddresses(addressStr string) []std.Address {
	if addressStr == "" {
		return []std.Address{}
	}

	addrStrings := strings.Split(addressStr, ":")
	addresses := make([]std.Address, 0, len(addrStrings))

	for _, addrStr := range addrStrings {
		addrStr = strings.TrimSpace(addrStr)
		if addrStr == "" {
			continue
		}

		addr := std.Address(addrStr)
		addresses = append(addresses, addr)
	}

	return addresses
}

/////////// EXECUTE PROPOSAL ///////////

func executeAddRole(cur realm, roleInfo RoleChangeInfo) error {
	if roleInfo.CheckerType == SingleAddressChecker && len(roleInfo.Addresses) == 1 {
		return access.CreateRole(cross, roleInfo.RoleName, roleInfo.Addresses[0])
	}

	return errors.New("invalid checker type or addresses for role creation")
}

func executeRemoveRole(cur realm, info RoleChangeInfo) error {
	panic("TODO")
}

func executeAddPermission(cur realm, roleInfo RoleChangeInfo) error {
	if !access.RoleExists(roleInfo.RoleName) {
		return ufmt.Errorf("role %s does not exist", roleInfo.RoleName)
	}

	// Create and add the permission checker
	checker := getPermissionChecker(roleInfo)
	return rbac.RegisterPermission(cross, roleInfo.RoleName, roleInfo.PermissionName, checker)
}

func executeRemovePermission(cur realm, roleInfo RoleChangeInfo) error {
	if !access.RoleExists(roleInfo.RoleName) {
		return ufmt.Errorf("role %s does not exist", roleInfo.RoleName)
	}

	return rbac.RemovePermission(cross, roleInfo.RoleName, roleInfo.PermissionName)
}

func executeUpdatePermission(cur realm, roleInfo RoleChangeInfo) error {
	if !access.RoleExists(roleInfo.RoleName) {
		return ufmt.Errorf("role %s does not exist", roleInfo.RoleName)
	}

	// Update the permission checker
	checker := getPermissionChecker(roleInfo)
	return rbac.UpdatePermission(cross, roleInfo.RoleName, roleInfo.PermissionName, checker)
}
