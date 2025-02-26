package governance

import (
	"errors"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
	prbac "gno.land/p/gnoswap/rbac"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"
	gs "gno.land/r/gnoswap/v1/gov/staker"
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

// proposeRoleChange creates a new role change proposal with string parameters.
// The action and checkerType parameters should be case-insensitive strings matching their respective types.
// The addresses parameter should be a colon-separated string of addresses.
func ProposeRoleChange(
	title string,
	description string,
	roleName string,
	permissionName string,
	action string,
	checkerType string,
	addresses string,
	tokenPath string,
	minAmount uint64,
) uint64 {
	// Convert string parameters to their respective types
	roleChangeAction := RoleChangeActionFromString(action)
	checkerTypeEnum := checkerTypeFromString(checkerType)
	addressList := parseAddresses(addresses)

	return proposeRoleChange(
		title,
		description,
		roleName,
		permissionName,
		roleChangeAction,
		checkerTypeEnum,
		addressList,
		tokenPath,
		minAmount,
	)
}

// proposeRoleChange creates a new role change proposal
func proposeRoleChange(
	title string,
	description string,
	roleName string,
	permissionName string,
	action RoleChangeAction,
	checkerType CheckerType,
	addresses []std.Address,
	tokenPath string,
	minAmount uint64,
) uint64 {
	common.IsHalted()

	en.MintAndDistributeGns()
	updateProposalsState()

	proposer := std.PreviousRealm().Address()

	enough, balance, wanted := checkEnoughXGnsToPropose(proposer)
	if !enough {
		panic(addDetailToError(
			errNotEnoughBalance,
			ufmt.Sprintf("proposer(%s) has not enough xGNS, balance(%d), wanted(%d)", proposer.String(), balance, wanted),
		))
	}

	now := uint64(time.Now().Unix())
	votingMax, possibleAddressWithWeight := gs.GetPossibleVotingAddressWithWeight(now - config.VotingWeightSmoothingDuration)

	maxVotingWeight := u256.NewUint(votingMax)
	quorumAmount := maxVotingWeight.Uint64() * config.Quorum / 100

	proposal := ProposalInfo{
		Proposer:     proposer,
		ProposalType: RoleChange,
		State: ProposalState{
			Created:   true,
			CreatedAt: now,
			Upcoming:  true,
		},
		Yea:                       u256.Zero(),
		Nay:                       u256.Zero(),
		MaxVotingWeight:           maxVotingWeight,
		PossibleAddressWithWeight: possibleAddressWithWeight,
		ConfigVersion:             uint64(configVersions.Size()),
		QuorumAmount:              quorumAmount,
		Title:                     title,
		Description:               description,
		RoleChange: RoleChangeInfo{
			RoleName:       roleName,
			PermissionName: permissionName,
			Action:         action,
			CheckerType:    checkerType,
			Addresses:      addresses,
			TokenPath:      tokenPath,
			MinAmount:      minAmount,
		},
	}

	proposalId++
	proposals.Set(formatUint(proposalId), proposal)
	latestProposalByProposer.Set(proposer.String(), proposalId)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"ProposeRoleChange",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"title", title,
		"description", description,
		"roleName", roleName,
		"permissionName", permissionName,
		"action", action.String(),
		"checkerType", checkerType.String(),
		"proposalId", formatUint(proposalId),
		"quorumAmount", formatUint(proposal.QuorumAmount),
		"maxVotingWeight", proposal.MaxVotingWeight.ToString(),
		"configVersion", formatUint(proposal.ConfigVersion),
		"createdAt", formatUint(proposal.State.CreatedAt),
	)

	return proposalId
}

// parseAddresses parses a colon-separated string of addresses into an array of std.Address.
// Example input: "addr1:addr2:addr3"
func parseAddresses(addressStr string) []std.Address {
	if addressStr == "" {
		return []std.Address{}
	}

	addrStrings := strings.Split(addressStr, ":")
	addresses := make([]std.Address, 0, len(addrStrings))

	for _, addr := range addrStrings {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		address := std.Address(addr)
		addresses = append(addresses, address)
	}

	return addresses
}
