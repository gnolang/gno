package launchpad

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// Deposit represents a deposit made by a user in a launchpad project.
//
// This struct contains the necessary data and methods to manage and distribute
// rewards for a specific deposit.
//
// Fields:
// - createdBlockTimeInfo (BlockTimeInfo): The block time and height information for the creation of the deposit.
// - endBlockTimeInfo (BlockTimeInfo): The block time and height information for the end of the deposit.
// - depositor (std.Address): The address of the depositor.
// - id (string): The unique identifier for the deposit.
// - projectID (string): The ID of the project associated with the deposit.
// - tier (int64): The tier of the deposit.
// - depositAmount (int64): The amount of the deposit.
// - withdrawnHeight (int64): The height at which the deposit was withdrawn.
type Deposit struct {
	createdBlockTimeInfo *BlockTimeInfo
	endBlockTimeInfo     *BlockTimeInfo

	depositor std.Address

	id              string
	projectID       string
	tier            int64 // 30, 60, 180 // instead of tierId
	depositAmount   int64
	withdrawnHeight int64
}

func (d *Deposit) ID() string {
	return d.id
}

func (d *Deposit) ProjectID() string {
	return d.projectID
}

func (d *Deposit) ProjectTierID() string {
	return makeProjectTierID(d.projectID, d.tier)
}

func (d *Deposit) Tier() int64 {
	return d.tier
}

func (d *Deposit) Depositor() std.Address {
	return d.depositor
}

func (d *Deposit) DepositAmount() int64 {
	return d.depositAmount
}

func (d *Deposit) DepositHeight() int64 {
	return d.createdBlockTimeInfo.BlockHeight()
}

func (d *Deposit) DepositTime() int64 {
	return d.createdBlockTimeInfo.BlockTime()
}

func (d *Deposit) IsOwner(address std.Address) bool {
	return d.depositor.String() == address.String()
}

func (d *Deposit) EndTime() int64 {
	return d.endBlockTimeInfo.BlockTime()
}

func (d *Deposit) EndHeight() int64 {
	return d.endBlockTimeInfo.BlockHeight()
}

func (d *Deposit) IsEnded(currentHeight int64) bool {
	return d.endBlockTimeInfo.BlockHeight() < currentHeight
}

func (d *Deposit) IsWithdrawn() bool {
	return d.withdrawnHeight > 0
}

func (d *Deposit) withdraw(currentHeight int64) int64 {
	d.withdrawnHeight = currentHeight

	previousDepositAmount := d.depositAmount
	d.depositAmount = 0

	return previousDepositAmount
}

// NewDeposit returns a pointer to a new Deposit with the given values.
func NewDeposit(
	projectID string,
	tier int64,
	depositor std.Address,
	depositAmount int64,
	createdBlockTimeInfo *BlockTimeInfo,
	endBlockTimeInfo *BlockTimeInfo,
) *Deposit {
	return &Deposit{
		id:                   makeDepositID(projectID, tier, depositor, createdBlockTimeInfo.BlockHeight()),
		projectID:            projectID,
		tier:                 tier,
		depositor:            depositor,
		depositAmount:        depositAmount,
		withdrawnHeight:      0,
		createdBlockTimeInfo: createdBlockTimeInfo,
		endBlockTimeInfo:     endBlockTimeInfo,
	}
}

func makeDepositID(projectID string, tier int64, depositor std.Address, height int64) string {
	return ufmt.Sprintf("%s:%d:%s:%d", projectID, tier, depositor.String(), height)
}
