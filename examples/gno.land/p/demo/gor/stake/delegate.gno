package stake

import (
	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/governance/checkpoints"
	"gno.land/p/demo/ufmt"
	std "std"
)

// any representation of voting power, like validator...
// there should be a process for any member to be a Delegate, like an applyMsg
type Delegate struct {
	kind                 string // TODO: self-delegate or delegate to others
	addr                 std.Address
	tokens               map[string]*checkpoints.History // denom => history(slice of snapshotted values)
	DelegtorShares       map[string]*checkpoints.History
	TallyDeductionShares uint64 // overrided voting powers will be deducted while tallying
	isBonded             bool   // status of delegate
	isJailed             bool   // jailed by misbehavior
	commission           int
	minSelfDelegation    uint64
}

func NewDelegate(minSelfDelegation uint64, denom string) *Delegate {
	return &Delegate{
		tokens:         map[string]*checkpoints.History{denom: checkpoints.NewHistory()},
		DelegtorShares: map[string]*checkpoints.History{denom: checkpoints.NewHistory()},
	}
}

func (d *Delegate) GetAddr() std.Address {
	return d.addr
}

func (d *Delegate) GetSnapshotTokens(snapshot int64, denom string) uint64 {
	history := d.tokens[denom]
	return history.GetAtBlock(snapshot - 1)
}

func (d *Delegate) AddSnapshotTokens(snapshot int64, tokens uint64, denom string) {
	history, ok := d.tokens[denom]
	if !ok {
		panic("should not happen")
	}
	history.PushWithOp(add, tokens)
}

func (d *Delegate) SubstractSnapshotTokens(snapshot int64, tokens uint64, denom string) {
	history, ok := d.tokens[denom]
	if !ok {
		panic("should not happen")
	}
	history.PushWithOp(subtract, tokens)
}

func (d *Delegate) GetSnapshotDelegatorShares(snapshot int64, denom string) uint64 {
	history := d.DelegtorShares[denom]
	return history.GetAtBlock(snapshot - 1) // only get from mined blocks
}

// operation updates history only valid after this block mined
func (d *Delegate) AddSnapshotShares(snapshot int64, shares uint64, denom string) {
	history, ok := d.DelegtorShares[denom]
	if !ok {
		panic("should not happen")
	}
	history.PushWithOp(add, shares)
}

func (d *Delegate) SubstractSnapshotShares(snapshot int64, shares uint64, denom string) {
	history, ok := d.DelegtorShares[denom]
	if !ok {
		panic("should not happen")
	}
	history.PushWithOp(subtract, shares)
}

func (d *Delegate) Shares2VotingPower(snapshot int64, shares uint64, denom string) uint64 {
	return shares * d.GetSnapshotTokens(snapshot, denom) / d.GetSnapshotDelegatorShares(snapshot, denom)
}

func (d *Delegate) VotingPower2Shares(snapshot int64, votingPower uint64, denom string) uint64 {
	return votingPower * d.GetSnapshotDelegatorShares(snapshot, denom) / d.GetSnapshotTokens(snapshot, denom)
}

// ------------------------------------------------------
type Delegation struct {
	delegator string
	delegate  string
	history   *checkpoints.History // a delegation may wait for a duration to be mature
	opts      *StakeOptions        // per delegation
}

func NewDelegation(opts *StakeOptions, delegator string, delegate string) *Delegation {
	d := &Delegation{}
	if opts == nil {
		d.opts = DefaultOptions()
	} else {
		d.opts = opts
	}
	d.delegator = delegator
	d.delegate = delegate
	d.history = checkpoints.NewHistory()
	return d
}

func (d *Delegation) GetDelegator() string {
	return d.delegator
}

func (d *Delegation) GetDelegate() string {
	return d.delegate
}

func (d *Delegation) GetSnapshotDelegationShares(blockNumber int64) uint64 {
	return d.history.GetAtBlock(blockNumber - 1)
}

func add(a uint64, b uint64) uint64 {
	return a + b
}

func subtract(a uint64, b uint64) uint64 {
	return a - b
}

func (d *Delegation) addDelegationShares(shares uint64) {
	d.history.PushWithOp(add, shares)
}

func (d *Delegation) substractDelegationShares(shares uint64) {
	d.history.PushWithOp(subtract, shares)
}

// append new entry
func (d *UnDelegation) appendUnbondShares(matureHeight int64, ubshares uint64) {
	ck := checkpoints.NewCheckPoint(matureHeight, ubshares)
	d.UDHistory.UpdateOrderedHistory(ck)
}

// delegation with a index of `delegateAddr`
type IndexedDelegation struct {
	indexedD map[string]*Delegation
}

func NewIndexedDelegation(index string, d *Delegation) *IndexedDelegation {
	return &IndexedDelegation{
		indexedD: map[string]*Delegation{index: d},
	}
}

func (inD *IndexedDelegation) IterateIndexedDelegation(cb func(index string, d *Delegation)) {
	for index, d := range inD.indexedD {
		cb(index, d)
	}
}

func (inD *IndexedDelegation) setDelegation(index string, d *Delegation) {
	inD.indexedD[index] = d
}

func (inD *IndexedDelegation) GetDelegationByIndex(index string) (*Delegation, error) {
	if d, ok := inD.indexedD[index]; ok {
		return d, nil
	}
	return nil, errors.New("no delegation exist for index%s: " + index)
}

// return an indexed delegation
func (s *Stake) GetDelegation(delegatorAddr string, delegateAddr string) (*IndexedDelegation, error) {
	p, found := s.delegations.Get(delegatorAddr)
	if !found {
		return nil, errors.New("delegation from not exist")
	}
	indexedDelegation := p.(*IndexedDelegation)

	_, err := indexedDelegation.GetDelegationByIndex(delegateAddr)
	if err != nil {
		return nil, err
	}

	return indexedDelegation, nil
}

func (s *Stake) GetDelegationByDelegator(delegatorAddr string) (*IndexedDelegation, error) {
	p, found := s.delegations.Get(delegatorAddr)
	if !found {
		return nil, errors.New("delegation from not exist")
	}
	indexedDelegation := p.(*IndexedDelegation)

	return indexedDelegation, nil
}

func (s *Stake) SetDelegation(delegator string, ind *IndexedDelegation) {
	s.delegations.Set(delegator, ind)
}

// ---------------------------------------

type UnDelegation struct {
	UDHistory *checkpoints.History
}

func NewUnDelegation() *UnDelegation {
	return &UnDelegation{
		UDHistory: new(checkpoints.History),
	}
}

type IndexedUnDelegation struct {
	inud map[string]*UnDelegation
}

func NewIndexedUnDelegation(index string, uds *UnDelegation) *IndexedUnDelegation {
	inD := &IndexedUnDelegation{
		inud: map[string]*UnDelegation{index: uds},
	}
	return inD
}

func (inuD *IndexedUnDelegation) setUnDelegation(index string, ud *UnDelegation) {
	inuD.inud[index] = ud
}

func (inuD *IndexedUnDelegation) getUnDelegation(index string) (*UnDelegation, error) {
	if d, ok := inuD.inud[index]; ok {
		return d, nil
	}
	return nil, errors.New("no Undelegation exist for index%s: " + index)
}

func (s *Stake) GetUnbondDelegation(delegatorAddr string, delegateAddr string) (*IndexedUnDelegation, error) {
	d, found := s.unbondDelegations.Get(delegatorAddr)
	if !found {
		return nil, errors.New("unbond delegation not exist")
	}
	inuD := d.(*IndexedUnDelegation)
	_, err := inuD.getUnDelegation(delegateAddr)
	if err != nil {
		return nil, err
	}
	return inuD, nil
}

func (s *Stake) SetUnDelegation(delegator string, inud *IndexedUnDelegation) {
	s.unbondDelegations.Set(delegator, inud)
}

// there things be done:
// 1. update delegation, delegator will receives shares that is calculated by amount * totalShares/bondedTokens
// 2. update delegate tokens, shares
// 3. take tokens delegated
// TODO: check if delegate coins exeedeed
func (s *Stake) Delegate(delegateAddr string, symbol string, amt uint64, isRedelegate bool) error {
	// TODO: should assert origin caller here?
	// std.AssertOriginCall()
	delegatorAddr := std.GetOrigCaller().String() // origin caller?
	var denom string
	var amount uint64
	var isNativeToken bool
	// if is delegating a native coin, the arg `token` is not used
	if s.GovToken.GetDenom() == "ugnot" { // the native coin
		isNativeToken = true
		send := std.GetOrigSend()
		denom = send[0].Denom // assume one denom per delegate
		amount = uint64(send.AmountOf(denom))
	} else { // if is delegating grcXX coin, get from args
		denom = symbol
		amount = amt
	}
	pkgAddr := std.GetOrigPkgAddr() // when testing, would be `main` pkg, or outer pkgs import this pkg
	snapshot := std.GetHeight()     // current block

	checkIsValidAddress(delegatorAddr)
	checkIsValidAddress(delegateAddr)

	if !isRedelegate {
		delegate, err := s.GetDelegateByAddr(delegateAddr)
		if err != nil {
			return err
		}

		var delegation *Delegation
		var indexedDelegation *IndexedDelegation
		// check if delegation exists
		indexedDelegation, err = s.GetDelegation(delegatorAddr, delegateAddr)
		if err == nil { // delegation exists
			// get existing delegation
			delegation, err = indexedDelegation.GetDelegationByIndex(delegateAddr)
			if err != nil {
				return err
			}
			// convert to shares
			shareIssue := delegate.VotingPower2Shares(snapshot, amount, denom)
			// delegation update shares
			delegation.addDelegationShares(shareIssue)
			indexedDelegation.setDelegation(delegateAddr, delegation)
			// add delegate token and shares
			delegate.AddSnapshotShares(snapshot, shareIssue, denom)
			delegate.AddSnapshotTokens(snapshot, amount, denom)
		} else { // no delegation exists
			// new delegation
			delegation = NewDelegation(nil, delegatorAddr, delegateAddr)
			// calculate shares to be issued
			shareIssue := delegate.VotingPower2Shares(snapshot, amount, denom)
			delegation.addDelegationShares(shareIssue)

			indexedDelegation, err = s.GetDelegationByDelegator(delegatorAddr)
			if err != nil { // the delegator did not delegate before
				indexedDelegation = NewIndexedDelegation(delegateAddr, delegation) // set delegation
			}

			indexedDelegation.setDelegation(delegateAddr, delegation) // insert one entry in the existing indexedDelegation map

			// update delegate tokens and shares
			delegate.AddSnapshotShares(snapshot, shareIssue, denom)
			delegate.AddSnapshotTokens(snapshot, amount, denom)

			// native coin is sent to pool by cli
			// grc20 coins sent by hand
			if !isNativeToken { // grcXX token
				s.GovToken.Transfer(std.Address(delegatorAddr), std.Address(pkgAddr), amount)
			}
			// TODO: bonded pool/unbonded pool?
		}
		// set delegation store
		s.SetDelegation(delegatorAddr, indexedDelegation)
	} else { // redelegate
		println("it'a a redelegate, to be implemented")
	}
	return nil
}

// deduct delegation shares, delegate tokens and shares
// token unbond will be returned to delegator later
// all undelegation tasks will be kept in a queue, maitained by Undelegation()
// return delegation, for the caller -> undelegation use
func (s *Stake) Unbond(delegatorAddr string, delegateAddr string, ubshares uint64) (d *Delegation, tokenUnbond uint64, err error) {
	// sanity check, if addr existed, amount not exeeded...
	checkIsValidAddress(delegatorAddr)
	checkIsValidAddress(delegateAddr)
	denom := s.GetDenom()
	snapshot := std.GetHeight()

	indexedDelegation, err := s.GetDelegation(delegatorAddr, delegateAddr)
	if err != nil {
		return nil, 0, err
	}

	delegation, err := indexedDelegation.GetDelegationByIndex(delegateAddr)
	if err != nil {
		return nil, 0, err
	}

	existShares := delegation.GetSnapshotDelegationShares(snapshot)
	if ubshares > existShares {
		return nil, 0, errors.New("unbonding shares exeeded maximum shares in delegation")
	}
	// update shares
	delegation.substractDelegationShares(ubshares)

	println("remained shares of delegation after substract is: ", delegation.GetSnapshotDelegationShares(snapshot))

	// set indexed delegation
	indexedDelegation.setDelegation(delegateAddr, delegation)
	// set store
	s.SetDelegation(delegatorAddr, indexedDelegation)

	// TODO: remove a delegation or set a flag when it goes to zero?

	// delegate sanity check
	delegate, err := s.GetDelegateByAddr(delegateAddr)
	if err != nil {
		return nil, 0, err
	}
	if ubshares > delegate.GetSnapshotDelegatorShares(snapshot, denom) {
		return nil, 0, errors.New("unbonding shares exeeded maximum shares in delegate")
	}
	// if delegate Unbond amounts exeeds minimum bond requirement, jail it
	if delegatorAddr == delegate.addr { // a delegate self unbound, like validator, with some constraints
		if delegate.GetSnapshotDelegatorShares(snapshot, denom)-ubshares < delegate.minSelfDelegation && !delegate.isJailed {
			delegate.isJailed = true
		}
	}

	// update delegate, if it's unbonded state, and shares zero, remove delegate
	tokenUB := delegate.Shares2VotingPower(snapshot, ubshares, denom) // calculate before deduct
	delegate.SubstractSnapshotShares(snapshot, ubshares, denom)
	delegate.SubstractSnapshotTokens(snapshot, tokenUB, denom)
	println("token unbond is: ", tokenUB)
	println("remained tokens of delegate is: ", delegate.GetSnapshotTokens(snapshot, denom))
	println("remained total delegator shares of delegate is: ", delegate.GetSnapshotDelegatorShares(snapshot, denom))
	// TODO: remove delegate or set jailed?, if it returns to zero?

	s.delegates.Set(delegateAddr, delegate)

	// NOTE: refund coins is handled by Undelegate, cuz it needs a unbonding period to finish
	return delegation, tokenUB, nil
}

func (s *Stake) Undelegate(delegatorAddr string, delegateAddr string, amount uint64) error {
	// unbond
	delegation, _, err := s.Unbond(delegatorAddr, delegateAddr, amount)
	if err != nil {
		return err
	}
	// put a task into queue, contains token unbond and other fields
	// new ubd item
	matureHeight := std.GetHeight() + delegation.opts.UnbondingLockDuration

	var indexedUnDelegation *IndexedUnDelegation
	var unDelegation *UnDelegation
	indexedUnDelegation, err = s.GetUnbondDelegation(delegatorAddr, delegateAddr)
	if err == nil {
		unDelegation, _ = indexedUnDelegation.getUnDelegation(delegateAddr)
		unDelegation.appendUnbondShares(matureHeight, amount)
	} else {
		unDelegation = NewUnDelegation()
		unDelegation.appendUnbondShares(matureHeight, amount)
		// new with given object
		indexedUnDelegation = NewIndexedUnDelegation(delegateAddr, unDelegation)
	}
	// set object
	indexedUnDelegation.setUnDelegation(delegateAddr, unDelegation)
	// set store
	s.SetUnDelegation(delegatorAddr, indexedUnDelegation)
	return nil
}

func (s *Stake) CompleteUnbondDelegation() error {
	snapshot := std.GetHeight()

	// TODO: defer to catch panic
	s.unbondDelegations.Iterate("", "", func(n *avl.Node) bool {
		indexedUndelegation := n.Value().(*IndexedUnDelegation)
		for delegateAddr, ud := range indexedUndelegation.inud {

			delegate, err := s.GetDelegateByAddr(delegateAddr)
			if err != nil {
				panic(err.Error())
			}
			// iterating different kinds of coins
			ud.UDHistory.IterateHistory(snapshot, func(ck checkpoints.Checkpoint) {
				unbondTokens := delegate.Shares2VotingPower(snapshot, ck.GetValue(), s.GetDenom())
				// send coins from pool to delegator
				s.Refund(unbondTokens)
			})
		}
		return false
	})
	return nil
}

func Redelegation(delegator, delegateAddr std.Address, amount uint64) error {
	// unbound, only moves between pools, no return to user, so undelegate is not needed

	// build new delegate pair
	return nil
}

func checkIsValidAddress(addr string) {
	if addr == "" {
		panic("invalid address")
	}
	return
}

func (s *Stake) RenderHome(account std.Address) string {
	totalSupply := s.getTotalSupply()
	str := ""
	str += ufmt.Sprintf("**Total supply**: %d\n", totalSupply)
	return str
}

func emit(event interface{}) {
	// TODO: should we do something there?
	// noop
}
