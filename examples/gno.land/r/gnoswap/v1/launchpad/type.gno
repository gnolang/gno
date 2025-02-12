package launchpad

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gov/xgns"
)

type ProjectStats struct {
	totalDeposit      uint64 // won't be decreased  (for statistics)
	actualDeposit     uint64 // will be decreased if deposit collected 'CollectDepositGns()'
	totalParticipant  uint64 // accu, won't be decreased (for statistics)
	actualParticipant uint64 // will be decreased if deposit collected 'CollectDepositGns()'
	totalCollected    uint64 // collect reward amount
}

// NewProjectStats returns a pointer to a new ProjectStats with the given values.
func NewProjectStats(
	totalDeposit, actualDeposit, totalParticipant, actualParticipant, totalCollected uint64,
) *ProjectStats {
	return &ProjectStats{
		totalDeposit:      totalDeposit,
		actualDeposit:     actualDeposit,
		totalParticipant:  totalParticipant,
		actualParticipant: actualParticipant,
		totalCollected:    totalCollected,
	}
}

func (ps *ProjectStats) TotalDeposit() uint64 {
	return ps.totalDeposit
}
func (ps *ProjectStats) SetTotalDeposit(v uint64) {
	ps.totalDeposit = v
}
func (ps *ProjectStats) ActualDeposit() uint64 {
	return ps.actualDeposit
}
func (ps *ProjectStats) SetActualDeposit(v uint64) {
	ps.actualDeposit = v
}
func (ps *ProjectStats) TotalParticipant() uint64 {
	return ps.totalParticipant
}
func (ps *ProjectStats) SetTotalParticipant(v uint64) {
	ps.totalParticipant = v
}
func (ps *ProjectStats) ActualParticipant() uint64 {
	return ps.actualParticipant
}
func (ps *ProjectStats) SetActualParticipant(v uint64) {
	ps.actualParticipant = v
}
func (ps *ProjectStats) TotalCollected() uint64 {
	return ps.totalCollected
}
func (ps *ProjectStats) SetTotalCollected(v uint64) {
	ps.totalCollected = v
}

type Project struct {
	id            string // 'tokenPath:createdHeight'
	name          string
	tokenPath     string
	depositAmount uint64
	recipient     std.Address          // string
	conditions    map[string]Condition // tokenPath -> Condition
	tiers         map[uint64]Tier
	tiersRatios   map[uint64]uint64
	created       TimeInfo
	started       TimeInfo
	ended         TimeInfo // same with tier 180's data
	stats         ProjectStats
	refund        RefundInfo
}

// NewProject returns a pointer to a new Project with the given values.
func NewProject(
	id, name, tokenPath string,
	depositAmount uint64,
	recipient std.Address,
	conditions map[string]Condition,
	tiers map[uint64]Tier,
	tiersRatios map[uint64]uint64,
	created, started, ended TimeInfo,
	stats ProjectStats,
	refund RefundInfo,
) *Project {
	return &Project{
		id:            id,
		name:          name,
		tokenPath:     tokenPath,
		depositAmount: depositAmount,
		recipient:     recipient,
		conditions:    conditions,
		tiers:         tiers,
		tiersRatios:   tiersRatios,
		created:       created,
		started:       started,
		ended:         ended,
		stats:         stats,
		refund:        refund,
	}
}

func (p *Project) ID() string {
	return p.id
}
func (p *Project) SetID(v string) {
	p.id = v
}
func (p *Project) Name() string {
	return p.name
}
func (p *Project) SetName(v string) {
	p.name = v
}
func (p *Project) TokenPath() string {
	return p.tokenPath
}
func (p *Project) SetTokenPath(v string) {
	p.tokenPath = v
}
func (p *Project) DepositAmount() uint64 {
	return p.depositAmount
}
func (p *Project) SetDepositAmount(v uint64) {
	p.depositAmount = v
}
func (p *Project) Recipient() std.Address {
	return p.recipient
}
func (p *Project) SetRecipient(v std.Address) {
	p.recipient = v
}
func (p *Project) Conditions() map[string]Condition {
	return p.conditions
}
func (p *Project) SetConditions(v map[string]Condition) {
	p.conditions = v
}
func (p *Project) Tiers() map[uint64]Tier {
	return p.tiers
}
func (p *Project) SetTiers(v map[uint64]Tier) {
	p.tiers = v
}
func (p *Project) TiersRatios() map[uint64]uint64 {
	return p.tiersRatios
}
func (p *Project) SetTiersRatios(v map[uint64]uint64) {
	p.tiersRatios = v
}
func (p *Project) Created() TimeInfo {
	return p.created
}
func (p *Project) SetCreated(v TimeInfo) {
	p.created = v
}
func (p *Project) Started() TimeInfo {
	return p.started
}
func (p *Project) SetStarted(v TimeInfo) {
	p.started = v
}
func (p *Project) Ended() TimeInfo {
	return p.ended
}
func (p *Project) SetEnded(v TimeInfo) {
	p.ended = v
}
func (p *Project) Stats() *ProjectStats {
	return &p.stats
}
func (p *Project) SetStats(v ProjectStats) {
	p.stats = v
}
func (p *Project) Refund() RefundInfo {
	return p.refund
}
func (p *Project) SetRefund(v RefundInfo) {
	p.refund = v
}
func (p *Project) Tier(key uint64) (Tier, error) {
	tier, exists := p.tiers[key]
	if !exists {
		return Tier{}, errors.New(addDetailToError(
			errDataNotFound, ufmt.Sprintf("tier(%s) not found", key)))
	}
	return tier, nil
}
func (p *Project) isActivated(currentHeight uint64) bool {
	return p.started.Height() <= currentHeight && currentHeight < p.ended.Height()
}
func (p *Project) checkDepositConditions(caller std.Address) {
	conditions := p.Conditions()
	if conditions == nil {
		panic(addDetailToError(
			errInvalidData, "conditions is nil"))
	}

	for _, condition := range conditions {
		// xGNS(or GNS) may have a zero condition
		if condition.MinAmount() == 0 {
			continue
		}

		var balance uint64
		if condition.TokenPath() == consts.GOV_XGNS_PATH {
			balance = xgns.BalanceOf(caller)
		} else {
			balance = common.BalanceOf(condition.TokenPath(), caller)
		}

		if balance < condition.MinAmount() {
			panic(addDetailToError(
				errNotEnoughBalance, ufmt.Sprintf("insufficient balance(%d) for condition token(%s)", balance, condition.tokenPath)))
		}
	}
}
func (p *Project) setTier(key uint64, tier Tier) {
	p.tiers[key] = tier
}
func (p *Project) calculateLeftReward() uint64 {
	return p.DepositAmount() - p.Stats().totalCollected
}

type Tier struct {
	id                     string // '{projectId}:duration' // duartion == 30, 90, 180
	collectWaitDuration    uint64 // block
	tierAmount             uint64
	tierAmountPerBlockX128 *u256.Uint
	started                TimeInfo // first deposit height
	ended                  TimeInfo //
	totalDepositAmount     uint64   // accumulated deposit amount
	actualDepositAmount    uint64   // actual deposit amount
	totalParticipant       uint64   // accumulated participant
	actualParticipant      uint64   // actual participant
	userCollectedAmount    uint64   // total collected amount by user
	calculatedAmount       uint64   // total calculated amount
	reward                 Reward
}

// NewTier returns a pointer to a new Tier with the given values.
func NewTier(
	id string,
	collectWaitDuration uint64,
	tierAmount uint64,
	tierAmountPerBlockX128 *u256.Uint,
	started, ended TimeInfo,
	totalDepositAmount, actualDepositAmount uint64,
	totalParticipant, actualParticipant uint64,
	userCollectedAmount, calculatedAmount uint64,
	reward Reward,
) *Tier {
	return &Tier{
		id:                     id,
		collectWaitDuration:    collectWaitDuration,
		tierAmount:             tierAmount,
		tierAmountPerBlockX128: tierAmountPerBlockX128,
		started:                started,
		ended:                  ended,
		totalDepositAmount:     totalDepositAmount,
		actualDepositAmount:    actualDepositAmount,
		totalParticipant:       totalParticipant,
		actualParticipant:      actualParticipant,
		userCollectedAmount:    userCollectedAmount,
		calculatedAmount:       calculatedAmount,
		reward:                 reward,
	}
}

func (t *Tier) ID() string {
	return t.id
}
func (t *Tier) SetID(v string) {
	t.id = v
}
func (t *Tier) CollectWaitDuration() uint64 {
	return t.collectWaitDuration
}
func (t *Tier) SetCollectWaitDuration(v uint64) {
	t.collectWaitDuration = v
}
func (t *Tier) TierAmount() uint64 {
	return t.tierAmount
}
func (t *Tier) SetTierAmount(v uint64) {
	t.tierAmount = v
}
func (t *Tier) TierAmountPerBlockX128() *u256.Uint {
	return t.tierAmountPerBlockX128
}
func (t *Tier) SetTierAmountPerBlockX128(v *u256.Uint) {
	t.tierAmountPerBlockX128 = v.Clone()
}
func (t *Tier) Started() TimeInfo {
	return t.started
}
func (t *Tier) SetStarted(v TimeInfo) {
	t.started = v
}
func (t *Tier) Ended() TimeInfo {
	return t.ended
}
func (t *Tier) SetEnded(v TimeInfo) {
	t.ended = v
}
func (t *Tier) TotalDepositAmount() uint64 {
	return t.totalDepositAmount
}
func (t *Tier) SetTotalDepositAmount(v uint64) {
	t.totalDepositAmount = v
}
func (t *Tier) ActualDepositAmount() uint64 {
	return t.actualDepositAmount
}
func (t *Tier) SetActualDepositAmount(v uint64) {
	t.actualDepositAmount = v
}
func (t *Tier) TotalParticipant() uint64 {
	return t.totalParticipant
}
func (t *Tier) SetTotalParticipant(v uint64) {
	t.totalParticipant = v
}
func (t *Tier) ActualParticipant() uint64 {
	return t.actualParticipant
}
func (t *Tier) SetActualParticipant(v uint64) {
	t.actualParticipant = v
}
func (t *Tier) UserCollectedAmount() uint64 {
	return t.userCollectedAmount
}
func (t *Tier) SetUserCollectedAmount(v uint64) {
	t.userCollectedAmount = v
}
func (t *Tier) CalculatedAmount() uint64 {
	return t.calculatedAmount
}
func (t *Tier) SetCalculatedAmount(v uint64) {
	t.calculatedAmount = v
}
func (t *Tier) Reward() Reward {
	return t.reward
}
func (t *Tier) SetReward(v Reward) {
	t.reward = v
}
func (t *Tier) isActivated(currentHeight uint64) bool {
	return t.started.Height() <= currentHeight && currentHeight < t.ended.Height()
}
func (t *Tier) isFirstDeposit() bool {
	return t.totalParticipant == 0
}
func (t *Tier) updateStarted(height, time uint64) {
	t.started.SetHeight(height)
	t.started.SetTime(time)
	t.reward.SetLastHeight(height)
}
func (t *Tier) rewardPerBlockUint64() uint64 {
	return u256.Zero().Rsh(t.tierAmountPerBlockX128, 128).Uint64()
}
func (t *Tier) isEnded(currentHeight uint64) bool {
	return t.ended.Height() < currentHeight
}
func (t *Tier) calculateLeftReward() uint64 {
	return t.tierAmount - t.userCollectedAmount
}

type RewardInfo struct {
	priceDebt     *u256.Uint // price debt per GNS stake, Q128
	depositAmount uint64     // amount of GNS staked
	rewardAmount  uint64     // calculated, not collected
	claimed       uint64     // amount of reward claimed so far
	startHeight   uint64     // height when launchpad started staking
	EndHeight     uint64     // end height of reward calculation
	LastHeight    uint64     // last height when reward was calculated
}

// NewRewardInfo returns a pointer to a new RewardInfo with the given values.
func NewRewardInfo(
	priceDebt *u256.Uint,
	depositAmount, rewardAmount, claimed, startHeight, endHeight, lastHeight uint64,
) *RewardInfo {
	return &RewardInfo{
		priceDebt:     priceDebt,
		depositAmount: depositAmount,
		rewardAmount:  rewardAmount,
		claimed:       claimed,
		startHeight:   startHeight,
		EndHeight:     endHeight,
		LastHeight:    lastHeight,
	}
}

func (r *RewardInfo) PriceDebt() *u256.Uint {
	return r.priceDebt
}
func (r *RewardInfo) SetPriceDebt(v *u256.Uint) {
	r.priceDebt = v
}
func (r *RewardInfo) DepositAmount() uint64 {
	return r.depositAmount
}
func (r *RewardInfo) SetDepositAmount(v uint64) {
	r.depositAmount = v
}
func (r *RewardInfo) RewardAmount() uint64 {
	return r.rewardAmount
}
func (r *RewardInfo) SetRewardAmount(v uint64) {
	r.rewardAmount = v
}
func (r *RewardInfo) Claimed() uint64 {
	return r.claimed
}
func (r *RewardInfo) SetClaimed(v uint64) {
	r.claimed = v
}
func (r *RewardInfo) StartHeight() uint64 {
	return r.startHeight
}
func (r *RewardInfo) SetStartHeight(v uint64) {
	r.startHeight = v
}
func (r *RewardInfo) GetEndHeight() uint64 {
	return r.EndHeight
}
func (r *RewardInfo) SetEndHeight(v uint64) {
	r.EndHeight = v
}
func (r *RewardInfo) GetLastHeight() uint64 {
	return r.LastHeight
}
func (r *RewardInfo) SetLastHeight(v uint64) {
	r.LastHeight = v
}
func (r *RewardInfo) calculateReward(accumRewardPerDeposit *u256.Uint) uint64 {
	actualRewardPerDeposit := u256.Zero().Sub(accumRewardPerDeposit, r.PriceDebt())
	reward := u256.Zero().Mul(actualRewardPerDeposit, u256.NewUint(r.DepositAmount()))
	reward = u256.Zero().Rsh(reward, 128)
	return reward.Uint64() - r.Claimed()
}

type Reward struct {
	accumRewardPerDeposit *u256.Uint // claimable Launchpad reward per GNS stake, Q128
	LastHeight            uint64     // last height when reward was calculated
	EndHeight             uint64     // end height of reward calculation
	info                  *avl.Tree  // depositId -> RewardInfo
}

// NewReward returns a pointer to a new Reward with the given values.
func NewReward(
	accumRewardPerDeposit *u256.Uint,
	lastHeight, endHeight uint64,
) *Reward {
	return &Reward{
		accumRewardPerDeposit: accumRewardPerDeposit,
		LastHeight:            lastHeight,
		EndHeight:             endHeight,
		info:                  avl.NewTree(),
	}
}

func (r *Reward) AccumRewardPerDeposit() *u256.Uint {
	return r.accumRewardPerDeposit
}
func (r *Reward) SetAccumRewardPerDeposit(v *u256.Uint) {
	r.accumRewardPerDeposit = v
}
func (r *Reward) GetLastHeight() uint64 {
	return r.LastHeight
}
func (r *Reward) SetLastHeight(v uint64) {
	r.LastHeight = v
}
func (r *Reward) GetEndHeight() uint64 {
	return r.EndHeight
}
func (r *Reward) SetEndHeight(v uint64) {
	r.EndHeight = v
}
func (r *Reward) Info() *avl.Tree {
	return r.info
}
func (r *Reward) SetInfo(v *avl.Tree) {
	r.info = v
}
func (r *Reward) accumulationRewardUint64() uint64 {
	return u256.Zero().Rsh(r.accumRewardPerDeposit, 128).Uint64()
}
func (r *Reward) InfoOf(depositId string) *RewardInfo {
	infoI, exists := r.info.Get(depositId)
	if !exists {
		panic(addDetailToError(
			errNotExistDeposit, ufmt.Sprintf("(%s)", depositId)))
	}

	return infoI.(*RewardInfo)
}
func (r *Reward) calculateRewardPerDeposit(rewardPerBlock *u256.Uint, totalStaked uint64) (*u256.Uint, error) {
	// blockDuration * rewardPerBlock / totalStaked
	currentHeight := uint64(std.GetHeight())
	lastUpdateHeight := r.GetLastHeight()
	if currentHeight == lastUpdateHeight {
		return u256.Zero(), nil
	}
	if currentHeight < lastUpdateHeight {
		return nil, errors.New(addDetailToError(
			errInvalidHeight, ufmt.Sprintf("currentHeight(%d) <= lastUpdateHeight(%d)", currentHeight, lastUpdateHeight)))
	}
	if rewardPerBlock.IsZero() {
		return nil, errors.New(addDetailToError(
			errNoLeftReward, ufmt.Sprintf("rewardPerBlock(%d)", rewardPerBlock)))
	}
	if totalStaked == 0 {
		return u256.Zero(), nil
	}
	if currentHeight > r.EndHeight {
		currentHeight = r.EndHeight
	}

	blockDuration := currentHeight - lastUpdateHeight
	totalReward := u256.Zero().Mul(u256.NewUint(blockDuration), rewardPerBlock)
	rewardPerDeposit := u256.Zero().Div(totalReward, u256.NewUint(totalStaked))
	return rewardPerDeposit, nil
}

func (r *Reward) addRewardPerDeposit(rewardPerDeposit *u256.Uint) {
	if rewardPerDeposit.IsZero() {
		return
	}
	r.SetAccumRewardPerDeposit(u256.Zero().Add(r.AccumRewardPerDeposit(), rewardPerDeposit))
}

func (r *Reward) finalize(currentHeight, totalStaked uint64, rewardPerBlock *u256.Uint) (*u256.Uint, error) {
	if currentHeight < r.LastHeight {
		// Not started yet
		return nil, errors.New(addDetailToError(
			errInvalidHeight, ufmt.Sprintf("currentHeight(%d) <= LastHeight(%d)", currentHeight, r.LastHeight)))
	}
	if r.LastHeight > r.EndHeight {
		// already ended
		return nil, errors.New(addDetailToError(
			errInvalidHeight, ufmt.Sprintf("LastHeight(%d) >= EndHeight(%d)", r.LastHeight, r.EndHeight)))
	}
	if currentHeight > r.EndHeight {
		currentHeight = r.EndHeight
	}
	return r.calculateRewardPerDeposit(rewardPerBlock, totalStaked)
}
func (r *Reward) deductReward(depositId string, currentHeight uint64) uint64 {
	if currentHeight < r.LastHeight {
		panic(addDetailToError(
			errInvalidRewardState,
			ufmt.Sprintf("currentHeight %d is less than LastHeight %d", currentHeight, r.LastHeight)))
	}

	deposit := deposits[depositId]
	if currentHeight < deposit.claimableHeight {
		panic(addDetailToError(
			errInvalidRewardState,
			ufmt.Sprintf("currentHeight %d is less than claimableHeight %d", currentHeight, deposit.claimableHeight)))
	}

	info := r.InfoOf(depositId)
	if currentHeight < info.StartHeight() {
		panic(addDetailToError(
			errInvalidRewardState,
			ufmt.Sprintf("currentHeight %d is less than StartHeight %d", currentHeight, info.StartHeight())))
	}
	reward64 := info.calculateReward(r.AccumRewardPerDeposit())

	if reward64 == 0 {
		return 0
	}
	info.SetClaimed(info.Claimed() + reward64)
	r.info.Set(depositId, info)

	return reward64
}

type Deposit struct {
	id                   string // 'projectId:tier:depositor:height'
	projectId            string
	tier                 string // 30, 60, 180 // instead of tierId
	depositor            std.Address
	amount               uint64
	depositHeight        uint64
	depositTime          uint64
	depositCollectHeight uint64 // withdraw deposited gns height
	depositCollectTime   uint64 // withdraw deposited gns timestamp
	claimableHeight      uint64 // claimable reward block height
	claimableTime        uint64 // claimable reward timestamp
	rewardAmount         uint64 // calculated, not collected
	rewardCollected      uint64 // accu, collected
	rewardCollectHeight  uint64 // last collected height
	rewardCollectTime    uint64 // last collected time
}

// NewDeposit returns a pointer to a new Deposit with the given values.
func NewDeposit(
	id, projectId, tier string,
	depositor std.Address,
	amount, depositHeight, depositTime uint64,
	depositCollectHeight, depositCollectTime uint64,
	claimableHeight, claimableTime uint64,
	rewardAmount, rewardCollected uint64,
	rewardCollectHeight, rewardCollectTime uint64,
) *Deposit {
	return &Deposit{
		id:                   id,
		projectId:            projectId,
		tier:                 tier,
		depositor:            depositor,
		amount:               amount,
		depositHeight:        depositHeight,
		depositTime:          depositTime,
		depositCollectHeight: depositCollectHeight,
		depositCollectTime:   depositCollectTime,
		claimableHeight:      claimableHeight,
		claimableTime:        claimableTime,
		rewardAmount:         rewardAmount,
		rewardCollected:      rewardCollected,
		rewardCollectHeight:  rewardCollectHeight,
		rewardCollectTime:    rewardCollectTime,
	}
}

func (d *Deposit) ID() string {
	return d.id
}
func (d *Deposit) SetID(v string) {
	d.id = v
}
func (d *Deposit) ProjectID() string {
	return d.projectId
}
func (d *Deposit) SetProjectID(v string) {
	d.projectId = v
}
func (d *Deposit) Tier() string {
	return d.tier
}
func (d *Deposit) SetTier(v string) {
	d.tier = v
}
func (d *Deposit) Depositor() std.Address {
	return d.depositor
}
func (d *Deposit) SetDepositor(v std.Address) {
	d.depositor = v
}
func (d *Deposit) Amount() uint64 {
	return d.amount
}
func (d *Deposit) SetAmount(v uint64) {
	d.amount = v
}
func (d *Deposit) DepositHeight() uint64 {
	return d.depositHeight
}
func (d *Deposit) SetDepositHeight(v uint64) {
	d.depositHeight = v
}
func (d *Deposit) DepositTime() uint64 {
	return d.depositTime
}
func (d *Deposit) SetDepositTime(v uint64) {
	d.depositTime = v
}
func (d *Deposit) DepositCollectHeight() uint64 {
	return d.depositCollectHeight
}
func (d *Deposit) SetDepositCollectHeight(v uint64) {
	d.depositCollectHeight = v
}
func (d *Deposit) DepositCollectTime() uint64 {
	return d.depositCollectTime
}
func (d *Deposit) SetDepositCollectTime(v uint64) {
	d.depositCollectTime = v
}
func (d *Deposit) ClaimableHeight() uint64 {
	return d.claimableHeight
}
func (d *Deposit) SetClaimableHeight(v uint64) {
	d.claimableHeight = v
}
func (d *Deposit) ClaimableTime() uint64 {
	return d.claimableTime
}
func (d *Deposit) SetClaimableTime(v uint64) {
	d.claimableTime = v
}
func (d *Deposit) RewardAmount() uint64 {
	return d.rewardAmount
}
func (d *Deposit) SetRewardAmount(v uint64) {
	d.rewardAmount = v
}
func (d *Deposit) RewardCollected() uint64 {
	return d.rewardCollected
}
func (d *Deposit) SetRewardCollected(v uint64) {
	d.rewardCollected = v
}
func (d *Deposit) RewardCollectHeight() uint64 {
	return d.rewardCollectHeight
}
func (d *Deposit) SetRewardCollectHeight(v uint64) {
	d.rewardCollectHeight = v
}
func (d *Deposit) RewardCollectTime() uint64 {
	return d.rewardCollectTime
}
func (d *Deposit) SetRewardCollectTime(v uint64) {
	d.rewardCollectTime = v
}
func (d *Deposit) isClaimable(currentHeight uint64) bool {
	return d.claimableHeight <= currentHeight
}

type TimeInfo struct {
	height uint64
	time   uint64
}

// NewTimeInfo returns a pointer to a new TimeInfo with the given values.
func NewTimeInfo(height, t uint64) TimeInfo {
	return TimeInfo{
		height: height,
		time:   t,
	}
}

func (ti *TimeInfo) Height() uint64 {
	return ti.height
}
func (ti *TimeInfo) SetHeight(v uint64) {
	ti.height = v
}
func (ti *TimeInfo) Time() uint64 {
	return ti.time
}
func (ti *TimeInfo) SetTime(v uint64) {
	ti.time = v
}

type RefundInfo struct {
	amount uint64
	height uint64
	time   uint64
}

// NewRefundInfo returns a pointer to a new RefundInfo with the given values.
func NewRefundInfo(amount, height, t uint64) RefundInfo {
	return RefundInfo{
		amount: amount,
		height: height,
		time:   t,
	}
}

func (ri *RefundInfo) Amount() uint64 {
	return ri.amount
}
func (ri *RefundInfo) SetAmount(v uint64) {
	ri.amount = v
}
func (ri *RefundInfo) Height() uint64 {
	return ri.height
}
func (ri *RefundInfo) SetHeight(v uint64) {
	ri.height = v
}
func (ri *RefundInfo) Time() uint64 {
	return ri.time
}
func (ri *RefundInfo) SetTime(v uint64) {
	ri.time = v
}

type Condition struct {
	tokenPath string
	minAmount uint64
}

// NewCondition returns a pointer to a new Condition with the given values.
func NewCondition(tokenPath string, minAmount uint64) Condition {
	return Condition{
		tokenPath: tokenPath,
		minAmount: minAmount,
	}
}

func (c *Condition) TokenPath() string {
	return c.tokenPath
}
func (c *Condition) SetTokenPath(v string) {
	c.tokenPath = v
}
func (c *Condition) MinAmount() uint64 {
	return c.minAmount
}
func (c *Condition) SetMinAmount(v uint64) {
	c.minAmount = v
}
