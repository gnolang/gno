package launchpad

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
)

// Project represents a launchpad project.
//
// This struct contains the necessary data and methods to manage and distribute
// rewards for a specific project.
//
// Fields:
// - id (string): The unique identifier for the project, formatted as "{tokenPath}:{createdHeight}".
// - name (string): The name of the project.
// - tokenPath (string): The path of the token associated with the project.
// - depositAmount (int64): The total amount of tokens deposited for the project.
// - recipient (std.Address): The address to receive the project's rewards.
// - conditions (map[string]*ProjectCondition): A map of token paths to their associated conditions.
// - tiers (map[int64]*ProjectTier): A map of tier durations to their associated tiers.
// - tiersRatios (map[int64]int64): A map of tier durations to their associated ratios.
// - createdBlockTimeInfo (BlockTimeInfo): The block time and height information for the creation of the project.
type Project struct {
	id                   string // 'tokenPath:createdHeight'
	name                 string
	tokenPath            string
	depositAmount        int64
	recipient            std.Address                  // string
	conditions           map[string]*ProjectCondition // tokenPath -> Condition
	tiers                map[int64]*ProjectTier
	tiersRatios          map[int64]int64
	createdBlockTimeInfo *BlockTimeInfo
}

func (p *Project) ID() string {
	return p.id
}

func (p *Project) Name() string {
	return p.name
}

func (p *Project) TokenPath() string {
	return p.tokenPath
}

func (p *Project) DepositAmount() int64 {
	return p.depositAmount
}

func (p *Project) Recipient() std.Address {
	return p.recipient
}

func (p *Project) TiersRatios() map[int64]int64 {
	return p.tiersRatios
}

func (p *Project) CreatedBlockTimeInfo() *BlockTimeInfo {
	return p.createdBlockTimeInfo
}

func (p *Project) CreatedAt() int64 {
	return p.createdBlockTimeInfo.BlockTime()
}

func (p *Project) CreatedHeight() int64 {
	return p.createdBlockTimeInfo.BlockHeight()
}

func (p *Project) StartTime() int64 {
	return p.standardTier().StartTime()
}

func (p *Project) StartHeight() int64 {
	return p.standardTier().StartHeight()
}

func (p *Project) EndTime() int64 {
	return p.standardTier().EndTime()
}

func (p *Project) EndHeight() int64 {
	return p.standardTier().EndHeight()
}

func (p *Project) IsActive(currentHeight int64) bool {
	return p.standardTier().isActivated(currentHeight)
}

func (p *Project) IsEnded(currentHeight int64) bool {
	return p.standardTier().isEnded(currentHeight)
}

func (p *Project) IsOwner(caller std.Address) bool {
	return p.recipient == caller
}

func (p *Project) RemainingAmount() int64 {
	remainingAmount := int64(0)

	for _, tier := range p.getTiers() {
		remainingAmount += tier.calculateLeftReward()
	}

	return remainingAmount
}

func (p *Project) CheckConditions(caller std.Address, balanceOfFunc func(tokenPath string, caller std.Address) int64) error {
	conditions := p.getConditions()
	if conditions == nil {
		return makeErrorWithDetails(errInvalidData, "conditions is nil")
	}

	for _, condition := range conditions {
		// xGNS(or GNS) may have a zero condition
		if !condition.IsAvailable() {
			continue
		}

		tokenPath := condition.TokenPath()
		balance := balanceOfFunc(tokenPath, caller)

		if err := condition.CheckCondition(tokenPath, balance); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) TotalDepositCount() int64 {
	totalRecipient := int64(0)

	for _, tier := range p.getTiers() {
		totalRecipient += tier.totalDepositCount
	}

	return totalRecipient
}

func (p *Project) TotalDepositAmount() int64 {
	totalDepositAmount := int64(0)

	for _, tier := range p.getTiers() {
		totalDepositAmount += tier.TotalDepositAmount()
	}

	return totalDepositAmount
}

func (p *Project) CurrentDepositCount() int64 {
	totalDepositCount := int64(0)

	for _, tier := range p.getTiers() {
		totalDepositCount += tier.CurrentDepositCount()
	}

	return totalDepositCount
}

func (p *Project) CurrentDepositAmount() int64 {
	totalDepositAmount := int64(0)

	for _, tier := range p.getTiers() {
		totalDepositAmount += tier.CurrentDepositAmount()
	}

	return totalDepositAmount
}

func (p *Project) TotalCollectedAmount() int64 {
	totalCollectedAmount := int64(0)

	for _, tier := range p.getTiers() {
		totalCollectedAmount += tier.TotalCollectedAmount()
	}

	return totalCollectedAmount
}

func (p *Project) getConditions() map[string]*ProjectCondition {
	return p.conditions
}

func (p *Project) getTiers() map[int64]*ProjectTier {
	return p.tiers
}

func (p *Project) getTier(duration int64) (*ProjectTier, error) {
	tier, exists := p.tiers[duration]
	if !exists {
		return nil, makeErrorWithDetails(errDataNotFound, ufmt.Sprintf("tier(%s) not found", duration))
	}

	return tier, nil
}

func (p *Project) standardTier() *ProjectTier {
	projectTier, exists := p.tiers[projectTier180]
	if !exists {
		return nil
	}

	return projectTier
}

func (p *Project) validateRefundRemainingAmount(currentHeight int64) error {
	if !p.IsEnded(currentHeight) {
		return errors.New(
			ufmt.Sprintf("project not ended yet(current:%d, endHeight: %d)", currentHeight, p.EndHeight()),
		)
	}

	if p.RemainingAmount() == 0 {
		return errors.New(
			ufmt.Sprintf("project has no remaining amount"),
		)
	}

	return nil
}

func (p *Project) addProjectTier(tierDuration int64, projectTier *ProjectTier) {
	p.tiers[tierDuration] = projectTier
}

func (p *Project) addProjectCondition(tokenPath string, condition *ProjectCondition) {
	p.conditions[tokenPath] = condition
}

func NewProject(
	name string,
	tokenPath string,
	depositAmount int64,
	recipient std.Address,
	createdBlockTimeInfo *BlockTimeInfo,
) *Project {

	return &Project{
		id:                   makeProjectID(tokenPath, createdBlockTimeInfo.BlockHeight()),
		name:                 name,
		tokenPath:            tokenPath,
		depositAmount:        depositAmount,
		recipient:            recipient,
		conditions:           make(map[string]*ProjectCondition),
		tiers:                make(map[int64]*ProjectTier),
		createdBlockTimeInfo: createdBlockTimeInfo,
	}
}

// makeProjectID generates a unique project ID based on the given token path and the current block height.
//
// The generated ID combines the `tokenPath` and the current block height in the following format:
// "{tokenPath}:{height}"
//
// Parameters:
// - tokenPath (string): The path of the token associated with the project.
//
// Returns:
// - string: A unique project ID in the format "tokenPath:height".
func makeProjectID(tokenPath string, createdHeight int64) string {
	return ufmt.Sprintf("%s:%d", tokenPath, createdHeight)
}
